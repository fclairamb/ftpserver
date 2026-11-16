// Package gdrive provides a Google Drive access layer
package gdrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	drv "github.com/fclairamb/afero-gdrive"
	drvoa "github.com/fclairamb/afero-gdrive/oauthhelper"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/utils"
)

// ErrMissingGoogleClientCredentials is returned when you have specified the google_client_id and/or
// google_client_secret
var ErrMissingGoogleClientCredentials = errors.New("missing the google client credentials")

// ErrAuthorizationPending is returned when the authorization of the same token file is already running for another
// login.
var ErrAuthorizationPending = errors.New("google drive authorization pending")

// ErrAuthorizationTimeout is returned when the first authorization wasn't completed in time.
var ErrAuthorizationTimeout = errors.New("google drive authorization not completed in time")

// ErrInvalidAuthorizationTimeout is returned when the authorization_timeout param isn't a positive Go duration.
var ErrInvalidAuthorizationTimeout = errors.New("invalid authorization_timeout")

// defaultAuthorizationTimeout is how long a login waits for the first authorization to be completed.
const defaultAuthorizationTimeout = 5 * time.Minute

var (
	// pendingAuthorizationsMu guards pendingAuthorizations.
	pendingAuthorizationsMu sync.Mutex
	// pendingAuthorizations holds the token files whose authorization is running.
	pendingAuthorizations = map[string]struct{}{}
)

// clientContext is the context of the Drive HTTP client. Tests override it to inject a fake transport.
var clientContext = context.Background

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access, logger *slog.Logger) (afero.Fs, error) {
	googleClientID := access.Params["google_client_id"]
	googleClientSecret := access.Params["google_client_secret"]
	tokenFile := access.Params["token_file"]
	basePath := access.Params["base_path"]

	if googleClientID == "" {
		googleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	}

	if googleClientSecret == "" {
		googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	if googleClientID == "" || googleClientSecret == "" {
		return nil, ErrMissingGoogleClientCredentials
	}

	if tokenFile == "" {
		tokenFile = fmt.Sprintf("gdrive_token_%s.json", access.User)
	}

	var token *oauth2.Token
	var err error

	saveToken := false

	if token, err = drvoa.LoadTokenFromFile(tokenFile); err != nil {
		logger.Warn(
			"Couldn't retrieve a token, we will need to generate one",
			"tokenFile", tokenFile,
			"userName", access.User,
			"err", err,
		)

		// A file that can't be decoded still returns a (partial) token
		token = nil
		saveToken = true
	} else if !token.Valid() {
		saveToken = true
	}

	authorizationTimeout := defaultAuthorizationTimeout

	if value := access.Params["authorization_timeout"]; value != "" {
		if authorizationTimeout, err = time.ParseDuration(value); err != nil || authorizationTimeout <= 0 {
			return nil, fmt.Errorf("%w: %q", ErrInvalidAuthorizationTimeout, value)
		}
	}

	if token == nil {
		token, err = authorize(googleClientID, googleClientSecret, tokenFile, authorizationTimeout, access, logger)
		if err != nil {
			return nil, err
		}
	}

	// The token is set, so no flow can start here. The client keeps a context that is never canceled: oauth2 uses
	// it for the later token refreshes.
	auth := drvoa.Auth{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Token:        token,
	}

	httpClient, err := auth.NewHTTPClient(clientContext())
	if err != nil {
		return nil, err
	}

	if saveToken {
		if errStoreToken := storeToken(tokenFile, auth.Token); errStoreToken != nil {
			return nil, fmt.Errorf("token couldn't be saved: %w", errStoreToken)
		}
	}

	gdriveFs, err := drv.New(httpClient)
	if err != nil {
		return nil, err
	}

	// Allowing to set the basePath in the driver
	if basePath != "" {
		basePath = utils.ReplaceEnvVars(basePath)

		if _, errSetRoot := gdriveFs.SetRootDirectory(basePath); errSetRoot != nil {
			return nil, fmt.Errorf("couldn't set the base path: %w", errSetRoot)
		}
	}

	return gdriveFs, err
}

// authorize runs the loopback authorization flow for tokenFile, bounded by timeout. Only one flow runs per token
// file: a concurrent login on the same file fails at once with ErrAuthorizationPending.
func authorize(
	clientID, clientSecret, tokenFile string,
	timeout time.Duration,
	access *confpar.Access,
	logger *slog.Logger,
) (*oauth2.Token, error) {
	key := tokenFile
	if abs, errAbs := filepath.Abs(tokenFile); errAbs == nil {
		key = abs
	}

	pendingAuthorizationsMu.Lock()

	if _, running := pendingAuthorizations[key]; running {
		pendingAuthorizationsMu.Unlock()

		return nil, fmt.Errorf("%w for token file %s", ErrAuthorizationPending, tokenFile)
	}

	pendingAuthorizations[key] = struct{}{}
	pendingAuthorizationsMu.Unlock()

	defer func() {
		pendingAuthorizationsMu.Lock()
		delete(pendingAuthorizations, key)
		pendingAuthorizationsMu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	auth := drvoa.Auth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// No Authenticate callback: afero-gdrive then runs the loopback flow, Google blocked the
		// out-of-band one (https://github.com/fclairamb/ftpserver/issues/1123).
		OpenURL: func(url string) error {
			logger.Warn(
				"Google Drive authorization needed: open the URL in a browser. If the browser can't reach "+
					"the server's 127.0.0.1 redirect, run curl on the server with the URL it redirected to",
				"url", url,
				"userName", access.User,
				"timeout", timeout,
			)

			return nil
		},
	}

	// Only used to get the token: the returned client would keep ctx, which is canceled on return.
	if _, err := auth.NewHTTPClient(ctx); err != nil {
		if ctx.Err() != nil {
			logger.Error(
				"Google Drive authorization not completed in time",
				"tokenFile", tokenFile,
				"userName", access.User,
				"timeout", timeout,
			)

			return nil, fmt.Errorf("%w (%s), token file %s: %w", ErrAuthorizationTimeout, timeout, tokenFile, err)
		}

		logger.Error(
			"Google Drive authorization failed",
			"tokenFile", tokenFile,
			"userName", access.User,
			"err", err,
		)

		return nil, fmt.Errorf("google drive authorization failed: %w", err)
	}

	return auth.Token, nil
}

// storeToken saves the token readable by its owner only, even if the file already exists.
func storeToken(path string, token *oauth2.Token) error {
	data, err := json.Marshal(token) //nolint:gosec // written to a 0600 file
	if err != nil {
		return fmt.Errorf("couldn't marshal token: %w", err)
	}

	if err = os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("couldn't write token file: %w", err)
	}

	if err = os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("couldn't set token file permissions: %w", err)
	}

	return nil
}

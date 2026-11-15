// Package gdrive provides a Google Drive access layer
package gdrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

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

		saveToken = true
	} else if !token.Valid() {
		saveToken = true
	}

	auth := drvoa.Auth{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Token:        token,
		// No Authenticate callback: afero-gdrive then runs the loopback flow, Google blocked the
		// out-of-band one (https://github.com/fclairamb/ftpserver/issues/1123).
		OpenURL: func(url string) error {
			logger.Warn(
				"Google Drive authorization needed: open the URL in a browser. If the browser can't reach "+
					"the server's 127.0.0.1 redirect, run curl on the server with the URL it redirected to",
				"url", url,
				"userName", access.User,
			)

			return nil
		},
	}

	httpClient, err := auth.NewHTTPClient(context.Background())
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

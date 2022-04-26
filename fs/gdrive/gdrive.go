// Package gdrive provides a Google Drive access layer
package gdrive

import (
	"context"
	"errors"
	"fmt"
	"os"

	drv "github.com/fclairamb/afero-gdrive"
	drvoa "github.com/fclairamb/afero-gdrive/oauthhelper"
	log "github.com/fclairamb/go-log"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ErrMissingGoogleClientCredentials is returned when you have specified the google_client_id and/or
// google_client_secret
var ErrMissingGoogleClientCredentials = errors.New("missing the google client credentials")

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access, logger log.Logger) (afero.Fs, error) {
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
		Authenticate: func(url string) (string, error) {
			fmt.Printf("Please go to %s and enter the received code:\n", url)
			var code string
			_, errScan := fmt.Scan(&code)

			return code, errScan
		},
	}

	httpClient, err := auth.NewHTTPClient(context.Background())
	if err != nil {
		return nil, err
	}

	if saveToken {
		if errStoreToken := drvoa.StoreTokenToFile(tokenFile, auth.Token); errStoreToken != nil {
			return nil, fmt.Errorf("token couldn't be saved: %w", errStoreToken)
		}
	}

	gdriveFs, err := drv.New(httpClient)
	if err != nil {
		return nil, err
	}

	gdriveFs.Logger = logger

	// Allowing to set the basePath in the driver
	if basePath != "" {
		if _, errSetRoot := gdriveFs.SetRootDirectory(basePath); errSetRoot != nil {
			return nil, fmt.Errorf("couldn't set the base path: %w", errSetRoot)
		}
	}

	return gdriveFs, err
}

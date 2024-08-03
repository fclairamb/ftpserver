package keycloak

import (
	"context"
	"errors"
	"os"

	"github.com/Nerzal/gocloak/v13"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/utils"
	"github.com/spf13/afero"
)

type KeycloakAuthenticator struct {
	client   gocloak.GoCloak
	clientID string
	secret   string
	realm    string
}

// CheckPasswd
func (a *KeycloakAuthenticator) CheckPasswd(user, pass string) (bool, error) {
	ctx := context.Background()
	_, err := a.client.Login(ctx, a.clientID, a.secret, a.realm, user, pass)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ErrMissingBasePath is triggered when the basePath property isn't specified
var ErrMissingBasePath = errors.New("basePath must be specified")
var ErrLogin = errors.New("login failed, User or Password error")
var ErrMissingKeycloakClientCredentials = errors.New("missing the keycloak client credentials")

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {

	keycloakClientID := access.Params["keycloak_client_id"]
	keycloakClientSecret := access.Params["keycloak_client_secret"]
	keycloakURL := access.Params["keycloak_url"]
	keycloakRealm := access.Params["keycloak_realm"]

	if keycloakClientID == "" {
		keycloakClientID = os.Getenv("KEYCLOAK_CLIENT_ID")
	}

	if keycloakClientSecret == "" {
		keycloakClientSecret = os.Getenv("KEYCLAOK_CLIENT_SECRET")
	}

	if keycloakClientID == "" || keycloakClientSecret == "" || keycloakURL == "" || keycloakRealm == "" {
		return nil, ErrMissingKeycloakClientCredentials
	}

	basePath := access.Params["base_path"]
	if basePath == "" {
		return nil, ErrMissingBasePath
	}

	// Keycloak Authenticator
	auth := &KeycloakAuthenticator{
		client:   *gocloak.NewClient(keycloakURL),
		clientID: keycloakClientID,
		secret:   keycloakClientSecret,
		realm:    keycloakRealm,
	}

	valid, _ := auth.CheckPasswd(access.User, access.Pass)
	if !valid {
		return nil, ErrLogin
	}

	basePath = utils.ReplaceEnvVars(basePath)

	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}

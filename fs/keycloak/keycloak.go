// Package keycloak provides a local file system access layer whose users are authenticated by Keycloak
package keycloak

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs/utils"
)

// Authenticator checks users credentials against a Keycloak server
type Authenticator struct {
	client   *gocloak.GoCloak
	clientID string
	secret   string
	realm    string
}

// KeycloakAuthenticator is the former name of Authenticator
//
// Deprecated: use Authenticator instead.
type KeycloakAuthenticator = Authenticator //nolint:revive // kept for backward compatibility

// CheckPasswd checks the user's credentials against Keycloak
func (a *Authenticator) CheckPasswd(user, pass string) (bool, error) {
	if _, err := a.client.Login(context.Background(), a.clientID, a.secret, a.realm, user, pass); err != nil {
		return false, err
	}

	return true, nil
}

var (
	// ErrMissingBasePath is triggered when the base_path property isn't specified
	ErrMissingBasePath = errors.New("base_path must be specified")

	// ErrLogin is returned when the user couldn't be authenticated by Keycloak
	ErrLogin = errors.New("login failed, User or Password error")

	// ErrMissingKeycloakClientCredentials is returned when the keycloak parameters are incomplete
	ErrMissingKeycloakClientCredentials = errors.New("missing the keycloak client credentials")

	// ErrInvalidUserName is returned when the user name can't be used as a directory name
	ErrInvalidUserName = errors.New("invalid user name")
)

// getParam returns a parameter, or the first non-empty of the fallback environment variables
func getParam(access *confpar.Access, name string, envVars ...string) string {
	if value := access.Params[name]; value != "" {
		return value
	}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			return value
		}
	}

	return ""
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	keycloakClientID := getParam(access, "keycloak_client_id", "KEYCLOAK_CLIENT_ID")
	// KEYCLAOK_CLIENT_SECRET is a typo that was supported by previous versions, we keep it for backward compatibility
	keycloakClientSecret := getParam(access, "keycloak_client_secret", "KEYCLOAK_CLIENT_SECRET", "KEYCLAOK_CLIENT_SECRET")
	keycloakURL := access.Params["keycloak_url"]
	keycloakRealm := access.Params["keycloak_realm"]

	if keycloakClientID == "" || keycloakClientSecret == "" || keycloakURL == "" || keycloakRealm == "" {
		return nil, ErrMissingKeycloakClientCredentials
	}

	basePath := access.Params["base_path"]
	if basePath == "" {
		return nil, ErrMissingBasePath
	}

	// The user name is used as a directory name, it must not allow to escape the base path
	if access.User == "" || access.User == "." || access.User == ".." || strings.ContainsAny(access.User, `/\`) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUserName, access.User)
	}

	auth := &Authenticator{
		client:   gocloak.NewClient(keycloakURL),
		clientID: keycloakClientID,
		secret:   keycloakClientSecret,
		realm:    keycloakRealm,
	}

	if valid, _ := auth.CheckPasswd(access.User, access.Pass); !valid {
		return nil, ErrLogin
	}

	basePath = path.Join(utils.ReplaceEnvVars(basePath), access.User)

	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}

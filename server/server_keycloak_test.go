package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// newFakeKeycloak starts a fake Keycloak token endpoint that only accepts the given credentials
func newFakeKeycloak(t *testing.T, user, pass string) *httptest.Server {
	t.Helper()

	keycloak := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token") || r.ParseForm() != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if r.PostForm.Get("username") != user || r.PostForm.Get("password") != pass {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))

			return
		}

		_, _ = w.Write([]byte(`{"access_token":"token","token_type":"Bearer","expires_in":300}`))
	}))
	t.Cleanup(keycloak.Close)

	return keycloak
}

func newKeycloakTestServer(t *testing.T, keycloakURL string) *Server {
	t.Helper()

	return newTestServer(t, &confpar.Content{
		Accesses: []*confpar.Access{{
			Fs:     "keycloak",
			Shared: true,
			Params: map[string]string{
				"keycloak_client_id":     "ftp",
				"keycloak_client_secret": "secret",
				"keycloak_url":           keycloakURL,
				"keycloak_realm":         "acme",
				"base_path":              t.TempDir(),
			},
		}},
	})
}

// TestKeycloakSharedChecksPassword makes sure a shared keycloak access still checks the password of each login
func TestKeycloakSharedChecksPassword(t *testing.T) {
	srv := newKeycloakTestServer(t, newFakeKeycloak(t, "alice", "good").URL)

	if _, err := srv.AuthUser(nil, "alice", "good"); err != nil {
		t.Fatalf("valid login failed: %v", err)
	}

	if _, err := srv.AuthUser(nil, "alice", "wrong"); err == nil {
		t.Fatal("login with a wrong password succeeded after a valid login")
	}

	if _, err := srv.AuthUser(nil, "alice", "good"); err != nil {
		t.Fatalf("second valid login failed: %v", err)
	}
}

// TestKeycloakWrongPassword makes sure a wrong password is refused when nobody logged in before
func TestKeycloakWrongPassword(t *testing.T) {
	srv := newKeycloakTestServer(t, newFakeKeycloak(t, "alice", "good").URL)

	if _, err := srv.AuthUser(nil, "alice", "wrong"); err == nil {
		t.Fatal("login with a wrong password succeeded")
	}
}

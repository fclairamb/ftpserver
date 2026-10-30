package config

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/fclairamb/ftpserver/config/confpar"
)

func newTestConfig(t *testing.T, accesses ...*confpar.Access) *Config {
	t.Helper()

	conf, err := FromContent(&confpar.Content{Accesses: accesses}, "test.json", slog.Default())
	if err != nil {
		t.Fatalf("FromContent: %v", err)
	}

	return conf
}

func TestGetAccess(t *testing.T) {
	conf := newTestConfig(t,
		&confpar.Access{User: "plain", Pass: "plain", Fs: "os"},
		&confpar.Access{User: "bcrypt", Pass: "$2b$05$aco32n/z66W2Va5bqZHXqOZVYZ0SWW3CvGVg8802jSJWRH15SYWUC", Fs: "os"},
		&confpar.Access{
			User: "sha512crypt",
			Pass: "$6$NM1oYAm3MGJY/TZa$srQwlvW1g7isQFL5.ECpxYo6sK4DNQXwlEmmIdQVbAGS2LJ6zb8ACrpMRGzWx4i9r1p886sT7EY" +
				"Ncva7Ypsc/1",
			Fs: "os",
		},
		&confpar.Access{User: "anonymous", Pass: "*", Fs: "os"},
	)

	for _, tc := range []struct {
		user, pass string
		ok         bool
	}{
		{"plain", "plain", true},
		{"plain", "wrong", false},
		{"bcrypt", "bcrypt", true},
		{"bcrypt", "wrong", false},
		{"sha512crypt", "sha512crypt", true},
		{"sha512crypt", "wrong", false},
		{"anonymous", "anything", true},
		{"unknown", "plain", false},
	} {
		access, err := conf.GetAccess(tc.user, tc.pass)

		if tc.ok {
			if err != nil || access == nil || access.User != tc.user {
				t.Errorf("%s/%s: expected access, got %v / %v", tc.user, tc.pass, access, err)
			}
		} else if !errors.Is(err, ErrUnknownUser) {
			t.Errorf("%s/%s: expected ErrUnknownUser, got %v / %v", tc.user, tc.pass, access, err)
		}
	}
}

// TestGetAccessKeycloakDoesNotShareState makes sure concurrent keycloak logins don't overwrite each other's
// credentials, which could otherwise make a user end up in another user's directory.
func TestGetAccessKeycloakDoesNotShareState(t *testing.T) {
	shared := &confpar.Access{Fs: "keycloak", Params: map[string]string{"base_path": "/ftp"}}
	conf := newTestConfig(t, shared)

	var wg sync.WaitGroup

	for _, user := range []string{"alice", "bob", "carol", "dave"} {
		wg.Add(1)

		go func() {
			defer wg.Done()

			access, err := conf.GetAccess(user, user+"-pass")
			if err != nil {
				t.Errorf("GetAccess: %v", err)

				return
			}

			if access.User != user || access.Pass != user+"-pass" {
				t.Errorf("got access for %s/%s instead of %s", access.User, access.Pass, user)
			}
		}()
	}

	wg.Wait()

	if shared.User != "" || shared.Pass != "" {
		t.Errorf("the shared access was modified: %s/%s", shared.User, shared.Pass)
	}
}

func TestHashPlaintextPasswords(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "ftpserver.json")
	content := `{
  "hash_plaintext_passwords": true,
  "accesses": [
    {"user": "plain", "pass": "secret", "fs": "os", "params": {"basePath": "/tmp"}},
    {"user": "anonymous", "pass": "*", "fs": "os", "params": {"basePath": "/tmp"}},
    {"user": "bcrypt", "pass": "$2b$05$aco32n/z66W2Va5bqZHXqOZVYZ0SWW3CvGVg8802jSJWRH15SYWUC", "fs": "os"}
  ]
}`

	if err := os.WriteFile(fileName, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	conf, err := NewConfig(fileName, slog.Default())
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	if !strings.HasPrefix(conf.Content.Accesses[0].Pass, "$2") {
		t.Errorf("plain-text password wasn't hashed: %s", conf.Content.Accesses[0].Pass)
	}

	if conf.Content.Accesses[1].Pass != "*" {
		t.Errorf("anonymous password shouldn't be hashed: %s", conf.Content.Accesses[1].Pass)
	}

	if _, err := conf.GetAccess("plain", "secret"); err != nil {
		t.Errorf("couldn't authenticate with the hashed password: %v", err)
	}

	// The saved file must be hashed too, and keep its restrictive permissions
	saved, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if strings.Contains(string(saved), `"secret"`) {
		t.Error("the saved config file still contains the plain-text password")
	}

	info, err := os.Stat(fileName)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file permissions changed to %o", perm)
	}
}

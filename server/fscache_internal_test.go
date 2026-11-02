package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	serverlib "github.com/fclairamb/ftpserverlib"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/config/confpar"
)

// fakeClientContext only implements what AuthUser uses when no logging is enabled.
type fakeClientContext struct {
	serverlib.ClientContext
}

func authFs(t *testing.T, srv *Server, user, pass string) afero.Fs {
	t.Helper()

	driver, err := srv.AuthUser(fakeClientContext{}, user, pass)
	if err != nil {
		t.Fatalf("AuthUser(%s): %v", user, err)
	}

	clientDriver, ok := driver.(*ClientDriver)
	if !ok {
		t.Fatalf("unexpected driver type %T", driver)
	}

	return clientDriver.Fs
}

func listNames(t *testing.T, fsys afero.Fs) map[string]bool {
	t.Helper()

	infos, err := afero.ReadDir(fsys, "/")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	names := make(map[string]bool, len(infos))
	for _, info := range infos {
		names[info.Name()] = true
	}

	return names
}

// TestSharedFsSameUserOtherAccess makes sure a shared access of a user isn't handed to another access of the same
// user name (here a read-only access on another directory).
func TestSharedFsSameUserOtherAccess(t *testing.T) {
	rwDir := t.TempDir()
	roDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(roDir, "ro-only.txt"), []byte("ro"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	srv := newTestServer(t, &confpar.Content{
		Accesses: []*confpar.Access{
			{User: "bob", Pass: "rw", Fs: "os", Shared: true, Params: map[string]string{"basePath": rwDir}},
			{User: "bob", Pass: "ro", Fs: "os", ReadOnly: true, Params: map[string]string{"basePath": roDir}},
		},
	})

	rwFs := authFs(t, srv, "bob", "rw")
	if err := afero.WriteFile(rwFs, "/rw.txt", []byte("rw"), 0o600); err != nil {
		t.Fatalf("writing on the read-write access: %v", err)
	}

	roFs := authFs(t, srv, "bob", "ro")
	if roFs == rwFs {
		t.Fatal("the read-only access got the shared fs of the read-write access")
	}

	if err := afero.WriteFile(roFs, "/new.txt", []byte("x"), 0o600); err == nil {
		t.Fatal("writing on the read-only access should fail")
	}

	names := listNames(t, roFs)
	if !names["ro-only.txt"] || names["rw.txt"] {
		t.Fatalf("the read-only access lists %v, expected the content of its own directory", names)
	}
}

// TestSharedFsWebhookWithoutUser makes sure a webhook reply without a user field never shares an instance between
// two logins.
func TestSharedFsWebhookWithoutUser(t *testing.T) {
	root := t.TempDir()

	for _, user := range []string{"alice", "bob"} {
		if err := os.Mkdir(filepath.Join(root, user), 0o700); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
	}

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			User string `json:"user"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fs":     "os",
			"shared": true,
			"params": map[string]string{"basePath": filepath.Join(root, creds.User)},
		})
	}))
	defer webhook.Close()

	srv := newTestServer(t, &confpar.Content{
		AccessesWebhook: &confpar.AccessesWebhook{URL: webhook.URL},
	})

	aliceFs := authFs(t, srv, "alice", "x")
	if err := afero.WriteFile(aliceFs, "/alice.txt", []byte("alice"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	bobFs := authFs(t, srv, "bob", "x")
	if bobFs == aliceFs {
		t.Fatal("bob got alice's shared fs")
	}

	if names := listNames(t, bobFs); names["alice.txt"] {
		t.Fatalf("bob sees alice's files: %v", names)
	}

	if err := afero.WriteFile(bobFs, "/bob.txt", []byte("bob"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "bob", "bob.txt")); err != nil {
		t.Fatalf("bob's file isn't in bob's directory: %v", err)
	}
}

// TestSharedFsReuseAndReload makes sure two logins on the same shared access share their instance, and that a
// configuration reload drops it.
func TestSharedFsReuseAndReload(t *testing.T) {
	dir := t.TempDir()
	content := &confpar.Content{
		Accesses: []*confpar.Access{
			{User: "carol", Pass: "pass", Fs: "os", Shared: true, Params: map[string]string{"basePath": dir}},
			{User: "dave", Pass: "pass", Fs: "os", Params: map[string]string{"basePath": dir}},
		},
	}

	confFile := filepath.Join(t.TempDir(), "ftpserver.json")

	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if err := os.WriteFile(confFile, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	conf, err := config.FromContent(content, confFile, slog.Default())
	if err != nil {
		t.Fatalf("FromContent: %v", err)
	}

	srv, err := NewServer(conf, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	first := authFs(t, srv, "carol", "pass")
	if second := authFs(t, srv, "carol", "pass"); second != first {
		t.Fatal("two logins on the same shared access should share their fs instance")
	}

	if authFs(t, srv, "dave", "pass") == authFs(t, srv, "dave", "pass") {
		t.Fatal("a non-shared access should get a new fs instance at each login")
	}

	if err := srv.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig: %v", err)
	}

	if afterReload := authFs(t, srv, "carol", "pass"); afterReload == first {
		t.Fatal("the shared fs instance should be dropped by a reload")
	}
}

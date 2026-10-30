package main

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fclairamb/ftpserver/config"
)

func newTestConfFile(t *testing.T) (string, string) {
	t.Helper()

	confFile := filepath.Join(t.TempDir(), "ftpserver.json")
	pass := createConfFile(confFile, slog.Default())

	if pass == "" {
		t.Fatal("No password returned")
	}

	return confFile, pass
}

func loadConf(t *testing.T, confFile string) *config.Config {
	t.Helper()

	conf, err := config.NewConfig(confFile, slog.Default())
	if err != nil {
		t.Fatalf("Can't load the created conf: %v", err)
	}

	return conf
}

func TestCreateConfFileNoDefaultPassword(t *testing.T) {
	confFile, _ := newTestConfFile(t)
	conf := loadConf(t, confFile)

	if _, err := conf.GetAccess("test", "test"); !errors.Is(err, config.ErrUnknownUser) {
		t.Fatalf("The test/test credentials must be rejected, got err=%v", err)
	}
}

func TestCreateConfFileGeneratedPassword(t *testing.T) {
	confFile, pass := newTestConfFile(t)
	conf := loadConf(t, confFile)

	if len(pass) < 16 {
		t.Fatalf("Password too short: %d characters", len(pass))
	}

	access, err := conf.GetAccess("test", pass)
	if err != nil {
		t.Fatalf("The generated password must be accepted: %v", err)
	}

	if access.Fs != "os" || access.Params["basePath"] != "/tmp" {
		t.Fatalf("Unexpected access: %+v", access)
	}

	content, err := os.ReadFile(confFile)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), pass) {
		t.Fatal("The password must not be stored in clear")
	}

	info, err := os.Stat(confFile)
	if err != nil {
		t.Fatal(err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("Unexpected file mode: %o", perm)
	}

	_, otherPass := newTestConfFile(t)
	if otherPass == pass {
		t.Fatal("Two config files must get different passwords")
	}
}

func TestCreateConfFileKeepsExistingFile(t *testing.T) {
	confFile := filepath.Join(t.TempDir(), "ftpserver.json")
	original := []byte(`{"version": 1, "accesses": []}`)

	if err := os.WriteFile(confFile, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if pass := createConfFile(confFile, slog.Default()); pass != "" {
		t.Fatal("No password should be generated for an existing file")
	}

	content, err := os.ReadFile(confFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != string(original) {
		t.Fatalf("The existing file was modified: %s", content)
	}
}

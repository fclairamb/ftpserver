package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/config/confpar"
)

// countingCloser counts the Close calls of the file systems created by a test server
type countingCloser struct {
	closes atomic.Int32
}

func (c *countingCloser) Close() error {
	c.closes.Add(1)

	return nil
}

// newClosingServer creates a server whose file systems are in-memory and share one closer that counts Close calls.
// Its configuration is written to a file, so that ReloadConfig works.
func newClosingServer(t *testing.T, accesses ...*confpar.Access) (*Server, *countingCloser) {
	t.Helper()

	content := &confpar.Content{Accesses: accesses}

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

	closer := &countingCloser{}
	srv.fsLoader = func(_ *confpar.Access, _ *slog.Logger) (afero.Fs, io.Closer, error) {
		return afero.NewMemMapFs(), closer, nil
	}

	return srv, closer
}

// connect opens a fake client session on the server
func connect(t *testing.T, srv *Server) fakeClientContext {
	t.Helper()

	cc := newFakeClient()
	if _, err := srv.ClientConnected(cc); err != nil {
		t.Fatalf("ClientConnected: %v", err)
	}

	return cc
}

func assertCloses(t *testing.T, closer *countingCloser, expected int32, step string) {
	t.Helper()

	if got := closer.closes.Load(); got != expected {
		t.Fatalf("%s: expected %d Close calls, got %d", step, expected, got)
	}
}

// TestNonSharedFsClosedOnDisconnect makes sure the file system of a non-shared access is closed when its session ends
func TestNonSharedFsClosedOnDisconnect(t *testing.T) {
	srv, closer := newClosingServer(t, &confpar.Access{User: "alice", Pass: "pass", Fs: "sftp"})

	cc := connect(t, srv)
	authClientFs(t, srv, cc, "alice", "pass")
	assertCloses(t, closer, 0, "after login")

	srv.ClientDisconnected(cc)
	assertCloses(t, closer, 1, "after disconnect")

	// Releasing the session again must not close it twice
	srv.releaseSession(cc.ID())
	assertCloses(t, closer, 1, "after a second release")
}

// TestNonSharedFsClosedOnRelogin makes sure a client logging in again in the same session releases its previous
// file system
func TestNonSharedFsClosedOnRelogin(t *testing.T) {
	srv, closer := newClosingServer(t, &confpar.Access{User: "alice", Pass: "pass", Fs: "sftp"})

	cc := connect(t, srv)
	authClientFs(t, srv, cc, "alice", "pass")
	authClientFs(t, srv, cc, "alice", "pass")
	assertCloses(t, closer, 1, "after the second login")

	srv.ClientDisconnected(cc)
	assertCloses(t, closer, 2, "after disconnect")
}

// TestUnauthenticatedDisconnect makes sure a client that never logged in closes nothing
func TestUnauthenticatedDisconnect(t *testing.T) {
	srv, closer := newClosingServer(t, &confpar.Access{User: "alice", Pass: "pass", Fs: "sftp"})

	cc := connect(t, srv)
	if _, err := srv.AuthUser(cc, "alice", "wrong"); err == nil {
		t.Fatal("expected a wrong password to be refused")
	}

	srv.ClientDisconnected(cc)
	assertCloses(t, closer, 0, "after disconnect")
}

// TestSharedFsClosedAfterReloadAndLastSession makes sure a shared file system stays open while it is cached, and is
// closed once a reload dropped it and its last session ended
func TestSharedFsClosedAfterReloadAndLastSession(t *testing.T) {
	srv, closer := newClosingServer(t, &confpar.Access{User: "carol", Pass: "pass", Fs: "sftp", Shared: true})

	first, second := connect(t, srv), connect(t, srv)
	authClientFs(t, srv, first, "carol", "pass")
	authClientFs(t, srv, second, "carol", "pass")

	srv.ClientDisconnected(first)
	srv.ClientDisconnected(second)
	assertCloses(t, closer, 0, "after both sessions ended, the instance is still cached")

	// Two new sessions use the cached instance, then a reload drops it
	third, fourth := connect(t, srv), connect(t, srv)
	authClientFs(t, srv, third, "carol", "pass")
	authClientFs(t, srv, fourth, "carol", "pass")

	if err := srv.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig: %v", err)
	}

	assertCloses(t, closer, 0, "after reload, while sessions still use it")

	srv.ClientDisconnected(third)
	assertCloses(t, closer, 0, "after the first session ended")

	srv.ClientDisconnected(fourth)
	assertCloses(t, closer, 1, "after the last session ended")
}

// TestSharedFsUnusedClosedOnReload makes sure a reload closes a dropped shared file system no session uses
func TestSharedFsUnusedClosedOnReload(t *testing.T) {
	srv, closer := newClosingServer(t, &confpar.Access{User: "carol", Pass: "pass", Fs: "sftp", Shared: true})

	cc := connect(t, srv)
	authClientFs(t, srv, cc, "carol", "pass")
	srv.ClientDisconnected(cc)

	if err := srv.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig: %v", err)
	}

	assertCloses(t, closer, 1, "after reload")

	// A new session gets a new instance, closed by the next reload once its session ended
	other := connect(t, srv)
	authClientFs(t, srv, other, "carol", "pass")
	srv.ClientDisconnected(other)

	if err := srv.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig: %v", err)
	}

	assertCloses(t, closer, 2, "after the second reload")
}

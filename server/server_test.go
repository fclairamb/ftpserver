package server_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/server"
)

// TestClientDriverSymlink checks that the local OS backend can create symlinks
// through the ClientDriver (issue #980).
func TestClientDriverSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "target.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("couldn't write target file: %v", err)
	}

	driver := &server.ClientDriver{Fs: afero.NewBasePathFs(afero.NewOsFs(), tmpDir)}

	if err := driver.Symlink("/target.txt", "/link.txt"); err != nil {
		t.Fatalf("Symlink failed: %v", err)
	}

	info, err := os.Lstat(filepath.Join(tmpDir, "link.txt"))
	if err != nil {
		t.Fatalf("couldn't lstat the symlink: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link.txt is not a symlink (mode %v)", info.Mode())
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "link.txt"))
	if err != nil {
		t.Fatalf("couldn't read through the symlink: %v", err)
	}

	if string(content) != "hello" {
		t.Fatalf("unexpected content through symlink: %q", content)
	}
}

// TestClientDriverSymlinkUnsupported checks that a backend without symlink
// support returns afero.ErrNoSymlink instead of panicking (issue #980).
func TestClientDriverSymlinkUnsupported(t *testing.T) {
	driver := &server.ClientDriver{Fs: afero.NewMemMapFs()}

	err := driver.Symlink("/target.txt", "/link.txt")
	if err == nil {
		t.Fatal("expected an error for a backend without symlink support")
	}

	if !errors.Is(err, afero.ErrNoSymlink) {
		t.Fatalf("expected afero.ErrNoSymlink, got: %v", err)
	}
}

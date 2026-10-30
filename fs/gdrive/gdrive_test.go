package gdrive

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	drvoa "github.com/fclairamb/afero-gdrive/oauthhelper"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestStoreTokenPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions only")
	}

	token := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}

	t.Run("new file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token.json")
		require.NoError(t, storeToken(path, token))

		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		loaded, err := drvoa.LoadTokenFromFile(path)
		require.NoError(t, err)
		require.Equal(t, "access", loaded.AccessToken)
		require.Equal(t, "refresh", loaded.RefreshToken)
	})

	t.Run("existing wide file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token.json")
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
		require.NoError(t, os.Chmod(path, 0o644))
		require.NoError(t, storeToken(path, token))

		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
}

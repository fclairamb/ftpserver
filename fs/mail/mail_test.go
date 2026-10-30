package mail

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fclairamb/ftpserver/config/confpar"
)

func newFs(t *testing.T, params map[string]string) *Fs {
	t.Helper()

	fs, err := LoadFs(&confpar.Access{Params: params})
	require.NoError(t, err)

	mfs, ok := fs.(*Fs)
	require.True(t, ok)

	return mfs
}

func TestWriteTooLarge(t *testing.T) {
	fs := newFs(t, map[string]string{"MaxSize": "1024"})
	f := &File{Path: "a.bin", Fs: fs}

	_, err := f.Write(make([]byte, 2048))
	require.ErrorIs(t, err, ErrFileTooLarge)
	require.LessOrEqual(t, len(f.Content), 1024)
}

func TestWriteWithinLimit(t *testing.T) {
	fs := newFs(t, map[string]string{"MaxSize": "1024"})
	f := &File{Path: "a.bin", Fs: fs}

	for range 4 {
		n, err := f.Write(make([]byte, 128))
		require.NoError(t, err)
		require.Equal(t, 128, n)
	}

	require.Len(t, f.Content, 512)
}

func TestMaxSizeParameter(t *testing.T) {
	for _, v := range []string{"abc", "0", "-1"} {
		_, err := LoadFs(&confpar.Access{Params: map[string]string{"MaxSize": v}})
		require.ErrorIs(t, err, ErrInvalidParameter, v)
	}

	require.Equal(t, int64(defaultMaxSize), newFs(t, map[string]string{}).MaxSize)
}

func TestCloseAfterRefusedWrite(t *testing.T) {
	// The dialer points to a closed port: if Close tried to send, it would return a different error.
	fs := newFs(t, map[string]string{"MaxSize": "10", "Host": "127.0.0.1", "Port": "1"})
	f := &File{Path: "a.bin", Fs: fs}

	_, err := f.Write(make([]byte, 20))
	require.Error(t, err)
	require.ErrorIs(t, f.Close(), ErrFileTooLarge)
}

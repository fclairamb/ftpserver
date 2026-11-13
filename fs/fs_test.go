package fs

import (
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/fclairamb/ftpserver/config/confpar"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// waitGoroutines waits for the number of goroutines to go back to at most max, and returns the last count
func waitGoroutines(maxGoroutines int) int {
	deadline := time.Now().Add(5 * time.Second)

	for {
		nb := runtime.NumGoroutine()
		if nb <= maxGoroutines || time.Now().After(deadline) {
			return nb
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func loadFsLoop(t *testing.T, access *confpar.Access) (afero.Fs, io.Closer) {
	t.Helper()

	var lastFs afero.Fs

	var lastCloser io.Closer

	for range 50 {
		fs, closer, err := LoadFs(access, discardLogger())
		require.NoError(t, err)

		if closer != nil {
			require.NoError(t, closer.Close())
		}

		lastFs, lastCloser = fs, closer
	}

	return lastFs, lastCloser
}

func TestLoadFsSyncAndDeleteReleasesWrapper(t *testing.T) {
	access := &confpar.Access{
		Fs:     "os",
		Params: map[string]string{"basePath": t.TempDir()},
		SyncAndDelete: &confpar.SyncAndDelete{
			Enable:    true,
			Directory: t.TempDir(),
		},
	}

	start := runtime.NumGoroutine()

	_, closer := loadFsLoop(t, access)
	require.NotNil(t, closer, "the sync & delete wrapper must be closable")

	require.LessOrEqual(t, waitGoroutines(start), start, "goroutines leaked by the sync & delete wrapper")
}

func TestLoadFsWithoutSyncAndDelete(t *testing.T) {
	access := &confpar.Access{
		Fs:     "os",
		Params: map[string]string{"basePath": t.TempDir()},
	}

	start := runtime.NumGoroutine()

	_, closer := loadFsLoop(t, access)
	require.Nil(t, closer, "the os backend holds nothing")

	require.LessOrEqual(t, waitGoroutines(start), start)
}

// recordingFs is a backend that holds something to release, and records the order of the copies and of its closing
type recordingFs struct {
	afero.Fs
	mu     sync.Mutex
	events []string
}

func (r *recordingFs) record(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

func (r *recordingFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
		// A slow backend: the copy is still queued when the session ends
		time.Sleep(100 * time.Millisecond)
		r.record("copy " + name)
	}

	return r.Fs.OpenFile(name, flag, perm)
}

func (r *recordingFs) Close() error {
	r.record("close")

	return nil
}

func TestWrapFsCopiesBeforeClosingBackend(t *testing.T) {
	req := require.New(t)
	backend := &recordingFs{Fs: afero.NewMemMapFs()}
	access := &confpar.Access{
		SyncAndDelete: &confpar.SyncAndDelete{Enable: true, Directory: t.TempDir()},
	}

	fs, closer, err := wrapFs(access, backend, discardLogger())
	req.NoError(err)
	req.NotNil(closer)

	file, err := fs.OpenFile("/file", os.O_CREATE|os.O_WRONLY, 0o600)
	req.NoError(err)
	_, err = file.WriteString("content")
	req.NoError(err)
	req.NoError(file.Close())

	req.NoError(closer.Close())

	content, err := afero.ReadFile(backend.Fs, "/file")
	req.NoError(err)
	req.Equal("content", string(content))

	backend.mu.Lock()
	defer backend.mu.Unlock()
	req.Equal([]string{"copy /file", "close"}, backend.events, "the backend must be closed after the copy")
}

func TestWrapFsBackendCloserWithoutWrapper(t *testing.T) {
	backend := &recordingFs{Fs: afero.NewMemMapFs()}

	_, closer, err := wrapFs(&confpar.Access{ReadOnly: true}, backend, discardLogger())
	require.NoError(t, err)
	require.NotNil(t, closer)
	require.NoError(t, closer.Close())
	require.Equal(t, []string{"close"}, backend.events)
}

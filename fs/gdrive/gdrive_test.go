package gdrive

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	drvoa "github.com/fclairamb/afero-gdrive/oauthhelper"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/fclairamb/ftpserver/config/confpar"
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

// urlRecorder is a slog handler that records the authorization URLs logged by the loopback flow.
type urlRecorder struct {
	mu   sync.Mutex
	urls []string
}

func (r *urlRecorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *urlRecorder) Handle(_ context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "url" {
			r.mu.Lock()
			r.urls = append(r.urls, attr.Value.String())
			r.mu.Unlock()
		}

		return true
	})

	return nil
}

func (r *urlRecorder) WithAttrs([]slog.Attr) slog.Handler { return r }

func (r *urlRecorder) WithGroup(string) slog.Handler { return r }

func (r *urlRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.urls)
}

func gdriveAccess(tokenFile, timeout string) *confpar.Access {
	return &confpar.Access{
		User: "gdrive",
		Fs:   "gdrive",
		Params: map[string]string{
			"google_client_id":      "fake-client-id",
			"google_client_secret":  "fake-client-secret",
			"token_file":            tokenFile,
			"authorization_timeout": timeout,
		},
	}
}

// loadFsAsync runs LoadFs in a goroutine and returns the channel of its error.
func loadFsAsync(access *confpar.Access, logger *slog.Logger) <-chan error {
	result := make(chan error, 1)

	go func() {
		_, err := LoadFs(access, logger)
		result <- err
	}()

	return result
}

func waitErr(t *testing.T, result <-chan error, deadline time.Duration) error {
	t.Helper()

	select {
	case err := <-result:
		return err
	case <-time.After(deadline):
		t.Fatalf("LoadFs still blocked after %s", deadline)

		return nil
	}
}

func waitURLs(t *testing.T, recorder *urlRecorder, expected int) {
	t.Helper()
	require.Eventually(t, func() bool { return recorder.count() >= expected }, 2*time.Second, 10*time.Millisecond)
}

func TestFirstAuthorizationTimesOut(t *testing.T) {
	recorder := &urlRecorder{}
	access := gdriveAccess(filepath.Join(t.TempDir(), "token.json"), "200ms")

	err := waitErr(t, loadFsAsync(access, slog.New(recorder)), 2*time.Second)
	require.ErrorIs(t, err, ErrAuthorizationTimeout)
	require.Equal(t, 1, recorder.count(), "the flow should have started")
}

func TestInvalidAuthorizationTimeout(t *testing.T) {
	for _, value := range []string{"forever", "-1s", "0s"} {
		_, err := LoadFs(gdriveAccess(filepath.Join(t.TempDir(), "token.json"), value), slog.New(&urlRecorder{}))
		require.ErrorIs(t, err, ErrInvalidAuthorizationTimeout, value)
	}
}

func TestOnePendingAuthorizationPerTokenFile(t *testing.T) {
	dir := t.TempDir()
	firstFile := filepath.Join(dir, "first.json")
	recorder := &urlRecorder{}
	logger := slog.New(recorder)

	first := loadFsAsync(gdriveAccess(firstFile, "1s"), logger)
	waitURLs(t, recorder, 1)

	// Same token file: fails at once, without a second flow.
	start := time.Now()
	_, err := LoadFs(gdriveAccess(firstFile, "1s"), logger)
	require.ErrorIs(t, err, ErrAuthorizationPending)
	require.Less(t, time.Since(start), 500*time.Millisecond)
	require.Equal(t, 1, recorder.count())

	// Another token file: starts its own flow.
	other := loadFsAsync(gdriveAccess(filepath.Join(dir, "other.json"), "300ms"), logger)
	waitURLs(t, recorder, 2)
	require.ErrorIs(t, waitErr(t, other, 2*time.Second), ErrAuthorizationTimeout)

	// Once the first one timed out, a new login on its token file can start a flow again.
	require.ErrorIs(t, waitErr(t, first, 3*time.Second), ErrAuthorizationTimeout)

	again := loadFsAsync(gdriveAccess(firstFile, "200ms"), logger)
	require.ErrorIs(t, waitErr(t, again, 2*time.Second), ErrAuthorizationTimeout)
	require.Equal(t, 3, recorder.count())
}

// fakeTransport answers every Drive API call with a 404, counting the calls.
type fakeTransport struct {
	calls atomic.Int32
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls.Add(1)

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":404,"message":"not found"}}`)),
		Request:    req,
	}, nil
}

func TestExistingTokenSkipsAuthorization(t *testing.T) {
	transport := &fakeTransport{}
	previous := clientContext
	clientContext = func() context.Context {
		return context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: transport})
	}

	t.Cleanup(func() { clientContext = previous })

	tokenFile := filepath.Join(t.TempDir(), "token.json")
	require.NoError(t, storeToken(tokenFile, &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}))

	recorder := &urlRecorder{}
	// The fake Drive API answers 404 to the root lookup, so LoadFs fails, but after using the stored token.
	err := waitErr(t, loadFsAsync(gdriveAccess(tokenFile, "1s"), slog.New(recorder)), 2*time.Second)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrAuthorizationTimeout)
	require.Positive(t, transport.calls.Load(), "the Drive API should have been called with the stored token")
	require.Equal(t, 0, recorder.count(), "no authorization flow should have started")
}

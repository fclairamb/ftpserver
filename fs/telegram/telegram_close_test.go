package telegram

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// fakeTelegramAPI answers getMe and counts the getUpdates polls
func fakeTelegramAPI(t *testing.T, polls *atomic.Int32) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"test","username":"test_bot"}}`))
		case strings.HasSuffix(r.URL.Path, "/getUpdates"):
			polls.Add(1)
			time.Sleep(10 * time.Millisecond)

			_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
		default:
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		}
	}))
	t.Cleanup(server.Close)

	return server
}

// TestCloseStopsPoller makes sure Close stops the bot's poller, and can be called twice
func TestCloseStopsPoller(t *testing.T) {
	var polls atomic.Int32

	api := fakeTelegramAPI(t, &polls)

	fs, err := loadFs(&confpar.Access{
		Fs:     "telegram",
		Params: map[string]string{"Token": "123:abc", "ChatID": "42"},
	}, slog.Default(), api.URL)
	if err != nil {
		t.Fatalf("loadFs: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for polls.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the poller should poll the API once started")
		}

		time.Sleep(10 * time.Millisecond)
	}

	done := make(chan error, 1)
	go func() { done <- fs.Close() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Close: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Close should return once the poller is stopped")
	}

	after := polls.Load()
	time.Sleep(200 * time.Millisecond)

	if got := polls.Load(); got != after {
		t.Fatalf("the poller should stop polling after Close, got %d polls after %d", got, after)
	}

	if err := fs.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

// TestCloseNeverStarted makes sure Close doesn't wait for a poller that was never started
func TestCloseNeverStarted(t *testing.T) {
	fs := &Fs{}

	if err := fs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Once closed, the poller can't be started anymore
	fs.start()

	if fs.started {
		t.Fatal("a closed file system should not start its poller")
	}
}

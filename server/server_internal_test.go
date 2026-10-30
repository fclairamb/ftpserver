package server

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	serverlib "github.com/fclairamb/ftpserverlib"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/config/confpar"
)

func newTestServer(t *testing.T, content *confpar.Content) *Server {
	t.Helper()

	conf, err := config.FromContent(content, "test.json", slog.Default())
	if err != nil {
		t.Fatalf("FromContent: %v", err)
	}

	srv, err := NewServer(conf, slog.Default())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	return srv
}

// TestWebhookWithoutTimeout makes sure a webhook without timeout works (it used to time out instantly)
func TestWebhookWithoutTimeout(t *testing.T) {
	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if r.Header.Get("Authorization") != "Bearer token" || req["pass"] != "secret" {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		_ = json.NewEncoder(w).Encode(confpar.Access{
			User:   req["user"],
			Fs:     "os",
			Params: map[string]string{"basePath": "/tmp"},
		})
	}))
	defer webhook.Close()

	srv := newTestServer(t, &confpar.Content{
		AccessesWebhook: &confpar.AccessesWebhook{
			URL:     webhook.URL,
			Headers: map[string]string{"Authorization": "Bearer token"},
		},
	})

	access, err := srv.getAccessFromWebhook("alice", "secret")
	if err != nil {
		t.Fatalf("getAccessFromWebhook: %v", err)
	}

	if access.User != "alice" || access.Fs != "os" || access.Params["basePath"] != "/tmp" {
		t.Fatalf("unexpected access: %+v", access)
	}

	if _, err := srv.getAccessFromWebhook("alice", "wrong"); err == nil {
		t.Fatal("expected an error for a refused authentication")
	}
}

func readBanner(t *testing.T, addr string) (net.Conn, string) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("couldn't read banner: %v", err)
	}

	return conn, line
}

func TestMaxClients(t *testing.T) {
	srv := newTestServer(t, &confpar.Content{
		ListenAddress: "127.0.0.1:0",
		MaxClients:    1,
	})

	ftpServer := serverlib.NewFtpServer(srv)
	if err := ftpServer.Listen(); err != nil {
		t.Fatalf("Listen: %v", err)
	}

	go func() { _ = ftpServer.Serve() }()

	defer func() { _ = ftpServer.Stop() }()

	first, banner := readBanner(t, ftpServer.Addr())
	defer func() { _ = first.Close() }()

	if !strings.HasPrefix(banner, "220") {
		t.Fatalf("first client should be accepted, got: %q", banner)
	}

	second, banner := readBanner(t, ftpServer.Addr())
	defer func() { _ = second.Close() }()

	if !strings.Contains(banner, "Too many clients") {
		t.Fatalf("second client should be refused, got: %q", banner)
	}
}

// TestSettingsWithoutPassivePortRange makes sure no port range in the config gives a nil interface to ftpserverlib.
// A typed nil (*PortRange)(nil) made every passive transfer panic with "NumberAttempts called using nil *PortRange
// pointer" (#1825).
func TestSettingsWithoutPassivePortRange(t *testing.T) {
	srv := newTestServer(t, &confpar.Content{})

	settings, err := srv.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	if settings.PassiveTransferPortRange != nil {
		t.Fatalf("expected a nil port range, got %#v", settings.PassiveTransferPortRange)
	}
}

func TestSettingsWithPassivePortRange(t *testing.T) {
	srv := newTestServer(t, &confpar.Content{
		PassiveTransferPortRange: &confpar.PortRange{Start: 2122, End: 2130},
	})

	settings, err := srv.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	portRange, ok := settings.PassiveTransferPortRange.(*serverlib.PortRange)
	if !ok || portRange.Start != 2122 || portRange.End != 2130 || portRange.NumberAttempts() != 9 {
		t.Fatalf("unexpected port range: %#v", settings.PassiveTransferPortRange)
	}
}

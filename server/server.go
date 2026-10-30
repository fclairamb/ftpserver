// Package server contains the core FTP server code
package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/afero"

	serverlib "github.com/fclairamb/ftpserverlib"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"
	"github.com/fclairamb/ftpserver/fs/fslog"
)

// Server structure
type Server struct {
	config          *config.Config
	logger          *slog.Logger
	nbClients       uint32
	nbClientsSync   sync.Mutex
	zeroClientEvent chan error
	tlsOnce         sync.Once
	tlsConfig       *tls.Config
	tlsError        error
	accesses        *fsCache
}

type fsCache struct {
	sync.Mutex
	accesses map[string]afero.Fs
}

func newFsCache() *fsCache {
	return &fsCache{
		accesses: make(map[string]afero.Fs),
	}
}

// ErrTimeout is returned when an operation timeouts
var ErrTimeout = errors.New("timeout")

// ErrNotEnabled is returned when a feature hasn't been enabled
var ErrNotEnabled = errors.New("not enabled")

// ErrTooManyClients is returned when the max_clients limit has been reached
var ErrTooManyClients = errors.New("too many clients")

// ErrWebhookStatus is returned when the accesses webhook doesn't return a 200 status code
var ErrWebhookStatus = errors.New("unexpected webhook status code")

// defaultWebhookTimeout is used when the accesses webhook timeout isn't specified
const defaultWebhookTimeout = 30 * time.Second

// maxWebhookResponseSize is the maximum size of an accesses webhook response
const maxWebhookResponseSize = 1 << 20

// NewServer creates a server instance
func NewServer(config *config.Config, logger *slog.Logger) (*Server, error) {
	return &Server{
		config:   config,
		logger:   logger,
		accesses: newFsCache(),
	}, nil
}

// GetSettings returns some general settings around the server setup
func (s *Server) GetSettings() (*serverlib.Settings, error) {
	conf := s.config.Content

	// An interface, not a *PortRange: a nil *PortRange in the PasvPortGetter field is not a nil interface, and
	// ftpserverlib then called NumberAttempts on it at the first passive transfer (#1825).
	var portRange serverlib.PasvPortGetter

	if conf.PassiveTransferPortRange != nil {
		portRange = &serverlib.PortRange{
			Start: conf.PassiveTransferPortRange.Start,
			End:   conf.PassiveTransferPortRange.End,
		}
	}

	var tlsRequired serverlib.TLSRequirement
	switch conf.TLSRequired {
	case "ImplicitEncryption":
		tlsRequired = serverlib.ImplicitEncryption
	case "MandatoryEncryption":
		tlsRequired = serverlib.MandatoryEncryption
	default:
		tlsRequired = serverlib.ClearOrEncrypted
	}

	return &serverlib.Settings{
		ListenAddr:               conf.ListenAddress,
		PublicHost:               conf.PublicHost,
		PassiveTransferPortRange: portRange,
		TLSRequired:              tlsRequired,
		IdleTimeout:              int(conf.IdleTimeout.Seconds()),
		EnableHASH:               conf.Extensions.EnableHASH,
		// Default to binary transfers. Without this, the transfer type defaults to ASCII
		// (the zero value of TransferType), which silently corrupts binary files (e.g. .exe)
		// when a client uploads/downloads without first issuing a "TYPE I" command. See #1532.
		DefaultTransferType: serverlib.TransferTypeBinary,
	}, nil
}

// ReloadConfig reloads the configuration file. Shared file systems are dropped so that they get re-created with
// the new configuration.
func (s *Server) ReloadConfig() error {
	if err := s.config.Load(); err != nil {
		return err
	}

	s.accesses.Lock()
	defer s.accesses.Unlock()

	s.accesses.accesses = make(map[string]afero.Fs)

	return nil
}

// ClientConnected is called to send the very first welcome message
func (s *Server) ClientConnected(cc serverlib.ClientContext) (string, error) {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()
	s.nbClients++
	s.logger.Info(
		"Client connected",
		"clientId", cc.ID(),
		"remoteAddr", cc.RemoteAddr(),
		"nbClients", s.nbClients,
	)

	// The client is counted until it disconnects, which also happens when we refuse it
	if maxClients := s.config.Content.MaxClients; maxClients > 0 && int64(s.nbClients) > int64(maxClients) {
		s.logger.Warn("Too many clients, refusing connection", "clientId", cc.ID(), "maxClients", maxClients)

		return "Too many clients, try again later", ErrTooManyClients
	}

	if s.config.Content.Logging.FtpExchanges {
		cc.SetDebug(true)
	}

	return "ftpserver", nil
}

// ClientDisconnected is called when the user disconnects, even if they never authenticated
func (s *Server) ClientDisconnected(cc serverlib.ClientContext) {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()

	s.nbClients--

	s.logger.Info(
		"Client disconnected",
		"clientId", cc.ID(),
		"remoteAddr", cc.RemoteAddr(),
		"nbClients", s.nbClients,
	)
	s.considerEnd()
}

// Stop will trigger a graceful stop of the server. All currently connected clients won't be disconnected instantly.
func (s *Server) Stop() {
	s.nbClientsSync.Lock()
	defer s.nbClientsSync.Unlock()
	s.zeroClientEvent = make(chan error, 1)
	s.considerEnd()
}

// WaitGracefully allows to gracefully wait for all currently connected clients before disconnecting
func (s *Server) WaitGracefully(timeout time.Duration) error {
	s.logger.Info("Waiting for last client to disconnect...")

	s.nbClientsSync.Lock()
	zeroClientEvent := s.zeroClientEvent
	s.nbClientsSync.Unlock()

	if zeroClientEvent == nil {
		// Stop wasn't called, there's nothing to wait for
		return nil
	}

	select {
	case err := <-zeroClientEvent:
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

func (s *Server) considerEnd() {
	if s.nbClients == 0 && s.zeroClientEvent != nil {
		s.zeroClientEvent <- nil
		close(s.zeroClientEvent)
	}
}

func (s *Server) loadFs(access *confpar.Access) (afero.Fs, error) {
	// The keycloak backend checks the user's credentials when creating its file system, a cached instance would
	// skip this check.
	if access.Fs == "keycloak" {
		return fs.LoadFs(access, s.logger)
	}

	cache := s.accesses
	cache.Lock()
	defer cache.Unlock()

	if cachedFs := cache.accesses[access.User]; cachedFs != nil {
		s.logger.Debug("Reusing fs instance", "user", access.User)

		return cachedFs, nil
	}

	newFs, err := fs.LoadFs(access, s.logger)
	if err != nil {
		return nil, err
	}
	if access.Shared {
		s.logger.Debug("Saving fs instance for later use", "user", access.User, "fsType", newFs.Name())
		cache.accesses[access.User] = newFs
	}

	return newFs, err
}

func (s *Server) getAccessFromWebhook(user, pass string) (*confpar.Access, error) {
	// Convert payload to JSON
	jsonData, err := json.Marshal(map[string]string{
		"user": user,
		"pass": pass,
	})
	if err != nil {
		return nil, err
	}

	webhook := s.config.Content.AccessesWebhook

	timeout := webhook.Timeout.Duration
	if timeout <= 0 {
		timeout = defaultWebhookTimeout
	}

	// Timeout is implemented with context termination
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrWebhookStatus, resp.StatusCode)
	}

	access := new(confpar.Access)
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxWebhookResponseSize)).Decode(access); err != nil {
		return nil, err
	}

	return access, nil
}

// AuthUser authenticates the user and selects an handling driver
func (s *Server) AuthUser(cc serverlib.ClientContext, user, pass string) (serverlib.ClientDriver, error) {
	var (
		access    *confpar.Access
		errAccess error
	)

	if s.config.Content.AccessesWebhook == nil {
		// Get the access from the configuration
		access, errAccess = s.config.GetAccess(user, pass)
	} else {
		// Get the access from the webhook, not the configuration
		access, errAccess = s.getAccessFromWebhook(user, pass)
	}
	if errAccess != nil {
		return nil, errAccess
	}

	accFs, errFs := s.loadFs(access)

	if errFs != nil {
		return nil, errFs
	}

	if s.config.Content.Logging.FtpExchanges || access.Logging.FtpExchanges {
		cc.SetDebug(true)
	}

	if s.config.Content.Logging.FileAccesses || access.Logging.FileAccesses {
		var err error

		logger := s.logger.With(
			"userName", user,
			"fs", access.Fs,
			"clientId", cc.ID(),
			"remoteAddr", cc.RemoteAddr(),
		)

		accFs, err = fslog.LoadFS(accFs, logger)

		if err != nil {
			return nil, err
		}
	}

	return &ClientDriver{
		Fs: accFs,
	}, nil
}

// The ClientDriver is the internal structure used for handling the client. At this stage it's limited to the afero.Fs
type ClientDriver struct {
	afero.Fs
}

// Symlink creates a symbolic link. It implements ftpserverlib's
// ClientDriverExtensionSymlink interface (the "SITE SYMLINK" command), delegating to
// the underlying filesystem when it supports symlinks (the local OS backend does).
// Backends that can't create symlinks return an *os.LinkError wrapping
// afero.ErrNoSymlink. See #980.
func (d *ClientDriver) Symlink(oldname, newname string) error {
	if linker, ok := d.Fs.(afero.Linker); ok {
		return linker.SymlinkIfPossible(oldname, newname)
	}

	return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: afero.ErrNoSymlink}
}

func (s *Server) loadTLSConfig() (*tls.Config, error) {
	tlsConf := s.config.Content.TLS
	if tlsConf == nil || tlsConf.ServerCert == nil {
		return nil, ErrNotEnabled
	}

	serverCert := tlsConf.ServerCert

	certBytes, errReadFileCert := os.ReadFile(serverCert.Cert)
	if errReadFileCert != nil {
		return nil, fmt.Errorf("could not load cert file: %s: %w", serverCert.Cert, errReadFileCert)
	}

	keyBytes, errReadFileKey := os.ReadFile(serverCert.Key)
	if errReadFileKey != nil {
		return nil, fmt.Errorf("could not load key file: %s: %w", serverCert.Key, errReadFileKey)
	}

	keypair, errKeyPair := tls.X509KeyPair(certBytes, keyBytes)
	if errKeyPair != nil {
		return nil, fmt.Errorf("could not parse key pairs: %w", errKeyPair)
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{keypair},
	}, nil
}

// GetTLSConfig returns a TLS Certificate to use
// The certificate could frequently change if we use something like "let's encrypt"
func (s *Server) GetTLSConfig() (*tls.Config, error) {
	// The function is called every single time a control or transfer connection requires a TLS connection. As such
	// it's important to cache it.
	s.tlsOnce.Do(func() {
		s.tlsConfig, s.tlsError = s.loadTLSConfig()
	})

	return s.tlsConfig, s.tlsError
}

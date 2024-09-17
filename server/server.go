// Package server contains the core FTP server code
package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/spf13/afero"

	serverlib "github.com/fclairamb/ftpserverlib"
	log "github.com/fclairamb/go-log"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"
	"github.com/fclairamb/ftpserver/fs/fslog"
)

// Server structure
type Server struct { // nolint: maligned
	config          *config.Config
	logger          log.Logger
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

// ErrNotImplemented is returned when we're using something that has not been implemented yet
// var ErrNotImplemented = errors.New("not implemented")

// ErrNotEnabled is returned when a feature hasn't been enabled
var ErrNotEnabled = errors.New("not enabled")

// NewServer creates a server instance
func NewServer(config *config.Config, logger log.Logger) (*Server, error) {
	return &Server{
		config:   config,
		logger:   logger,
		accesses: newFsCache(),
	}, nil
}

// GetSettings returns some general settings around the server setup
func (s *Server) GetSettings() (*serverlib.Settings, error) {
	conf := s.config.Content

	var portRange *serverlib.PortRange

	if conf.PassiveTransferPortRange != nil {
		portRange = &serverlib.PortRange{
			Start: conf.PassiveTransferPortRange.Start,
			End:   conf.PassiveTransferPortRange.End,
		}
	}

	return &serverlib.Settings{
		ListenAddr:               conf.ListenAddress,
		PublicHost:               conf.PublicHost,
		PassiveTransferPortRange: portRange,
	}, nil
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

	if s.config.Content.Logging.FtpExchanges {
		cc.SetDebug(true)
	}

	return "ftpserver", nil
}

// ClientDisconnected is called when the user disconnects, even if he never authenticated
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

	defer func() { s.zeroClientEvent = nil }()

	select {
	case err := <-s.zeroClientEvent:
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

// AuthUser authenticates the user and selects an handling driver
func (s *Server) AuthUser(cc serverlib.ClientContext, user, pass string) (serverlib.ClientDriver, error) {
	access, errAccess := s.config.GetAccess(user, pass)
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

func (s *Server) loadTLSConfig() (*tls.Config, error) {
	tlsConf := s.config.Content.TLS
	if tlsConf == nil || tlsConf.ServerCert == nil {
		return nil, ErrNotEnabled
	}

	serverCert := tlsConf.ServerCert

	certBytes, errReadFileCert := ioutil.ReadFile(serverCert.Cert)
	if errReadFileCert != nil {
		return nil, fmt.Errorf("could not load cert file: %s: %w", serverCert.Cert, errReadFileCert)
	}

	keyBytes, errReadFileKey := ioutil.ReadFile(serverCert.Key)
	if errReadFileKey != nil {
		return nil, fmt.Errorf("could not load key file: %s: %w", serverCert.Cert, errReadFileCert)
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

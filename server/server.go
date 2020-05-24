// Package server contains the core FTP server code
package server

import (
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"github.com/spf13/afero"

	serverlib "github.com/fclairamb/ftpserverlib"
	"github.com/fclairamb/ftpserverlib/log"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/fs"
)

// Server structure
type Server struct {
	config          *config.Config
	logger          log.Logger
	nbClients       uint32
	nbClientsSync   sync.Mutex
	zeroClientEvent chan error
}

// ErrTimeout is returned when an operation timeouts
var ErrTimeout = errors.New("timeout")

// NewServer creates a server instance
func NewServer(config *config.Config, logger log.Logger) (*Server, error) {
	return &Server{
		config: config,
		logger: logger,
	}, nil
}

// GetSettings returns some general settings around the server setup
func (s *Server) GetSettings() (*serverlib.Settings, error) {
	conf := s.config.Content

	var portRange *serverlib.PortRange = nil

	if conf.PassiveTransferPortRange != nil {
		portRange = &serverlib.PortRange{
			Start: conf.PassiveTransferPortRange.Start,
			End:   conf.PassiveTransferPortRange.End,
		}
	}

	return &serverlib.Settings{
		ListenAddr:               conf.ListenAddress,
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
	cc.SetDebug(true)

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

// AuthUser authenticates the user and selects an handling driver
func (s *Server) AuthUser(cc serverlib.ClientContext, user, pass string) (serverlib.ClientDriver, error) {
	access, errAccess := s.config.GetAccess(user, pass)
	if errAccess != nil {
		return nil, errAccess
	}

	accFs, errFs := fs.LoadFs(access)

	if errFs != nil {
		return nil, errFs
	}

	return &ClientDriver{
		Fs: accFs,
	}, nil
}

// The ClientDriver is the internal structure used for handling the client. At this stage it's limited to the afero.Fs
type ClientDriver struct {
	afero.Fs
}

// GetTLSConfig returns a TLS Certificate to use
// The certificate could frequently change if we use something like "let's encrypt"
func (s *Server) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("not implemented")
}

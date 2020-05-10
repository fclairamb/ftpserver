package server

import (
	"crypto/tls"
	"errors"
	"github.com/fclairamb/ftpserver/config"
	fs "github.com/fclairamb/ftpserver/fs"
	serverlib "github.com/fclairamb/ftpserverlib"
	"github.com/fclairamb/ftpserverlib/log"
	"github.com/spf13/afero"
)

type Server struct {
	config *config.Config
	logger log.Logger
}

func NewServer(config *config.Config, logger log.Logger) (*Server, error) {
	return &Server{
		config: config,
		logger: logger,
	}, nil
}

// GetSettings returns some general settings around the server setup
func (s *Server) GetSettings() (*serverlib.Settings, error) {
	conf := s.config.Content
	return &serverlib.Settings{
		ListenAddr: conf.ListenAddress,
	}, nil
}

// WelcomeUser is called to send the very first welcome message
func (s *Server) ClientConnected(cc serverlib.ClientContext) (string, error) {
	s.logger.Info("Client connected", "clientId", cc.ID(), "remoteAddr", cc.RemoteAddr())
	cc.SetDebug(true)
	return "ftpserver", nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (s *Server) ClientDisconnected(cc serverlib.ClientContext) {
	s.logger.Info("Client disconnected", "clientId", cc.ID(), "remoteAddr", cc.RemoteAddr())
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

type ClientDriver struct {
	afero.Fs
}

// GetTLSConfig returns a TLS Certificate to use
// The certificate could frequently change if we use something like "let's encrypt"
func (s *Server) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("not implemented")
}

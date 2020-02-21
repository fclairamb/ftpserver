// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"fmt"
	"net"

	"github.com/fclairamb/ftpserver/server/log"
)

const (
	// logKeyMsg is the human-readable part of the log
	logKeyMsg = "msg"
	// logKeyAction is the machine-readable part of the log
	logKeyAction = "action"
)

// CommandDescription defines which function should be used and if it should be open to anyone or only logged in users
type CommandDescription struct {
	Open bool                       // Open to clients without auth
	Fn   func(*clientHandler) error // Function to handle it
}

// This is shared between FtpServer instances as there's no point in making the FTP commands behave differently
// between them.
var commandsMap = map[string]*CommandDescription{
	// Authentication
	"USER": {Fn: (*clientHandler).handleUSER, Open: true},
	"PASS": {Fn: (*clientHandler).handlePASS, Open: true},

	// TLS handling
	"AUTH": {Fn: (*clientHandler).handleAUTH, Open: true},
	"PROT": {Fn: (*clientHandler).handlePROT, Open: true},
	"PBSZ": {Fn: (*clientHandler).handlePBSZ, Open: true},

	// Misc
	"CLNT": {Fn: (*clientHandler).handleCLNT, Open: true},
	"FEAT": {Fn: (*clientHandler).handleFEAT, Open: true},
	"SYST": {Fn: (*clientHandler).handleSYST, Open: true},
	"NOOP": {Fn: (*clientHandler).handleNOOP, Open: true},
	"OPTS": {Fn: (*clientHandler).handleOPTS, Open: true},

	// File access
	"SIZE": {Fn: (*clientHandler).handleSIZE},
	"STAT": {Fn: (*clientHandler).handleSTAT},
	"MDTM": {Fn: (*clientHandler).handleMDTM},
	"MFMT": {Fn: (*clientHandler).handleMFMT},
	"RETR": {Fn: (*clientHandler).handleRETR},
	"STOR": {Fn: (*clientHandler).handleSTOR},
	"APPE": {Fn: (*clientHandler).handleAPPE},
	"DELE": {Fn: (*clientHandler).handleDELE},
	"RNFR": {Fn: (*clientHandler).handleRNFR},
	"RNTO": {Fn: (*clientHandler).handleRNTO},
	"ALLO": {Fn: (*clientHandler).handleALLO},
	"REST": {Fn: (*clientHandler).handleREST},
	"SITE": {Fn: (*clientHandler).handleSITE},

	// Directory handling
	"CWD":  {Fn: (*clientHandler).handleCWD},
	"PWD":  {Fn: (*clientHandler).handlePWD},
	"CDUP": {Fn: (*clientHandler).handleCDUP},
	"NLST": {Fn: (*clientHandler).handleNLST},
	"LIST": {Fn: (*clientHandler).handleLIST},
	"MLSD": {Fn: (*clientHandler).handleMLSD},
	"MLST": {Fn: (*clientHandler).handleMLST},
	"MKD":  {Fn: (*clientHandler).handleMKD},
	"RMD":  {Fn: (*clientHandler).handleRMD},

	// Connection handling
	"TYPE": {Fn: (*clientHandler).handleTYPE},
	"PASV": {Fn: (*clientHandler).handlePASV},
	"EPSV": {Fn: (*clientHandler).handlePASV},
	"PORT": {Fn: (*clientHandler).handlePORT},
	"QUIT": {Fn: (*clientHandler).handleQUIT, Open: true},
}

// FtpServer is where everything is stored
// We want to keep it as simple as possible
type FtpServer struct {
	Logger        log.Logger   // Go-Kit logger
	settings      *Settings    // General settings
	listener      net.Listener // listener used to receive files
	clientCounter uint32       // Clients counter
	driver        MainDriver   // Driver to handle the client authentication and the file access driver selection
}

func (server *FtpServer) loadSettings() error {
	s, err := server.driver.GetSettings()

	if err != nil {
		return err
	}

	if s.Listener == nil && s.ListenAddr == "" {
		s.ListenAddr = "0.0.0.0:2121"
	}

	// florent(2018-01-14): #58: IDLE timeout: Default idle timeout will be set at 900 seconds
	if s.IdleTimeout == 0 {
		s.IdleTimeout = 900
	}

	if s.ConnectionTimeout == 0 {
		s.ConnectionTimeout = 30
	}

	server.settings = s

	return nil
}

// Listen starts the listening
// It's not a blocking call
func (server *FtpServer) Listen() error {
	err := server.loadSettings()

	if err != nil {
		return fmt.Errorf("could not load settings: %v", err)
	}

	if server.settings.Listener != nil {
		server.listener = server.settings.Listener
	} else {
		server.listener, err = net.Listen("tcp", server.settings.ListenAddr)

		if err != nil {
			server.Logger.Error(logKeyMsg, "Cannot listen", "err", err)
			return err
		}
	}

	server.Logger.Info(logKeyMsg, "Listening...", logKeyAction, "ftp.listening", "address", server.listener.Addr())

	return err
}

// Serve accepts and processes any new incoming client
func (server *FtpServer) Serve() {
	for {
		connection, err := server.listener.Accept()
		if err != nil {
			if server.listener != nil {
				server.Logger.Error(logKeyMsg, "Accept error", "err", err)
			}

			break
		}

		server.clientArrival(connection)
	}
}

// ListenAndServe simply chains the Listen and Serve method calls
func (server *FtpServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	server.Logger.Info(logKeyMsg, "Starting...", logKeyAction, "ftp.starting")

	server.Serve()

	// At this precise time, the clients are still connected. We are just not accepting clients anymore.

	return nil
}

// NewFtpServer creates a new FtpServer instance
func NewFtpServer(driver MainDriver) *FtpServer {
	return &FtpServer{
		driver: driver,
		Logger: log.NewNopGKLogger(),
	}
}

// Addr shows the listening address
func (server *FtpServer) Addr() string {
	if server.listener != nil {
		return server.listener.Addr().String()
	}

	return ""
}

// Stop closes the listener
func (server *FtpServer) Stop() {
	if server.listener != nil {
		if err := server.listener.Close(); err != nil {
			server.Logger.Warn(
				"msg", "Could not close listener",
				"action", "err.closing_listener",
				"err", err,
			)
		}
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) clientArrival(conn net.Conn) {
	server.clientCounter++
	id := server.clientCounter

	c := server.newClientHandler(conn, id)
	go c.HandleCommands()

	c.logger.Info(logKeyMsg, "FTP Client connected", logKeyAction, "ftp.connected", "clientIp", conn.RemoteAddr())
}

// clientDeparture
func (server *FtpServer) clientDeparture(c *clientHandler) {
	c.logger.Info(logKeyMsg, "FTP Client disconnected", logKeyAction, "ftp.disconnected", "clientIp", c.conn.RemoteAddr())
}

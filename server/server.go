// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"fmt"
	"net"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	// logKeyMsg is the human-readable part of the log
	logKeyMsg = "msg"
	// logKeyAction is the machine-readable part of the log
	logKeyAction = "action"
)

// CommandDescription defines which function should be used and if it should be open to anyone or only logged in users
type CommandDescription struct {
	Open bool                 // Open to clients without auth
	Fn   func(*clientHandler) // Function to handle it
}

var commandsMap map[string]*CommandDescription

func init() {
	// This is shared between FtpServer instances as there's no point in making the FTP commands behave differently
	// between them.

	commandsMap = make(map[string]*CommandDescription)

	// Authentication
	commandsMap["USER"] = &CommandDescription{Fn: (*clientHandler).handleUSER, Open: true}
	commandsMap["PASS"] = &CommandDescription{Fn: (*clientHandler).handlePASS, Open: true}

	// TLS handling
	commandsMap["AUTH"] = &CommandDescription{Fn: (*clientHandler).handleAUTH, Open: true}
	commandsMap["PROT"] = &CommandDescription{Fn: (*clientHandler).handlePROT, Open: true}
	commandsMap["PBSZ"] = &CommandDescription{Fn: (*clientHandler).handlePBSZ, Open: true}

	// Misc
	commandsMap["FEAT"] = &CommandDescription{Fn: (*clientHandler).handleFEAT, Open: true}
	commandsMap["SYST"] = &CommandDescription{Fn: (*clientHandler).handleSYST, Open: true}
	commandsMap["NOOP"] = &CommandDescription{Fn: (*clientHandler).handleNOOP, Open: true}
	commandsMap["OPTS"] = &CommandDescription{Fn: (*clientHandler).handleOPTS, Open: true}

	// File access
	commandsMap["SIZE"] = &CommandDescription{Fn: (*clientHandler).handleSIZE}
	commandsMap["STAT"] = &CommandDescription{Fn: (*clientHandler).handleSTAT}
	commandsMap["MDTM"] = &CommandDescription{Fn: (*clientHandler).handleMDTM}
	commandsMap["RETR"] = &CommandDescription{Fn: (*clientHandler).handleRETR}
	commandsMap["STOR"] = &CommandDescription{Fn: (*clientHandler).handleSTOR}
	commandsMap["APPE"] = &CommandDescription{Fn: (*clientHandler).handleAPPE}
	commandsMap["DELE"] = &CommandDescription{Fn: (*clientHandler).handleDELE}
	commandsMap["RNFR"] = &CommandDescription{Fn: (*clientHandler).handleRNFR}
	commandsMap["RNTO"] = &CommandDescription{Fn: (*clientHandler).handleRNTO}
	commandsMap["ALLO"] = &CommandDescription{Fn: (*clientHandler).handleALLO}
	commandsMap["REST"] = &CommandDescription{Fn: (*clientHandler).handleREST}
	commandsMap["SITE"] = &CommandDescription{Fn: (*clientHandler).handleSITE}

	// Directory handling
	commandsMap["CWD"] = &CommandDescription{Fn: (*clientHandler).handleCWD}
	commandsMap["PWD"] = &CommandDescription{Fn: (*clientHandler).handlePWD}
	commandsMap["CDUP"] = &CommandDescription{Fn: (*clientHandler).handleCDUP}
	commandsMap["NLST"] = &CommandDescription{Fn: (*clientHandler).handleLIST}
	commandsMap["LIST"] = &CommandDescription{Fn: (*clientHandler).handleLIST}
	commandsMap["MLSD"] = &CommandDescription{Fn: (*clientHandler).handleMLSD}
	commandsMap["MKD"] = &CommandDescription{Fn: (*clientHandler).handleMKD}
	commandsMap["RMD"] = &CommandDescription{Fn: (*clientHandler).handleRMD}

	// Connection handling
	commandsMap["TYPE"] = &CommandDescription{Fn: (*clientHandler).handleTYPE}
	commandsMap["PASV"] = &CommandDescription{Fn: (*clientHandler).handlePASV}
	commandsMap["EPSV"] = &CommandDescription{Fn: (*clientHandler).handlePASV}
	commandsMap["PORT"] = &CommandDescription{Fn: (*clientHandler).handlePORT}
	commandsMap["QUIT"] = &CommandDescription{Fn: (*clientHandler).handleQUIT, Open: true}
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
			level.Error(server.Logger).Log(logKeyMsg, "Cannot listen", "err", err)
			return err
		}
	}

	level.Info(server.Logger).Log(logKeyMsg, "Listening...", logKeyAction, "ftp.listening", "address", server.listener.Addr())

	return err
}

// Serve accepts and process any new client coming
func (server *FtpServer) Serve() {
	for {
		connection, err := server.listener.Accept()
		if err != nil {
			if server.listener != nil {
				level.Error(server.Logger).Log(logKeyMsg, "Accept error", "err", err)
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

	level.Info(server.Logger).Log(logKeyMsg, "Starting...", logKeyAction, "ftp.starting")

	server.Serve()

	// Note: At this precise time, the clients are still connected. We are just not accepting clients anymore.

	return nil
}

// NewFtpServer creates a new FtpServer instance
func NewFtpServer(driver MainDriver) *FtpServer {
	return &FtpServer{
		driver: driver,
		Logger: log.NewNopLogger(),
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
		server.listener.Close()
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) clientArrival(conn net.Conn) error {
	server.clientCounter++
	id := server.clientCounter

	c := server.newClientHandler(conn, id)
	go c.HandleCommands()

	level.Info(c.logger).Log(logKeyMsg, "FTP Client connected", logKeyAction, "ftp.connected", "clientIp", conn.RemoteAddr())

	return nil
}

// clientDeparture
func (server *FtpServer) clientDeparture(c *clientHandler) {
	level.Info(c.logger).Log(logKeyMsg, "FTP Client disconnected", logKeyAction, "ftp.disconnected", "clientIp", c.conn.RemoteAddr())
}

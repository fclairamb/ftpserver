// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
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
	commandsMap["USER"] = &CommandDescription{Fn: (*clientHandler).handleUSER, Open: true }
	commandsMap["PASS"] = &CommandDescription{Fn: (*clientHandler).handlePASS, Open: true }

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

	// Directory handling
	commandsMap["CWD"] = &CommandDescription{Fn: (*clientHandler).handleCWD}
	commandsMap["PWD"] = &CommandDescription{Fn: (*clientHandler).handlePWD}
	commandsMap["CDUP"] = &CommandDescription{Fn: (*clientHandler).handleCDUP}
	commandsMap["NLST"] = &CommandDescription{Fn: (*clientHandler).handleLIST}
	commandsMap["LIST"] = &CommandDescription{Fn: (*clientHandler).handleLIST}
	commandsMap["MKD"] = &CommandDescription{Fn: (*clientHandler).handleMKD}
	commandsMap["RMD"] = &CommandDescription{Fn: (*clientHandler).handleRMD}

	// Connection handling
	commandsMap["TYPE"] = &CommandDescription{Fn: (*clientHandler).handleTYPE}
	commandsMap["PASV"] = &CommandDescription{Fn: (*clientHandler).handlePASV}
	commandsMap["EPSV"] = &CommandDescription{Fn: (*clientHandler).handlePASV}
	commandsMap["PORT"] = &CommandDescription{Fn: (*clientHandler).handlePORT}
	commandsMap["QUIT"] = &CommandDescription{Fn: (*clientHandler).handleQUIT, Open: true}
}

type FtpServer struct {
	Settings         *Settings                 // General settings
	Listener         net.Listener              // Listener used to receive files
	StartTime        time.Time                 // Time when the server was started
	connectionsById  map[uint32]*clientHandler // Connections map
	connectionsMutex sync.RWMutex              // Connections map sync
	clientCounter    uint32                    // Clients counter
	driver           ServerDriver              // Driver to handle the client authentication and the file access driver selection
}

func (server *FtpServer) loadSettings() {
	s := server.driver.GetSettings()
	if s.ListenHost == "" {
		s.ListenHost = "0.0.0.0"
	}

	if s.ListenPort == 0 { // For the default value (0)
		// We take the default port (2121)
		s.ListenPort = 2121
	} else if s.ListenPort == -1 { // For the automatic value
		// We let the system decide (0)
		s.ListenPort = 0
	}
	if s.MaxConnections == 0 {
		s.MaxConnections = 10000
	}
	server.Settings = s
}

func (server *FtpServer) Listen() error {
	server.loadSettings()
	var err error

	server.Listener, err = net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", server.Settings.ListenHost, server.Settings.ListenPort),
	)

	if err != nil {
		log15.Error("Cannot listen", "err", err)
		return err
	}

	log15.Info("Listening...", "address", server.Listener.Addr())

	return err
}

func (server *FtpServer) Serve() {
	for {
		connection, err := server.Listener.Accept()
		if err != nil {
			if server.Listener != nil {
				log15.Error("Accept error", "err", err)
			}
			break
		} else {
			c := server.NewClientHandler(connection)
			go c.HandleCommands()
		}
	}
}

func (server *FtpServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	log15.Info("Starting...")

	server.Serve()

	// Note: At this precise time, the clients are still connected. We are just not accepting clients anymore.

	return nil
}

func NewFtpServer(driver ServerDriver) *FtpServer {
	return &FtpServer{
		driver:          driver,
		StartTime:       time.Now().UTC(), // Might make sense to put it in Start method
		connectionsById: make(map[uint32]*clientHandler),
	}
}

func (server *FtpServer) Stop() {
	if server.Listener != nil {
		l := server.Listener
		server.Listener = nil
		l.Close()
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) clientArrival(c *clientHandler) error {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	server.connectionsById[c.Id] = c
	nb := len(server.connectionsById)

	log15.Info("FTP Client connected", "action", "ftp.connected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", nb)

	if nb > server.Settings.MaxConnections {
		return errors.New(fmt.Sprintf("Too many clients %d > %d", nb, server.Settings.MaxConnections))
	} else {
		return nil
	}
}

// When a client leaves
func (server *FtpServer) clientDeparture(c *clientHandler) {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	delete(server.connectionsById, c.Id)

	log15.Info("FTP Client disconnected", "action", "ftp.disconnected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", len(server.connectionsById))
}

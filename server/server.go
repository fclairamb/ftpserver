package server

import (
	"time"
	"net"
	"sync"
	"gopkg.in/inconshreveable/log15.v2"
	"errors"
	"fmt"
)

var commandsMap map[string]func(*clientHandler)

func init() {
	// This is shared between FtpServer instances as there's no point in making the FTP commands behave differently
	// between them.

	commandsMap = make(map[string]func(*clientHandler))

	// Authentication
	commandsMap["USER"] = (*clientHandler).handleUSER
	commandsMap["PASS"] = (*clientHandler).handlePASS

	// File access
	commandsMap["STAT"] = (*clientHandler).handleSTAT
	commandsMap["SIZE"] = (*clientHandler).handleSIZE
	commandsMap["MDTM"] = (*clientHandler).handleMDTM
	commandsMap["RETR"] = (*clientHandler).handleRETR
	commandsMap["STOR"] = (*clientHandler).handleSTOR
	commandsMap["APPE"] = (*clientHandler).handleAPPE
	commandsMap["DELE"] = (*clientHandler).handleDELE
	commandsMap["RNFR"] = (*clientHandler).handleRNFR
	commandsMap["RNTO"] = (*clientHandler).handleRNTO

	// Directory handling
	commandsMap["CWD"] = (*clientHandler).handleCWD
	commandsMap["PWD"] = (*clientHandler).handlePWD
	commandsMap["CDUP"] = (*clientHandler).handleCDUP
	commandsMap["NLST"] = (*clientHandler).handleLIST
	commandsMap["LIST"] = (*clientHandler).handleLIST
	commandsMap["MKD"] = (*clientHandler).handleMKD
	commandsMap["RMD"] = (*clientHandler).handleRMD

	// Connection handling
	commandsMap["TYPE"] = (*clientHandler).handleTYPE
	commandsMap["PASV"] = (*clientHandler).handlePASV
	commandsMap["EPSV"] = (*clientHandler).handlePASV
	commandsMap["QUIT"] = (*clientHandler).handleQUIT

	// Misc
	commandsMap["FEAT"] = (*clientHandler).handleFEAT
	commandsMap["SYST"] = (*clientHandler).handleSYST
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

func NewFtpServer(driver ServerDriver) *FtpServer {
	return &FtpServer{
		driver: driver,
		StartTime: time.Now().UTC(), // Might make sense to put it in Start method
		connectionsById: make(map[uint32]*clientHandler),
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) clientArrival(c *clientHandler) error {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()


	server.connectionsById[c.Id] = c
	nb := len(server.connectionsById)

	log15.Info("Client connected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", nb)

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

	log15.Info("Client disconnected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", len(server.connectionsById))
}

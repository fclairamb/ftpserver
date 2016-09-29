package server

import (
	"time"
	"net"
	"sync"
	"gopkg.in/inconshreveable/log15.v2"
	"errors"
	"fmt"
)

var commandsMap map[string]func(*ClientHandler)

func init() {
	// This is shared between FtpServer instances as there's no point in making the FTP commands behave differently
	// between them.

	commandsMap = make(map[string]func(*ClientHandler))

	// Authentication
	commandsMap["USER"] = (*ClientHandler).handleUSER
	commandsMap["PASS"] = (*ClientHandler).handlePASS

	// File access
	commandsMap["STAT"] = (*ClientHandler).handleSTAT
	commandsMap["SIZE"] = (*ClientHandler).handleSIZE
	commandsMap["MDTM"] = (*ClientHandler).handleMDTM
	commandsMap["RETR"] = (*ClientHandler).handleRETR
	commandsMap["STOR"] = (*ClientHandler).handleSTOR
	commandsMap["APPE"] = (*ClientHandler).handleAPPE
	commandsMap["DELE"] = (*ClientHandler).handleDELE
	commandsMap["RNFR"] = (*ClientHandler).handleRNFR
	commandsMap["RNTO"] = (*ClientHandler).handleRNTO

	// Directory handling
	commandsMap["CWD"] = (*ClientHandler).handleCWD
	commandsMap["PWD"] = (*ClientHandler).handlePWD
	commandsMap["CDUP"] = (*ClientHandler).handleCDUP
	commandsMap["NLST"] = (*ClientHandler).handleLIST
	commandsMap["LIST"] = (*ClientHandler).handleLIST
	commandsMap["MKD"] = (*ClientHandler).handleMKD
	commandsMap["RMD"] = (*ClientHandler).handleRMD

	// Connection handling
	commandsMap["TYPE"] = (*ClientHandler).handleTYPE
	commandsMap["PASV"] = (*ClientHandler).handlePASV
	commandsMap["EPSV"] = (*ClientHandler).handlePASV
	commandsMap["QUIT"] = (*ClientHandler).handleQUIT

	// Misc
	commandsMap["SYST"] = (*ClientHandler).handleSYST
}

type FtpServer struct {
	Settings         *Settings                 // General settings
	Listener         net.Listener              // Listener used to receive files
	StartTime        time.Time                 // Time when the server was started
	connectionsById  map[uint32]*ClientHandler // Connections map
	connectionsMutex sync.RWMutex              // Connections map sync
	clientCounter    uint32                    // Clients counter
	driver           ServerDriver              // Driver to handle the client authentication and the file access driver selection
}

func NewFtpServer(driver ServerDriver) *FtpServer {
	return &FtpServer{
		driver: driver,
		StartTime: time.Now().UTC(), // Might make sense to put it in Start method
		connectionsById: make(map[uint32]*ClientHandler),
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) clientArrival(c *ClientHandler) error {
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
func (server *FtpServer) clientDeparture(c *ClientHandler) {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	delete(server.connectionsById, c.Id)

	log15.Info("Client disconnected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", len(server.connectionsById))
}

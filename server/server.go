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
	commandsMap["USER"] = (*ClientHandler).HandleUSER
	commandsMap["PASS"] = (*ClientHandler).HandlePASS

	// File access
	commandsMap["STAT"] = (*ClientHandler).HandleSTAT
	commandsMap["SIZE"] = (*ClientHandler).HandleSIZE
	commandsMap["MDTM"] = (*ClientHandler).HandleMDTM
	commandsMap["RETR"] = (*ClientHandler).HandleRETR
	commandsMap["STOR"] = (*ClientHandler).HandleSTOR
	commandsMap["APPE"] = (*ClientHandler).HandleAPPE
	commandsMap["DELE"] = (*ClientHandler).HandleDELE
	commandsMap["RNFR"] = (*ClientHandler).HandleRNFR
	commandsMap["RNTO"] = (*ClientHandler).HandleRNTO

	// Directory handling
	commandsMap["CWD"] = (*ClientHandler).HandleCWD
	commandsMap["PWD"] = (*ClientHandler).HandlePWD
	commandsMap["CDUP"] = (*ClientHandler).HandleCDUP
	commandsMap["NLST"] = (*ClientHandler).HandleLIST
	commandsMap["LIST"] = (*ClientHandler).HandleLIST
	commandsMap["MKD"] = (*ClientHandler).HandleMKD
	commandsMap["RMD"] = (*ClientHandler).HandleRMD

	// Connection handling
	commandsMap["TYPE"] = (*ClientHandler).HandleTYPE
	commandsMap["PASV"] = (*ClientHandler).HandlePASV
	commandsMap["EPSV"] = (*ClientHandler).HandlePASV
	commandsMap["QUIT"] = (*ClientHandler).HandleQUIT

	// Misc
	commandsMap["SYST"] = (*ClientHandler).HandleSYST
}

type FtpServer struct {
	Settings         *Settings                 // General settings
	Listener         net.Listener              // Listener used to receive files
	StartTime        time.Time                 // Time when the server was started
	connectionsById  map[uint32]*ClientHandler // Connections map
	connectionsMutex sync.RWMutex              // Connections map sync
	clientCounter    uint32                    // Clients counter
	driver           Driver                    // Driver to handle all the actual authentication and files access logic
}

func NewFtpServer(driver Driver) *FtpServer {
	return &FtpServer{
		driver: driver,
		StartTime: time.Now().UTC(), // Might make sense to put it in Start method
		connectionsById: make(map[uint32]*ClientHandler),
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) ClientArrival(c *ClientHandler) error {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	nb := len(server.connectionsById)

	log15.Info("Client connected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", nb)

	server.connectionsById[c.Id] = c

	if nb > server.Settings.MaxConnections {
		return errors.New(fmt.Sprintf("Too many clients %d > %d", nb, server.Settings.MaxConnections))
	} else {
		return nil
	}
}

// When a client leaves
func (server *FtpServer) ClientDeparture(c *ClientHandler) {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	log15.Info("Client disconnected", "id", c.Id, "src", c.conn.RemoteAddr(), "total", len(server.connectionsById))

	delete(server.connectionsById, c.Id)
}

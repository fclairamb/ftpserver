package server

import (
	"time"
	"net"
	"sync"
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
	ConnectionsById  map[string]*ClientHandler // Connections map
	PassiveCount     int                       // Number of passive connections opened
	StartTime        int64                     // Time when the server was started
	connectionsMutex sync.RWMutex              // Connections map sync
	driver           Driver                    // Driver to handle all the actual authentication and files access logic
}

func NewFtpServer(driver Driver) *FtpServer {
	return &FtpServer{
		driver: driver,
		StartTime: time.Now().Unix(), // Might make sense to put it in Start method
		ConnectionsById: make(map[string]*ClientHandler),
	}
}

// When a client connects, the server could refuse the connection
func (server *FtpServer) ClientArrival(c *ClientHandler) error {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	server.ConnectionsById[c.Id] = c

	return nil
}

// When a client leaves
func (server *FtpServer) ClientDeparture(c *ClientHandler) {
	server.connectionsMutex.Lock()
	defer server.connectionsMutex.Unlock()

	delete(server.ConnectionsById, c.Id)
}

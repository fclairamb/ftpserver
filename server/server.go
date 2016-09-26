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
	commandsMap["USER"] = (*ClientHandler).HandleUser
	commandsMap["PASS"] = (*ClientHandler).HandlePass

	// File access
	commandsMap["STAT"] = (*ClientHandler).HandleStat
	commandsMap["SIZE"] = (*ClientHandler).HandleSize
	commandsMap["RETR"] = (*ClientHandler).HandleRetr
	commandsMap["STOR"] = (*ClientHandler).HandleStore
	commandsMap["APPE"] = (*ClientHandler).HandleAppend

	// Directory handling
	commandsMap["CWD"] = (*ClientHandler).HandleCwd
	commandsMap["PWD"] = (*ClientHandler).HandlePwd
	commandsMap["CDUP"] = (*ClientHandler).HandleCdUp
	commandsMap["NLST"] = (*ClientHandler).HandleList
	commandsMap["LIST"] = (*ClientHandler).HandleList
	commandsMap["MKD"] = (*ClientHandler).HandleMkd

	// Connection handling
	commandsMap["TYPE"] = (*ClientHandler).HandleType
	commandsMap["PASV"] = (*ClientHandler).HandlePassive
	commandsMap["EPSV"] = (*ClientHandler).HandlePassive
	commandsMap["QUIT"] = (*ClientHandler).HandleQuit

	// Misc
	commandsMap["SYST"] = (*ClientHandler).HandleSyst
}

type FtpServer struct {
	Settings         *Settings                 // General settings
	Listener         net.Listener              // Listener used to receive files
	ConnectionsById  map[string]*ClientHandler // Connections map
	PassiveCount     int                       // Number of passive connections opened
	StartTime        int64                     // Time when the server was started
	connectionsMutex sync.RWMutex                // Connections map sync
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

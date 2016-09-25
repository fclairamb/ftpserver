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

	commandsMap["USER"] = (*ClientHandler).HandleUser
	commandsMap["PASS"] = (*ClientHandler).HandlePass
	commandsMap["STOR"] = (*ClientHandler).HandleStore
	commandsMap["APPE"] = (*ClientHandler).HandleAppend
	commandsMap["STAT"] = (*ClientHandler).HandleStat

	commandsMap["SYST"] = (*ClientHandler).HandleSyst
	commandsMap["PWD"] = (*ClientHandler).HandlePwd
	commandsMap["TYPE"] = (*ClientHandler).HandleType
	commandsMap["PASV"] = (*ClientHandler).HandlePassive
	commandsMap["EPSV"] = (*ClientHandler).HandlePassive
	commandsMap["NLST"] = (*ClientHandler).HandleList
	commandsMap["LIST"] = (*ClientHandler).HandleList
	commandsMap["QUIT"] = (*ClientHandler).HandleQuit
	commandsMap["CWD"] = (*ClientHandler).HandleCwd
	commandsMap["CDUP"] = (*ClientHandler).HandleCdUp
	commandsMap["SIZE"] = (*ClientHandler).HandleSize
	commandsMap["RETR"] = (*ClientHandler).HandleRetr
}

type FtpServer struct {
	Settings        *Settings                 // General settings
	driver          Driver                    // Driver to handle all the actual authentication and files access logic
	Listener        net.Listener              // Listener used to receive files
	ConnectionsById map[string]*ClientHandler // Connections map
	sync            sync.Mutex
	PassiveCount    int                       // Number of passive connections opened
	StartTime       int64                     // Time when the server was started
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
	server.sync.Lock()
	defer server.sync.Unlock()

	server.ConnectionsById[c.Id] = c

	return nil
}

// When a client leaves
func (server *FtpServer) ClientDeparture(c *ClientHandler) {
	server.sync.Lock()
	defer server.sync.Unlock()

	delete(server.ConnectionsById, c.Id)
}

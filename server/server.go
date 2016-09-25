package server

import (
	"time"
	"bufio"
	"net"
	"sync"
	"fmt"
	"strings"
	"io"
	"gopkg.in/inconshreveable/log15.v2"
)

var commandsMap map[string]func(*ClientHandler)

func init() {
	// This is shared between FtpServer instances as there's no point in making the FTP commands behave differently
	// between them.

	commandsMap = make(map[string]func(*ClientHandler))

	commandsMap["USER"] = (*ClientHandler).HandleUser
	commandsMap["PASS"] = (*ClientHandler).HandlePass
	commandsMap["STOR"] = (*ClientHandler).HandleStore
	commandsMap["APPE"] = (*ClientHandler).HandleStore
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
	driver        Driver                    // Driver to handle all the actual authentication and files access logic
	Listener      net.Listener              // Listener used to receive files
	ConnectionMap map[string]*ClientHandler // Connections map
	PassiveCount  int                       // Number of passive connections opened
	StartTime     int64                     // Time when the server was started
}

type ClientHandler struct {
	daddy       *FtpServer          // Server on which the connection was accepted
	writer      *bufio.Writer       // Writer on the TCP connection
	reader      *bufio.Reader       // Reader on the TCP connection
	conn        net.Conn            // TCP connection
	waiter      sync.WaitGroup
	user        string
	homeDir     string
	path        string
	ip          string
	command     string
	param       string
	total       int64
	buffer      []byte
	cid         string
	connectedAt int64
	passives    map[string]*Passive // Index of all the passive connections that are associated to this control connection
	lastPassCid string
	userInfo    map[string]string
	debug       bool                // Show debugging info on the server side
}

func NewFtpServer(driver Driver) *FtpServer {
	return &FtpServer{
		driver: driver,
		StartTime: time.Now().Unix(), // Might make sense to put it in Start method
		ConnectionMap: make(map[string]*ClientHandler),
	}
}

func (server *FtpServer) NewClientHandler(connection net.Conn, cid string, now int64) *ClientHandler {

	p := &ClientHandler{
		daddy: server,
		conn: connection,
		writer: bufio.NewWriter(connection),
		reader: bufio.NewReader(connection),
		connectedAt: time.Now().UTC().UnixNano(),
		path: "/",
		ip: connection.RemoteAddr().String(), // TODO: Do we need this ?
		passives: make(map[string]*Passive),
		userInfo: make(map[string]string),
	}

	// Just respecting the existing logic here, this could be probably be dropped at some point
	p.userInfo["path"] = p.path

	return p
}

func (p *ClientHandler) UserInfo() map[string]string {
	return p.userInfo
}

func (p *ClientHandler) Path() string {
	return p.userInfo["path"]
}

func (p *ClientHandler) SetPath(path string) {
	p.userInfo["path"] = path
}

func (p *ClientHandler) lastPassive() *Passive {
	passive := p.passives[p.lastPassCid]
	if passive == nil {
		return nil
	}
	passive.command = p.command
	passive.param = p.param
	return passive
}

func (p *ClientHandler) HandleCommands() {
	//fmt.Println(p.id, " Got client on: ", p.ip)
	if msg, err := p.daddy.driver.WelcomeUser(p); err == nil {
		p.writeMessage(220, msg)
	} else {
		p.writeMessage(500, msg)
	}

	for {
		line, err := p.reader.ReadString('\n')

		if p.debug {
			log15.Info("FTP RECV", "action", "ftp.cmd_recv", "line", line)
		}

		if err != nil {
			delete(p.daddy.ConnectionMap, p.cid)
			//fmt.Println(p.id, " end ", len(ConnectionMap))
			if err == io.EOF {
				//continue
			}
			break
		}
		command, param := parseLine(line)
		p.command = command
		p.param = param

		fn := commandsMap[command]
		if fn == nil {
			p.writeMessage(550, "not allowed")
		} else {
			fn(p)
		}
	}
}

func (p *ClientHandler) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	if p.debug {
		log15.Info("FTP SEND", "action", "ftp.cmd_send", "line", line)
	}
	p.writer.WriteString(line)
	p.writer.Flush()
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

package server

import "bufio"
import "fmt"
import "io"
import "net"
import "strings"
import "sync"
import (
	"time"
)

var CommandMap map[string]func(*ClientHandler)
var ConnectionMap map[string]*ClientHandler
var PassiveCount int
var UpSince int64

// TODO: Put this in a server handler struct
var driver Driver

type ClientHandler struct {
	writer        *bufio.Writer
	reader        *bufio.Reader
	theConnection net.Conn
	waiter        sync.WaitGroup
	user          string
	homeDir       string
	path          string
	ip            string
	command       string
	param         string
	total         int64
	buffer        []byte
	cid           string
	connectedAt   int64
	passives      map[string]*Passive
	lastPassCid   string
	userInfo      map[string]string
}

func init() {
	UpSince = time.Now().Unix()

	CommandMap = make(map[string]func(*ClientHandler))

	CommandMap["USER"] = (*ClientHandler).HandleUser
	CommandMap["PASS"] = (*ClientHandler).HandlePass
	CommandMap["STOR"] = (*ClientHandler).HandleStore
	CommandMap["APPE"] = (*ClientHandler).HandleStore
	CommandMap["STAT"] = (*ClientHandler).HandleStat

	CommandMap["SYST"] = (*ClientHandler).HandleSyst
	CommandMap["PWD"] = (*ClientHandler).HandlePwd
	CommandMap["TYPE"] = (*ClientHandler).HandleType
	CommandMap["PASV"] = (*ClientHandler).HandlePassive
	CommandMap["EPSV"] = (*ClientHandler).HandlePassive
	CommandMap["NLST"] = (*ClientHandler).HandleList
	CommandMap["LIST"] = (*ClientHandler).HandleList
	CommandMap["QUIT"] = (*ClientHandler).HandleQuit
	CommandMap["CWD"] = (*ClientHandler).HandleCwd
	CommandMap["SIZE"] = (*ClientHandler).HandleSize
	CommandMap["RETR"] = (*ClientHandler).HandleRetr

	ConnectionMap = make(map[string]*ClientHandler)
}

func NewParadise(connection net.Conn, cid string, now int64) *ClientHandler {
	p := ClientHandler{}

	p.writer = bufio.NewWriter(connection)
	p.reader = bufio.NewReader(connection)
	p.path = "/"
	p.theConnection = connection
	p.ip = connection.RemoteAddr().String()
	p.cid = cid
	p.connectedAt = now
	p.passives = make(map[string]*Passive)
	p.userInfo = make(map[string]string)
	p.userInfo["path"] = "/"
	return &p
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
	p.writeMessage(220, "Welcome to Paradise")
	for {
		line, err := p.reader.ReadString('\n')
		if err != nil {
			delete(ConnectionMap, p.cid)
			//fmt.Println(p.id, " end ", len(ConnectionMap))
			if err == io.EOF {
				//continue
			}
			break
		}
		command, param := parseLine(line)
		p.command = command
		p.param = param

		fn := CommandMap[command]
		if fn == nil {
			p.writeMessage(550, "not allowed")
		} else {
			fn(p)
		}
	}
}

func (p *ClientHandler) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
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

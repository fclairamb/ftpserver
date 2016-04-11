package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

var Settings ParadiseSettings
var CommandMap map[string]func(*Paradise)
var ConnectionMap map[string]*Paradise

type Passive struct {
	ListenAt        int64
	ListenSuccessAt int64
	ListenFailedAt  int64
	CloseSuccessAt  int64
	CloseFailedAt   int64
	connectedAt     int64
	passive         *net.TCPConn
	command         string
	cid             string
}

type Paradise struct {
	writer                 *bufio.Writer
	reader                 *bufio.Reader
	theConnection          net.Conn
	passiveConn            *net.TCPConn
	waiter                 sync.WaitGroup
	user                   string
	homeDir                string
	path                   string
	ip                     string
	command                string
	param                  string
	total                  int64
	buffer                 []byte
	cid                    string
	connectedAt            int64
	passives               map[string]*Passive
	passiveListenAt        int64
	passiveListenSuccessAt int64
	passiveListenFailedAt  int64
	passiveCloseSuccessAt  int64
	passiveCloseFailedAt   int64
}

func NewPassive(passive *net.TCPConn, cid string, now int64) *Passive {
	p := Passive{}
	p.cid = cid
	p.connectedAt = now
	p.passive = passive
	return &p
}

func NewParadise(connection net.Conn, cid string, now int64) *Paradise {
	p := Paradise{}

	p.writer = bufio.NewWriter(connection)
	p.reader = bufio.NewReader(connection)
	p.path = "/"
	p.theConnection = connection
	p.ip = connection.RemoteAddr().String()
	p.cid = cid
	p.connectedAt = now
	return &p
}

func (self *Paradise) HandleCommands() {
	//fmt.Println(self.id, " Got client on: ", self.ip)
	self.writeMessage(220, "Welcome to Paradise")
	for {
		line, err := self.reader.ReadString('\n')
		if err != nil {
			delete(ConnectionMap, self.cid)
			//fmt.Println(self.id, " end ", len(ConnectionMap))
			if err == io.EOF {
				//continue
			}
			break
		}
		command, param := parseLine(line)
		self.command = command
		self.param = param

		fn := CommandMap[command]
		if fn == nil {
			self.writeMessage(550, "not allowed")
		} else {
			fn(self)
		}
	}
}

func (self *Paradise) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	self.writer.WriteString(line)
	self.writer.Flush()
}

func (self *Paradise) closePassiveConnection() {
	if self.passiveConn != nil {
		self.passiveConn.Close()
	}
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

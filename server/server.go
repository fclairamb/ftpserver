package server

import (
	"bufio"
	"fmt"
	//	"io"
	"net"
	"strings"
	"sync"
)

type ConnectionHolder struct {
}

type Paradise struct {
	writer        *bufio.Writer
	reader        *bufio.Reader
	TheConnection net.Conn
	passiveConn   *net.TCPConn
	Waiter        sync.WaitGroup
	User          string
	HomeDir       string
	Path          string
	Ip            string
	Command       string
	Param         string
}

/*
func HandleCommands(holder *ConnectionHolder) {
	p := NewParadise(holder)

	p.writeMessage(220, "Welcome to Paradise")
	for {
		line, err := p.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}
			break
		}
		holder.Command, holder.Param = parseLine(line)

		if command == "USER" {
			p.handleUser()
		} else if command == "PASS" {
			p.handlePass()
		} else {
			p.writeMessage(550, "not allowed")
		}

		// close passive connection each time
		p.ClosePassiveConnection()
	}
}
*/

func NewParadise(connection net.Conn) *Paradise {
	p := Paradise{}

	p.writer = bufio.NewWriter(connection)
	p.reader = bufio.NewReader(connection)
	return &p
}

func (self *Paradise) HandleCommands() {
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

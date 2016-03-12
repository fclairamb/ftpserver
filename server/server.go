package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type ConnectionHolder struct {
}

type Paradise struct {
	writer        *bufio.Writer
	reader        *bufio.Reader
	theConnection net.Conn
	passiveConn   *net.TCPConn
	Waiter        sync.WaitGroup
	user          string
	HomeDir       string
	path          string
	ip            string
	command       string
	param         string
}

func NewParadise(connection net.Conn) *Paradise {
	p := Paradise{}

	p.writer = bufio.NewWriter(connection)
	p.reader = bufio.NewReader(connection)
	p.path = "/"
	p.theConnection = connection
	p.ip = connection.RemoteAddr().String()
	return &p
}

func (self *Paradise) HandleCommands() {
	fmt.Println("Got client on: ", self.ip)
	self.writeMessage(220, "Welcome to Paradise")
	for {
		line, err := self.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}
			break
		}
		command, param := parseLine(line)
		self.command = command
		self.param = param

		if command == "USER" {
			self.handleUser()
		} else if command == "PASS" {
			self.handlePass()
		} else if command == "SYST" {
			self.handleSyst()
		} else if command == "PWD" {
			self.handlePwd()
		} else if command == "TYPE" {
		} else if command == "EPSV" || command == "PASV" {
		} else if command == "LIST" || command == "NLST" {
		} else if command == "QUIT" {
		} else if command == "CWD" {
		} else if command == "SIZE" {
		} else if command == "RETR" {
		} else if command == "STAT" {
		} else if command == "STOR" || command == "APPE" {
		} else {
			self.writeMessage(550, "not allowed")
		}

		// close passive connection each time
		self.closePassiveConnection()
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

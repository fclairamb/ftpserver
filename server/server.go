package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
)

var CommandMap map[string]string

type Paradise struct {
	writer        *bufio.Writer
	reader        *bufio.Reader
	theConnection net.Conn
	passiveConn   *net.TCPConn
	waiter        sync.WaitGroup
	user          string
	homeDir       string
	path          string
	ip            string
	command       string
	param         string
	total         int64
	buffer        []byte
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

		//var t T
		method_name := CommandMap[command]
		thing := reflect.ValueOf(self).MethodByName(method_name)
		thing_s := fmt.Sprintf("%s", thing)
		if thing_s == "<invalid reflect.Value>" {
			self.writeMessage(550, "not allowed")
		} else {
			thing.Call([]reflect.Value{})
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

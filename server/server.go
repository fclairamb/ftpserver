package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type ConnectionHolder struct {
	TheConnection net.Conn
	PassiveConn   *net.TCPConn
	Waiter        sync.WaitGroup
	User          string
	HomeDir       string
	Path          string
}

type Paradise struct {
	writer *bufio.Writer
	reader *bufio.Reader
	holder *ConnectionHolder
}

func HandleCommands(holder *ConnectionHolder) {
	p := NewParadise(holder)

	p.writeMessage(220, "Welcome to Paradise")
}

func NewParadise(holder *ConnectionHolder) *Paradise {
	p := Paradise{}

	p.writer = bufio.NewWriter(holder.TheConnection)
	p.reader = bufio.NewReader(holder.TheConnection)
	p.holder = holder
	return &p
}

func (self *Paradise) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	self.writer.WriteString(line)
	self.writer.Flush()
}

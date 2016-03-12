package server

import (
	//"bufio"
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

type ParadiseWriter struct {
}

func HandleCommands(holder *ConnectionHolder) {
	//var cw *bufio.Writer
	//var cr *bufio.Reader
	//cw = bufio.NewWriter(holder.TheConnection)
	//cr = bufio.NewReader(holder.TheConnection)

	//writeMessage(220, "Welcome to Paradise", cw)
}

package server

import (
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

func HandleCommands(holder *ConnectionHolder) {
}

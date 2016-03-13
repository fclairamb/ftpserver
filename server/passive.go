package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (self *Paradise) HandlePassive() {
	fmt.Println(self.ip, self.command, self.param)

	laddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	passiveListen, _ := net.ListenTCP("tcp", laddr)

	add := passiveListen.Addr()
	parts := strings.Split(add.String(), ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])

	self.waiter.Add(1)

	go func() {
		self.passiveConn, _ = passiveListen.AcceptTCP()
		passiveListen.Close()
		self.waiter.Done()
	}()

	if self.command == "PASV" {
		p1 := port / 256
		p2 := port - (p1 * 256)
		addr := self.theConnection.LocalAddr()
		tokens := strings.Split(addr.String(), ":")
		host := tokens[0]
		quads := strings.Split(host, ".")
		target := fmt.Sprintf("(%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2)
		msg := "Entering Passive Mode " + target
		self.writeMessage(227, msg)
	} else {
		msg := fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", port)
		self.writeMessage(229, msg)
	}
}

package server

import "fmt"
import "sync"
import "net"
import "strconv"
import "strings"
import "time"

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func getThatPassiveConnection(passiveListen *net.TCPListener, p *Paradise) {
	var perr error
	p.passiveConn, perr = passiveListen.AcceptTCP()
	if perr != nil {
		p.passiveListenFailedAt = time.Now().Unix()
		p.waiter.Done()
		return
	}
	passiveListen.Close()
	p.passiveListenSuccessAt = time.Now().Unix()
	p.waiter.Done()
}

func (self *Paradise) HandlePassive() {
	//fmt.Println(self.ip, self.command, self.param)

	laddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	passiveListen, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		self.writeMessage(550, "Error with passive: "+err.Error())
		return
	}

	add := passiveListen.Addr()
	parts := strings.Split(add.String(), ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])

	self.waiter.Add(1)
	self.passiveListenFailedAt = 0
	self.passiveListenSuccessAt = 0
	self.passiveListenAt = time.Now().Unix()
	go getThatPassiveConnection(passiveListen, self)

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

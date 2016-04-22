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

type Passive struct {
	listenSuccessAt int64
	listenFailedAt  int64
	closeSuccessAt  int64
	closeFailedAt   int64
	listenAt        int64
	connectedAt     int64
	connection      *net.TCPConn
	command         string
	cid             string
	port            int
	waiter          sync.WaitGroup
}

func getThatPassiveConnection(passiveListen *net.TCPListener, p *Passive) {
	var perr error
	p.connection, perr = passiveListen.AcceptTCP()
	if perr != nil {
		p.listenFailedAt = time.Now().Unix()
		p.waiter.Done()
		return
	}
	passiveListen.Close()
	// start reading from p.passive, it will block, wait for err. Err means client killed connection.
	p.listenSuccessAt = time.Now().Unix()
	p.waiter.Done()
}

func NewPassive(passiveListen *net.TCPListener, cid string, now int64) *Passive {
	p := Passive{}
	p.cid = cid
	p.listenAt = now

	add := passiveListen.Addr()
	parts := strings.Split(add.String(), ":")
	p.port, _ = strconv.Atoi(parts[len(parts)-1])

	p.waiter.Add(1)
	p.listenFailedAt = 0
	p.listenSuccessAt = 0
	p.listenAt = time.Now().Unix()
	go getThatPassiveConnection(passiveListen, &p)

	return &p
}

func anotherPassiveIsAvail() bool {
	return false
}

func (self *Paradise) HandlePassive() {
	//fmt.Println(self.ip, self.command, self.param)

	laddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	passiveListen, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		self.writeMessage(550, "Error with passive: "+err.Error())
		return
	}
	if anotherPassiveIsAvail() {
		self.writeMessage(550, "Use other passive connection first.")
		return
	}

	cid := genClientID()
	p := NewPassive(passiveListen, cid, time.Now().Unix())
	self.lastPassCid = cid
	self.passives[cid] = p

	if self.command == "PASV" {
		p1 := p.port / 256
		p2 := p.port - (p1 * 256)
		addr := self.theConnection.LocalAddr()
		tokens := strings.Split(addr.String(), ":")
		host := tokens[0]
		quads := strings.Split(host, ".")
		target := fmt.Sprintf("(%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2)
		msg := "Entering Passive Mode " + target
		self.writeMessage(227, msg)
	} else {
		msg := fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", p.port)
		self.writeMessage(229, msg)
	}
}

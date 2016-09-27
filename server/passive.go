package server

import (
	"time"
	"sync"
	"net"
	"strings"
	"strconv"
	"fmt"
)

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
	param           string
	cid             string
	port            int
	waiter          sync.WaitGroup
}

func (c *ClientHandler) closePassive(passive *Passive) {
	if passive.connection != nil {
		err := passive.connection.Close()
		if err != nil {
			passive.closeFailedAt = time.Now().Unix()
		} else {
			passive.closeSuccessAt = time.Now().Unix()

		}
	}

	delete(c.passives, passive.cid)
	c.daddy.PassiveCount--
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

func (c *ClientHandler) NewPassive(passiveListen *net.TCPListener, cid string, now int64) *Passive {
	c.daddy.PassiveCount++
	p := Passive{}
	p.cid = cid
	p.listenAt = now

	add := passiveListen.Addr()
	parts := strings.Split(add.String(), ":")
	p.port, _ = strconv.Atoi(parts[len(parts) - 1])

	p.waiter.Add(1)
	p.listenFailedAt = 0
	p.listenSuccessAt = 0
	p.listenAt = time.Now().Unix()
	go getThatPassiveConnection(passiveListen, &p)

	return &p
}

func (c *ClientHandler) HandlePassive() {
	laddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	passiveListen, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		c.writeMessage(550, "Error with passive: " + err.Error())
		return
	}

	cid := genClientID()
	passive := c.NewPassive(passiveListen, cid, time.Now().Unix())
	passive.command = c.command
	passive.param = c.param
	c.lastPassCid = cid
	c.passives[cid] = passive

	if c.command == "PASV" {
		p1 := passive.port / 256
		p2 := passive.port - (p1 * 256)
		addr := c.conn.LocalAddr() // <-- I don't think this will be enough
		tokens := strings.Split(addr.String(), ":")
		host := tokens[0]
		quads := strings.Split(host, ".")
		c.writeMessage(227, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2))
	} else {
		c.writeMessage(229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", passive.port))
	}
}

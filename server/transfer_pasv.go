package server

import (
	"time"
	"net"
	"strings"
	"fmt"
	"gopkg.in/inconshreveable/log15.v2"
	"crypto/tls"
)

// Active/Passive transfer connection handler
type transferHandler interface {
	// Get the connection to transfer data on
	Open() (net.Conn, error)

	// Close the connection (and any associated resource)
	Close() error
}

// Passive connection
type passiveTransferHandler struct {
	listener    net.Listener     // TCP or SSL Listener
	tcpListener *net.TCPListener // TCP Listener (only keeping it to define a deadline during the accept)
	Port        int              // TCP Port we are listening on
	connection  net.Conn         // TCP Connection established
}

func (c *clientHandler) handlePASV() {
	addr, _ := net.ResolveTCPAddr("tcp", ":0")
	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log15.Error("Could not listen", "err", err)
		return
	}

	// The listener will either be plain TCP or TLS
	var listener net.Listener
	if c.transferTls {
		if tlsConfig, err := c.daddy.driver.GetTLSConfig(); err == nil {
			listener = tls.NewListener(tcpListener, tlsConfig)
		} else {
			c.writeMessage(550, fmt.Sprintf("Cannot get a TLS config: %v", err))
			return
		}
	} else {
		listener = tcpListener
	}

	p := &passiveTransferHandler{
		tcpListener: tcpListener,
		listener: listener,
		Port: tcpListener.Addr().(*net.TCPAddr).Port,
	}

	// We should rewrite this part
	if c.command == "PASV" {
		p1 := p.Port / 256
		p2 := p.Port - (p1 * 256)
		addr := c.conn.LocalAddr()
		tokens := strings.Split(addr.String(), ":")
		host := tokens[0]
		quads := strings.Split(host, ".")
		c.writeMessage(227, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2))
	} else {
		c.writeMessage(229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", p.Port))
	}

	c.transfer = p
}

func (p *passiveTransferHandler) ConnectionWait(wait time.Duration) (net.Conn, error) {
	if p.connection == nil {
		p.tcpListener.SetDeadline(time.Now().Add(wait))
		var err error
		if p.connection, err = p.listener.Accept(); err == nil {
			return p.connection, nil
		} else {
			return nil, err
		}
	}

	return p.connection, nil
}

func (p *passiveTransferHandler) Open() (net.Conn, error) {
	return p.ConnectionWait(time.Minute)
}

// Closing only the client connection is not supported at that time
func (p *passiveTransferHandler) Close() error {
	if p.tcpListener != nil {
		p.tcpListener.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
	return nil
}

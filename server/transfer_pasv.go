// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/fclairamb/ftpserver/server/log"
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
	settings    *Settings        // Settings
	logger      log.Logger       // Logger
}

func (c *clientHandler) getCurrentIP() ([]string, error) {
	// Provide our external IP address so the ftp client can connect back to us
	ip := c.server.settings.PublicHost

	// If we don't have an IP address, we can take the one that was used for the current connection
	if ip == "" {
		// Defer to the user provided resolver.
		if c.server.settings.PublicIPResolver != nil {
			var err error
			ip, err = c.server.settings.PublicIPResolver(c)

			if err != nil {
				return nil, fmt.Errorf("couldn't fetch public IP: %v", err)
			}
		} else {
			ip = strings.Split(c.conn.LocalAddr().String(), ":")[0]
		}
	}

	return strings.Split(ip, "."), nil
}

func (c *clientHandler) handlePASV() error {
	addr, _ := net.ResolveTCPAddr("tcp", ":0")

	var tcpListener *net.TCPListener

	var err error

	portRange := c.server.settings.PassiveTransferPortRange

	if portRange != nil {
		for start := portRange.Start; start < portRange.End; start++ {
			port := portRange.Start + rand.Intn(portRange.End-portRange.Start)
			laddr, err2 := net.ResolveTCPAddr("tcp", "0.0.0.0:"+fmt.Sprintf("%d", port))

			if err2 != nil {
				continue
			}

			tcpListener, err = net.ListenTCP("tcp", laddr)
			if err == nil {
				break
			}
		}
	} else {
		tcpListener, err = net.ListenTCP("tcp", addr)
	}

	if err != nil {
		c.logger.Error(logKeyMsg, "Could not listen", "err", err)
		return nil
	}

	// The listener will either be plain TCP or TLS
	var listener net.Listener

	if c.transferTLS {
		if tlsConfig, err := c.server.driver.GetTLSConfig(); err == nil {
			listener = tls.NewListener(tcpListener, tlsConfig)
		} else {
			c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Cannot get a TLS config: %v", err))
			return nil
		}
	} else {
		listener = tcpListener
	}

	p := &passiveTransferHandler{
		tcpListener: tcpListener,
		listener:    listener,
		Port:        tcpListener.Addr().(*net.TCPAddr).Port,
		settings:    c.server.settings,
		logger:      c.logger,
	}

	// We should rewrite this part
	if c.command == "PASV" {
		p1 := p.Port / 256
		p2 := p.Port - (p1 * 256)
		quads, err2 := c.getCurrentIP()

		if err2 != nil {
			return err2
		}

		c.writeMessage(
			StatusEnteringPASV,
			fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2))
	} else {
		c.writeMessage(StatusEnteringEPSV, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", p.Port))
	}

	c.transfer = p

	return nil
}

func (p *passiveTransferHandler) ConnectionWait(wait time.Duration) (net.Conn, error) {
	if p.connection == nil {
		var err error
		if err = p.tcpListener.SetDeadline(time.Now().Add(wait)); err != nil {
			return nil, fmt.Errorf("failed to set deadline: %v", err)
		}

		p.connection, err = p.listener.Accept()

		if err != nil {
			return nil, err
		}
	}

	return p.connection, nil
}

func (p *passiveTransferHandler) Open() (net.Conn, error) {
	timeout := time.Duration(time.Second.Nanoseconds() * int64(p.settings.ConnectionTimeout))
	return p.ConnectionWait(timeout)
}

// Closing only the client connection is not supported at that time
func (p *passiveTransferHandler) Close() error {
	if p.tcpListener != nil {
		if err := p.tcpListener.Close(); err != nil {
			p.logger.Warn(
				"msg", "Problem closing passive listener",
				"action", "err.closing_passive_listener",
				"err", err,
			)
		}
	}

	if p.connection != nil {
		if err := p.connection.Close(); err != nil {
			p.logger.Warn(
				"msg", "Problem closing passive connecting",
				"action", "err.closing_passive_connection",
				"err", err,
			)
		}
	}

	return nil
}

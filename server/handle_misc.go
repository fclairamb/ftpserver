package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
)

func (c *clientHandler) handleAUTH() {
	if tlsConfig, err := c.daddy.driver.GetTLSConfig(); err == nil {
		c.writeMessage(234, "AUTH command ok. Expecting TLS Negotiation.")
		c.conn = tls.Server(c.conn, tlsConfig)
		c.reader = bufio.NewReader(c.conn)
		c.writer = bufio.NewWriter(c.conn)
	} else {
		c.writeMessage(550, fmt.Sprintf("Cannot get a TLS config: %v", err))
	}
}

func (c *clientHandler) handlePROT() {
	// P for Private, C for Clear
	c.transferTLS = c.param == "P"
	c.writeMessage(200, "OK")
}

func (c *clientHandler) handlePBSZ() {
	c.writeMessage(200, "Whatever")
}

func (c *clientHandler) handleSYST() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *clientHandler) handleSTAT() {
	// STAT is a bit tricky

	if c.param == "" { // Without a file, it's the server stat
		c.handleSTATServer()
	} else { // With a file/dir it's the file or the dir's files stat
		c.handleSTATFile()
	}
}

func (c *clientHandler) handleSITE() {
	spl := strings.SplitN(c.param, " ", 2)
	if len(spl) > 1 {
		if strings.ToUpper(spl[0]) == "CHMOD" {
			c.handleCHMOD(spl[1])
			return
		}
	}
	c.writeMessage(500, "Not understood SITE subcommand")
}

func (c *clientHandler) handleSTATServer() {
	c.writeLine("213- FTP server status:")
	duration := time.Now().UTC().Sub(c.connectedAt)
	duration -= duration % time.Second
	c.writeLine(fmt.Sprintf(
		"Connected to %s from %s for %s",
		c.daddy.settings.ListenAddr,
		c.conn.RemoteAddr(),
		duration,
	))
	if c.user != "" {
		c.writeLine(fmt.Sprintf("Logged in as %s", c.user))
	} else {
		c.writeLine("Not logged in yet")
	}
	c.writeLine("ftpserver - golang FTP server")
	defer c.writeMessage(213, "End")
}

func (c *clientHandler) handleOPTS() {
	args := strings.SplitN(c.param, " ", 2)
	if strings.ToUpper(args[0]) == "UTF8" {
		c.writeMessage(200, "I'm in UTF8 only anyway")
	} else {
		c.writeMessage(500, "Don't know this option")
	}
}

func (c *clientHandler) handleNOOP() {
	c.writeMessage(200, "OK")
}

func (c *clientHandler) handleFEAT() {
	c.writeLine("211- These are my features")
	defer c.writeMessage(211, "end")

	features := []string{
		"UTF8",
		"SIZE",
		"MDTM",
		"REST STREAM",
	}

	if !c.daddy.settings.DisableMLSD {
		features = append(features, "MLSD")
	}

	for _, f := range features {
		c.writeLine(" " + f)
	}
}

func (c *clientHandler) handleTYPE() {
	switch c.param {
	case "I":
		c.writeMessage(200, "Type set to binary")
	case "A":
		c.writeMessage(200, "WARNING: ASCII isn't correctly supported")
	default:
		c.writeMessage(500, "Not understood")
	}
}

func (c *clientHandler) handleQUIT() {
	c.writeMessage(221, "Goodbye")
	c.disconnect()
	c.reader = nil
}

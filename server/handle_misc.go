package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"strings"
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
	c.transferTls = (c.param == "P")
	c.writeMessage(200, "OK")
}

func (c *clientHandler) handlePBSZ() {
	c.writeMessage(200, "Whatever")
}

func (c *clientHandler) handleSYST() {
	c.writeMessage(215, "UNIX Type: L8")
}

func (c *clientHandler) handleOPTS() {
	args := strings.SplitN(c.param, " ", 2)
	if args[0] == "UTF8" {
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

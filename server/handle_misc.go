package server

import (
	"crypto/tls"
	"bufio"
	"fmt"
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

func (c *clientHandler) handleFEAT() {
	c.writer.WriteString("211- These are my features\r\n")

	c.writer.WriteString(" UTF8\r\n")

	c.writeMessage(211, "end")
}

func (c *clientHandler) handleTYPE() {
	c.writeMessage(200, "Type set to binary")
}

func (c *clientHandler) handleQUIT() {
	//fmt.Println("Goodbye")
	c.writeMessage(221, "Goodbye")
	c.disconnect()
}

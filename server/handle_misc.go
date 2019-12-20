package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
)

func (c *clientHandler) handleAUTH() error {
	if tlsConfig, err := c.server.driver.GetTLSConfig(); err == nil {
		c.writeMessage(StatusAuthAccepted, "AUTH command ok. Expecting TLS Negotiation.")
		c.conn = tls.Server(c.conn, tlsConfig)
		c.reader = bufio.NewReader(c.conn)
		c.writer = bufio.NewWriter(c.conn)
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Cannot get a TLS config: %v", err))
	}
	return nil
}

func (c *clientHandler) handlePROT() error {
	// P for Private, C for Clear
	c.transferTLS = c.param == "P"
	c.writeMessage(StatusOK, "OK")
	return nil
}

func (c *clientHandler) handlePBSZ() error {
	c.writeMessage(StatusOK, "Whatever")
	return nil
}

func (c *clientHandler) handleSYST() error {
	c.writeMessage(StatusSystemType, "UNIX Type: L8")
	return nil
}

func (c *clientHandler) handleSTAT() error {
	// STAT is a bit tricky

	if c.param == "" { // Without a file, it's the server stat
		return c.handleSTATServer()
	} else { // With a file/dir it's the file or the dir's files stat
		return c.handleSTATFile()
	}
}

func (c *clientHandler) handleSITE() error {
	spl := strings.SplitN(c.param, " ", 2)
	if len(spl) > 1 {
		if strings.ToUpper(spl[0]) == "CHMOD" {
			c.handleCHMOD(spl[1])
			return nil
		}
	}
	c.writeMessage(StatusSyntaxErrorNotRecognised, "Not understood SITE subcommand")
	return nil
}

func (c *clientHandler) handleSTATServer() error {
	//c.writeLine(fmt.Sprintf("%d-FTP server status:", StatusFileStatus))

	m := c.multilineAnswer(StatusFileStatus, "Server status")
	defer m()
	duration := time.Now().UTC().Sub(c.connectedAt)
	duration -= duration % time.Second
	c.writeLine(fmt.Sprintf(
		"Connected to %s from %s for %s",
		c.server.settings.ListenAddr,
		c.conn.RemoteAddr(),
		duration,
	))
	if c.user != "" {
		c.writeLine(fmt.Sprintf("Logged in as %s", c.user))
	} else {
		c.writeLine("Not logged in yet")
	}
	c.writeLine("ftpserver - golang FTP server")
	return nil
	// defer c.writeMessage(StatusFileStatus, "End")
}

func (c *clientHandler) handleOPTS() error {
	args := strings.SplitN(c.param, " ", 2)
	if strings.ToUpper(args[0]) == "UTF8" {
		c.writeMessage(StatusOK, "I'm in UTF8 only anyway")
	} else {
		c.writeMessage(StatusSyntaxErrorNotRecognised, "Don't know this option")
	}
	return nil
}

func (c *clientHandler) handleNOOP() error {
	c.writeMessage(StatusOK, "OK")
	return nil
}

func (c *clientHandler) handleFEAT() error {
	c.writeLine(fmt.Sprintf("%d- These are my features", StatusSystemStatus))
	defer c.writeMessage(StatusSystemStatus, "end")

	features := []string{
		"UTF8",
		"SIZE",
		"MDTM",
		"REST STREAM",
	}

	if !c.server.settings.DisableMLSD {
		features = append(features, "MLSD")
	}

	if !c.server.settings.DisableMLST {
		features = append(features, "MLST")
	}

	for _, f := range features {
		c.writeLine(" " + f)
	}
	return nil
}

func (c *clientHandler) handleTYPE() error {
	switch c.param {
	case "I":
		c.writeMessage(StatusOK, "Type set to binary")
	case "A":
		c.writeMessage(StatusOK, "WARNING: ASCII isn't correctly supported")
	default:
		c.writeMessage(StatusSyntaxErrorNotRecognised, "Not understood")
	}
	return nil
}

func (c *clientHandler) handleQUIT() error {
	c.writeMessage(StatusClosingControlConn, "Goodbye")
	c.disconnect()
	c.reader = nil
	return nil
}

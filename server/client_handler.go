// Package server is the core of the library
package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/fclairamb/ftpserver/server/log"
)

// nolint: maligned
type clientHandler struct {
	id          uint32               // ID of the client
	server      *FtpServer           // Server on which the connection was accepted
	driver      ClientHandlingDriver // Client handling driver
	conn        net.Conn             // TCP connection
	writer      *bufio.Writer        // Writer on the TCP connection
	reader      *bufio.Reader        // Reader on the TCP connection
	user        string               // Authenticated user
	path        string               // Current path
	clnt        string               // Identified client
	command     string               // Command received on the connection
	param       string               // Param of the FTP command
	connectedAt time.Time            // Date of connection
	ctxRnfr     string               // Rename from
	ctxRest     int64                // Restart point
	debug       bool                 // Show debugging info on the server side
	transfer    transferHandler      // Transfer connection (only passive is implemented at this stage)
	transferTLS bool                 // Use TLS for transfer connection
	logger      log.Logger           // Client handler logging
}

// newClientHandler initializes a client handler when someone connects
func (server *FtpServer) newClientHandler(connection net.Conn, id uint32) *clientHandler {
	p := &clientHandler{
		server:      server,
		conn:        connection,
		id:          id,
		writer:      bufio.NewWriter(connection),
		reader:      bufio.NewReader(connection),
		connectedAt: time.Now().UTC(),
		path:        "/",
		logger:      server.Logger.With("clientId", id),
	}

	// Just respecting the existing logic here, this could be probably be dropped at some point

	return p
}

func (c *clientHandler) disconnect() {
	if err := c.conn.Close(); err != nil {
		c.logger.Warn(
			"msg", "Problem disconnecting a client",
			"action", "ftp.err_disconnecting",
			"err", err)
	}
}

// Path provides the current working directory of the client
func (c *clientHandler) Path() string {
	return c.path
}

// SetPath changes the current working directory
func (c *clientHandler) SetPath(path string) {
	c.path = path
}

// Debug defines if we will list all interaction
func (c *clientHandler) Debug() bool {
	return c.debug
}

// SetDebug changes the debug flag
func (c *clientHandler) SetDebug(debug bool) {
	c.debug = debug
}

// ID provides the client's ID
func (c *clientHandler) ID() uint32 {
	return c.id
}

// RemoteAddr returns the remote network address.
func (c *clientHandler) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr returns the local network address.
func (c *clientHandler) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *clientHandler) end() {
	c.server.driver.UserLeft(c)
	c.server.clientDeparture(c)

	if c.transfer != nil {
		if err := c.transfer.Close(); err != nil {
			c.logger.Warn(
				"msg", "Problem closing a transfer",
				"action", "ftp.err_closing_transfer",
				"err", err,
			)
		}
	}
}

// HandleCommands reads the stream of commands
func (c *clientHandler) HandleCommands() {
	defer c.end()

	if msg, err := c.server.driver.WelcomeUser(c); err == nil {
		c.writeMessage(StatusServiceReady, msg)
	} else {
		c.writeMessage(StatusSyntaxErrorNotRecognised, msg)
		return
	}

	for {
		if c.reader == nil {
			if c.debug {
				c.logger.Debug(logKeyMsg, "Clean disconnect", logKeyAction, "ftp.disconnect", "clean", true)
			}

			return
		}

		// florent(2018-01-14): #58: IDLE timeout: Preparing the deadline before we read
		if c.server.settings.IdleTimeout > 0 {
			if err := c.conn.SetDeadline(
				time.Now().Add(time.Duration(time.Second.Nanoseconds() * int64(c.server.settings.IdleTimeout)))); err != nil {
				c.logger.Error(logKeyMsg, "Network error", logKeyAction, "tcp.set_deadline", "err", err)
			}
		}

		line, err := c.reader.ReadString('\n')

		if err != nil {
			c.handleCommandsStreamError(err)
			return
		}

		if c.debug {
			c.logger.Debug(logKeyMsg, "FTP RECV", logKeyAction, "ftp.cmd_recv", "line", line)
		}

		c.handleCommand(line)
	}
}

func (c *clientHandler) handleCommandsStreamError(err error) {
	// florent(2018-01-14): #58: IDLE timeout: Adding some code to deal with the deadline
	switch err := err.(type) {
	case net.Error:
		if err.Timeout() {
			// We have to extend the deadline now
			if err := c.conn.SetDeadline(time.Now().Add(time.Minute)); err != nil {
				c.logger.Error(logKeyMsg, "Could not set deadline", logKeyAction, "ftp.deadline_fail", "err", err)
			}

			c.logger.Info(logKeyMsg, "IDLE timeout", logKeyAction, "ftp.idle_timeout", "err", err)
			c.writeMessage(
				StatusServiceNotAvailable,
				fmt.Sprintf("command timeout (%d seconds): closing control connection", c.server.settings.IdleTimeout))

			if err := c.writer.Flush(); err != nil {
				c.logger.Error(logKeyMsg, "Network flush error", logKeyAction, "ftp.flush_error", "err", err)
			}

			if err := c.conn.Close(); err != nil {
				c.logger.Error(logKeyMsg, "Network close error", logKeyAction, "ftp.close_error", "err", err)
			}

			break
		}

		c.logger.Error(logKeyMsg, "Network error", logKeyAction, "ftp.net_error", "err", err)
	default:
		if err == io.EOF {
			if c.debug {
				c.logger.Debug(logKeyMsg, "TCP disconnect", logKeyAction, "ftp.disconnect", "clean", false)
			}
		} else {
			c.logger.Error(logKeyMsg, "Read error", logKeyAction, "ftp.read_error", "err", err)
		}
	}
}

// handleCommand takes care of executing the received line
func (c *clientHandler) handleCommand(line string) {
	command, param := parseLine(line)
	c.command = strings.ToUpper(command)
	c.param = param

	cmdDesc := commandsMap[c.command]
	if cmdDesc == nil {
		c.writeMessage(StatusSyntaxErrorNotRecognised, "Unknown command")
		return
	}

	if c.driver == nil && !cmdDesc.Open {
		c.writeMessage(StatusNotLoggedIn, "Please login with USER and PASS")
		return
	}

	// Let's prepare to recover in case there's a command error
	defer func() {
		if r := recover(); r != nil {
			c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Unhandled internal error: %s", r))
		}
	}()

	if err := cmdDesc.Fn(c); err != nil {
		c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Error: %s", err))
	}
}

func (c *clientHandler) writeLine(line string) {
	if c.debug {
		c.logger.Debug(logKeyMsg, "FTP SEND", logKeyAction, "ftp.cmd_send", "line", line)
	}

	if _, err := c.writer.WriteString(fmt.Sprintf("%s\r\n", line)); err != nil {
		c.logger.Warn(
			logKeyMsg, "Message could not be sent",
			logKeyAction, "err.cmd_send",
			"line", line,
			"err", err,
		)
	}

	if err := c.writer.Flush(); err != nil {
		c.logger.Warn(
			logKeyMsg, "Couldn't flush line",
			logKeyAction, "err.client_flush",
			"err", err,
		)
	}
}

func (c *clientHandler) writeMessage(code int, message string) {
	c.writeLine(fmt.Sprintf("%d %s", code, message))
}

func (c *clientHandler) TransferOpen() (net.Conn, error) {
	if c.transfer == nil {
		c.writeMessage(StatusActionNotTaken, "No passive connection declared")
		return nil, errors.New("no passive connection declared")
	}

	c.writeMessage(StatusFileStatusOK, "Using transfer connection")
	conn, err := c.transfer.Open()

	if err == nil && c.debug {
		c.logger.Debug(
			logKeyMsg, "FTP Transfer connection opened",
			logKeyAction, "ftp.transfer_open",
			"remoteAddr", conn.RemoteAddr().String(),
			"localAddr", conn.LocalAddr().String())
	}

	return conn, err
}

func (c *clientHandler) TransferClose() {
	if c.transfer != nil {
		c.writeMessage(StatusClosingDataConn, "Closing transfer connection")

		if err := c.transfer.Close(); err != nil {
			c.logger.Warn(
				logKeyMsg, "Problem closing tranfer connection",
				logKeyAction, "err.closing_transfer",
				"err", err,
			)
		}

		c.transfer = nil

		if c.debug {
			c.logger.Debug(logKeyMsg, "FTP Transfer connection closed", logKeyAction, "ftp.transfer_close")
		}
	}
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}

	return params[0], params[1]
}

// For future use
func (c *clientHandler) multilineAnswer(code int, message string) func() {
	c.writeLine(fmt.Sprintf("%d-%s", code, message))

	return func() {
		c.writeLine(fmt.Sprintf("%d End", code))
	}
}

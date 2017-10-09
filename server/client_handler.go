package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type clientHandler struct {
	ID          uint32               // ID of the client
	daddy       *FtpServer           // Server on which the connection was accepted
	driver      ClientHandlingDriver // Client handling driver
	conn        net.Conn             // TCP connection
	writer      *bufio.Writer        // Writer on the TCP connection
	reader      *bufio.Reader        // Reader on the TCP connection
	user        string               // Authenticated user
	path        string               // Current path
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
func (server *FtpServer) newClientHandler(connection net.Conn) *clientHandler {

	id := server.clientCounter

	server.clientCounter++

	p := &clientHandler{
		daddy:       server,
		conn:        connection,
		ID:          id,
		writer:      bufio.NewWriter(connection),
		reader:      bufio.NewReader(connection),
		connectedAt: time.Now().UTC(),
		path:        "/",
		logger:      log.With(server.Logger, "clientId", id),
	}

	// Just respecting the existing logic here, this could be probably be dropped at some point

	return p
}

func (c *clientHandler) disconnect() {
	c.conn.Close()
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

func (c *clientHandler) end() {
	if c.transfer != nil {
		c.transfer.Close()
	}
}

// HandleCommands reads the stream of commands
func (c *clientHandler) HandleCommands() {
	defer c.daddy.clientDeparture(c)
	defer c.end()

	if err := c.daddy.clientArrival(c); err != nil {
		c.writeMessage(500, "Can't accept you - "+err.Error())
		return
	}

	defer c.daddy.driver.UserLeft(c)

	//fmt.Println(c.id, " Got client on: ", c.ip)
	if msg, err := c.daddy.driver.WelcomeUser(c); err == nil {
		c.writeMessage(220, msg)
	} else {
		c.writeMessage(500, msg)
		return
	}

	for {
		if c.reader == nil {
			if c.debug {
				level.Debug(c.logger).Log(logKeyMsg, "Clean disconnect", logKeyAction, "ftp.disconnect", "clean", true)
			}
			return
		}

		line, err := c.reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if c.debug {
					level.Debug(c.logger).Log(logKeyMsg, "TCP disconnect", logKeyAction, "ftp.disconnect", "clean", false)
				}
			} else {
				level.Error(c.logger).Log(logKeyMsg, "Read error", logKeyAction, "ftp.read_error", "err", err)
			}
			return
		}

		if c.debug {
			level.Debug(c.logger).Log(logKeyMsg, "FTP RECV", logKeyAction, "ftp.cmd_recv", "line", line)
		}

		c.handleCommand(line)
	}
}

// handleCommand takes care of executing the received line
func (c *clientHandler) handleCommand(line string) {
	command, param := parseLine(line)
	c.command = strings.ToUpper(command)
	c.param = param

	cmdDesc := commandsMap[c.command]
	if cmdDesc == nil {
		c.writeMessage(500, "Unknown command")
		return
	}

	if c.driver == nil && !cmdDesc.Open {
		c.writeMessage(530, "Please login with USER and PASS")
		return
	}

	// Let's prepare to recover in case there's a command error
	defer func() {
		if r := recover(); r != nil {
			c.writeMessage(500, fmt.Sprintf("Internal error: %s", r))
		}
	}()
	cmdDesc.Fn(c)
}

func (c *clientHandler) writeLine(line string) {
	if c.debug {
		level.Debug(c.logger).Log(logKeyMsg, "FTP SEND", logKeyAction, "ftp.cmd_send", "line", line)
	}
	c.writer.Write([]byte(line))
	c.writer.Write([]byte("\r\n"))
	c.writer.Flush()
}

func (c *clientHandler) writeMessage(code int, message string) {
	c.writeLine(fmt.Sprintf("%d %s", code, message))
}

func (c *clientHandler) TransferOpen() (net.Conn, error) {
	if c.transfer == nil {
		c.writeMessage(550, "No passive connection declared")
		return nil, errors.New("no passive connection declared")
	}
	c.writeMessage(150, "Using transfer connection")
	conn, err := c.transfer.Open()
	if err == nil && c.debug {
		level.Debug(c.logger).Log(logKeyMsg, "FTP Transfer connection opened", logKeyAction, "ftp.transfer_open", "remoteAddr", conn.RemoteAddr().String(), "localAddr", conn.LocalAddr().String())
	}
	return conn, err
}

func (c *clientHandler) TransferClose() {
	if c.transfer != nil {
		c.writeMessage(226, "Closing transfer connection")
		c.transfer.Close()
		c.transfer = nil
		if c.debug {
			level.Debug(c.logger).Log(logKeyMsg, "FTP Transfer connection closed", logKeyAction, "ftp.transfer_close")
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

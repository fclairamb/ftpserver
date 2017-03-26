package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

type clientHandler struct {
	Id          uint32               // Id of the client
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
	ctx_rnfr    string               // Rename from
	ctx_rest    int64                // Restart point
	debug       bool                 // Show debugging info on the server side
	transfer    transferHandler      // Transfer connection (only passive is implemented at this stage)
	transferTls bool                 // Use TLS for transfer connection
}

func (server *FtpServer) NewClientHandler(connection net.Conn) *clientHandler {

	server.clientCounter += 1

	p := &clientHandler{
		daddy:       server,
		conn:        connection,
		Id:          server.clientCounter,
		writer:      bufio.NewWriter(connection),
		reader:      bufio.NewReader(connection),
		connectedAt: time.Now().UTC(),
		path:        "/",
	}

	// Just respecting the existing logic here, this could be probably be dropped at some point

	return p
}

func (c *clientHandler) disconnect() {
	c.conn.Close()
}

func (c *clientHandler) Path() string {
	return c.path
}

func (c *clientHandler) SetPath(path string) {
	c.path = path
}

func (c *clientHandler) Debug() bool {
	return c.debug
}

func (c *clientHandler) SetDebug(debug bool) {
	c.debug = debug
}

func (c *clientHandler) end() {
	if c.transfer != nil {
		c.transfer.Close()
	}
}

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
				log15.Debug("Clean disconnect", "action", "ftp.disconnect", "id", c.Id, "clean", true)
			}
			return
		}

		line, err := c.reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if c.debug {
					log15.Debug("TCP disconnect", "action", "ftp.disconnect", "id", c.Id, "clean", false)
				}
			} else {
				log15.Error("Read error", "action", "ftp.read_error", "id", c.Id, "err", err)
			}
			return
		}

		if c.debug {
			log15.Debug("FTP RECV", "action", "ftp.cmd_recv", "id", c.Id, "line", line)
		}

		command, param := parseLine(line)
		c.command = strings.ToUpper(command)
		c.param = param

		fn := commandsMap[c.command]
		if fn == nil {
			c.writeMessage(550, "Not handled")
		} else {
			fn(c)
		}
	}
}

func (c *clientHandler) writeLine(line string) {
	if c.debug {
		log15.Debug("FTP SEND", "action", "ftp.cmd_send", "id", c.Id, "line", line)
	}
	c.writer.Write([]byte(line))
	c.writer.Write([]byte("\r\n"))
	c.writer.Flush()
}

func (c *clientHandler) writeMessage(code int, message string) {
	c.writeLine(fmt.Sprintf("%d %s", code, message))
}

func (c *clientHandler) TransferOpen() (net.Conn, error) {
	if c.transfer != nil {
		c.writeMessage(150, "Using transfer connection")
		conn, err := c.transfer.Open()
		if err == nil && c.debug {
			log15.Debug("FTP Transfer connection opened", "action", "ftp.transfer_open", "id", c.Id, "remoteAddr", conn.RemoteAddr().String(), "localAddr", conn.LocalAddr().String())
		}
		return conn, err
	} else {
		c.writeMessage(550, "No passive connection declared")
		return nil, errors.New("No passive connection declared")
	}
}

func (c *clientHandler) TransferClose() {
	if c.transfer != nil {
		c.writeMessage(226, "Closing transfer connection")
		c.transfer.Close()
		c.transfer = nil
		if c.debug {
			log15.Debug("FTP Transfer connection closed", "action", "ftp.transfer_close", "id", c.Id)
		}
	}
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

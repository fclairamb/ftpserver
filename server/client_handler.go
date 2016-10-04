package server

import (
	"gopkg.in/inconshreveable/log15.v2"
	"fmt"
	"net"
	"bufio"
	"time"
	"strings"
	"errors"
	"io"
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
	userInfo    map[string]string    // Various user information (shared between server and driver)
	debug       bool                 // Show debugging info on the server side
	transfer    transferHandler      // Transfer connection (only passive is implemented at this stage)
	transferTls bool                 // Use TLS for transfer connection
}

func (server *FtpServer) NewClientHandler(connection net.Conn) *clientHandler {

	server.clientCounter += 1

	p := &clientHandler{
		daddy: server,
		conn: connection,
		Id: server.clientCounter,
		writer: bufio.NewWriter(connection),
		reader: bufio.NewReader(connection),
		connectedAt: time.Now().UTC(),
		path: "/",
		userInfo: make(map[string]string),
	}

	// Just respecting the existing logic here, this could be probably be dropped at some point

	return p
}

func (c *clientHandler) disconnect() {
	c.conn.Close()
}

func (c *clientHandler) UserInfo() map[string]string {
	return c.userInfo
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
		c.writeMessage(500, "Can't accept you - " + err.Error())
		return
	}

	defer c.daddy.driver.UserLeft(c)

	//fmt.Println(p.id, " Got client on: ", p.ip)
	if msg, err := c.daddy.driver.WelcomeUser(c); err == nil {
		c.writeMessage(220, msg)
	} else {
		c.writeMessage(500, msg)
		return
	}

	for {
		line, err := c.reader.ReadString('\n')

		if c.debug {
			log15.Info("FTP RECV", "action", "ftp.cmd_recv", "line", line)
		}

		if err != nil {
			if err == io.EOF {
				log15.Info("Client disconnected", "id", c.Id)
			} else {
				log15.Error("TCP error", "err", err)
			}
			return
		}

		command, param := parseLine(line)
		c.command = command
		c.param = param

		fn := commandsMap[command]
		if fn == nil {
			c.writeMessage(550, "not allowed")
		} else {
			fn(c)
		}
	}
}

func (c *clientHandler) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	if c.debug {
		log15.Info("FTP SEND", "action", "ftp.cmd_send", "line", line)
	}
	c.writer.WriteString(line)
	c.writer.Flush()
}

func (c *clientHandler) TransferOpen() (net.Conn, error) {
	if c.transfer != nil {
		c.writeMessage(150, "Using transfer connection")
		return c.transfer.Open()
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
	}
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

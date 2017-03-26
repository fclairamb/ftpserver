package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (c *clientHandler) handlePORT() {
	raddr := parseRemoteAddr(c.param)

	c.writeMessage(200, "PORT command successful")

	c.transfer = &activeTransferHandler{raddr: raddr}
}

// Active connection
type activeTransferHandler struct {
	// remote address of the client
	raddr *net.TCPAddr

	conn net.Conn
}

func (a *activeTransferHandler) Open() (net.Conn, error) {
	laddr, _ := net.ResolveTCPAddr("tcp", ":20")
	// TODO(mgenov): support dialing with timeout
	// Issues:
	//	https://github.com/golang/go/issues/3097
	// 	https://github.com/golang/go/issues/4842
	conn, err := net.DialTCP("tcp", laddr, a.raddr)

	if err != nil {
		return nil, fmt.Errorf("could not establish active connection due: %v", err)
	}

	// keep connection as it will be closed by Close()
	a.conn = conn

	return a.conn, nil
}

// Close closes only if connection is established
func (a *activeTransferHandler) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// parseRemoteAddr parses remote address of the client from param. This address
// is used for establishing a connection with the client.
//
// Param Format: 192,168,150,80,14,178
// Host: 192.168.150.80
// Port: (14 * 256) + 148
func parseRemoteAddr(param string) *net.TCPAddr {
	//TODO(mgenov): ensure that format of the params is valid
	params := strings.Split(param, ",")
	ip := strings.Join(params[0:4], ".")

	p1, _ := strconv.Atoi(params[4])
	p2, _ := strconv.Atoi(params[5])
	port := (p1 * 256) + p2

	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port))
	return addr
}

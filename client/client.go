package client

import "net"
import "fmt"
import "net/textproto"
import "bufio"

type Client struct {
	address string
	reader  *textproto.Reader
	writer  *textproto.Writer
	conn    net.Conn
}

func NewClient() *Client {
	c := Client{}
	c.address = "127.0.0.1:2121"
	return &c
}

func (c *Client) read() {
	code, msg, err := c.reader.ReadResponse(0)
	fmt.Println(code, msg, err)
}
func (c *Client) send(text string) {
	err := c.writer.PrintfLine("USER Bad")
	fmt.Println(err)
}

func (c *Client) Connect() {
	c.conn, _ = net.DialTimeout("tcp", c.address, 10000000)

	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.writer = textproto.NewWriter(bufio.NewWriter(c.conn))

	c.read()
	c.send("USER bad")
	c.read()
	c.send("PASS security")
}

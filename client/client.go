package client

import "net"
import "fmt"
import "net/textproto"
import "bufio"
import "strings"
import "io"

type Client struct {
	address    string
	reader     *textproto.Reader
	writer     *textproto.Writer
	conn       net.Conn
	passive    net.Conn
	lastMsg    string
	passReader *bufio.Reader
	passWriter *bufio.Writer
}

func NewClient() *Client {
	c := Client{}
	c.address = "127.0.0.1:2121"
	return &c
}

func (c *Client) read() {
	code, msg, err := c.reader.ReadResponse(0)
	fmt.Println(code, msg)
	c.lastMsg = msg
	if err != nil {
		fmt.Println(err)
	}
}

// how to break, send USER bad everytime
func (c *Client) send(text string) {
	err := c.writer.PrintfLine(text)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Client) Connect() {
	c.conn, _ = net.DialTimeout("tcp", c.address, 10000000)

	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.writer = textproto.NewWriter(bufio.NewWriter(c.conn))

	c.read()
	c.send("USER bad")
	c.read()
	c.send("PASS security")
	c.read()
}

func (c *Client) List() {
	c.openPassive()
	c.send("LIST")
	c.read()
	for {
		line, err := c.passReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Print(line)
		if err != nil {
			fmt.Println(err)
		}
	}
	if true {
		c.passive.Close()
	}
}

func (c *Client) openPassive() {
	c.send("EPSV")
	c.read()
	fmt.Println("PORT ", c.lastMsg)

	port := strings.TrimRight(c.lastMsg, "(|)")[35:40]
	c.passive, _ = net.DialTimeout("tcp", "127.0.0.1:"+port, 10000000)
	c.passReader = bufio.NewReader(c.passive)
	c.passWriter = bufio.NewWriter(c.passive)
}

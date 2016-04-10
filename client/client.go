package client

import "net"
import "fmt"
import "net/textproto"
import "bufio"
import "strings"
import "math/rand"
import "time"

type Client struct {
	reader     *textproto.Reader
	writer     *textproto.Writer
	conn       net.Conn
	passive    net.Conn
	lastMsg    string
	passReader *bufio.Reader
	passWriter *bufio.Writer
	id         int
}

func NewClient(id int) *Client {
	c := Client{}
	c.id = id
	return &c
}

func (c *Client) read(print bool) {
	if c.reader == nil {
		return
	}
	code, msg, err := c.reader.ReadResponse(0)
	if print {
		fmt.Println(code, msg)
	}
	c.lastMsg = msg
	if err != nil {
		fmt.Println(err)
	}
}

// how to break, send USER bad everytime
func (c *Client) send(text string) {
	if c.writer == nil {
		return
	}
	err := c.writer.PrintfLine(text)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Client) Connect() {
	var err error
	c.conn, err = net.DialTimeout("tcp", "127.0.0.1:2121", 10000000)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.reader = textproto.NewReader(bufio.NewReader(c.conn))
	c.writer = textproto.NewWriter(bufio.NewWriter(c.conn))

	c.read(false)
	c.send("USER bad")
	c.read(false)
	c.send("PASS security")
	c.read(false)
}

func (c *Client) List() {
	c.openPassive()
	c.send("LIST")
	c.read(false)
	for {
		line, err := c.passReader.ReadString('\n')
		if line == "\r\n" {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	if true {
		c.passive.Close()
	}
	c.read(false)
}

func (c *Client) Quit() {
	c.send("QUIT")
	c.read(true)
}

func (c *Client) Stor(size int64) {
	//fmt.Println(c.id, " Sending ", size)
	c.openPassive()
	c.send("STOR fake_file.dat")
	c.read(false)
	c.passWriter.Write(fakeFile(size))
	c.passWriter.Flush()
	if true {
		c.passive.Close()
	}
	c.read(false)
}

func fakeFile(size int64) []byte {
	bytes := make([]byte, size)
	s1 := rand.NewSource(time.Now().UnixNano())
	for i := int64(0); i < size; i++ {
		bytes[i] = byte(s1.Int63())
	}
	return bytes
}

func (c *Client) openPassive() {
	c.send("EPSV")
	c.read(false)
	//fmt.Println("PORT ", c.lastMsg)

	port := strings.TrimRight(c.lastMsg, "(|)")[35:40]
	c.passive, _ = net.DialTimeout("tcp", "127.0.0.1:"+port, 10000000)
	c.passReader = bufio.NewReader(c.passive)
	c.passWriter = bufio.NewWriter(c.passive)
}

package main

import "testing"
import "os"
import "io"
import "fmt"
import "net"
import "bufio"
import "net/textproto"
import "time"
import "strings"
import "paradise/server"
import "paradise/client"

var file *os.File
var fileBytes []byte

func TestMain(m *testing.M) {

	file, _ = os.Open("/Users/aa/Movies/hi5.mp4")
	fileBytes = make([]byte, 512)
	file.Read(fileBytes)

	os.Exit(m.Run())
}

func TestSimple(t *testing.T) {
	go server.Start()
	time.Sleep(1 * (time.Second * 1))
	testConnect(t)
	if false {
		t.Errorf("test")
	}
}

func openPassive(reader *textproto.Reader, writer *textproto.Writer) (passive net.Conn, passReader *bufio.Reader, passWriter *bufio.Writer) {
	err := writer.PrintfLine("EPSV")
	_, msg, err := reader.ReadResponse(0)
	fmt.Println("PORT ", msg)
	port := strings.TrimRight(msg, "(|)")[35:40]
	passive, err = net.DialTimeout("tcp", "127.0.0.1:"+port, 10000000)
	fmt.Println(passive, err)
	passReader = bufio.NewReader(passive)
	passWriter = bufio.NewWriter(passive)
	return
}

func testConnect(t *testing.T) {
	c := client.NewClient()
	fmt.Println(c)
}

func testConnect3(t *testing.T) {
	conn, _ := net.DialTimeout("tcp", "127.0.0.1:2121", 10000000)

	reader := textproto.NewReader(bufio.NewReader(conn))
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	code, msg, err := reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("USER Bad")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("PASS Security")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("QUIT")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)
}

func testCommandList(t *testing.T) {
	conn, _ := net.DialTimeout("tcp", "127.0.0.1:2121", 10000000)

	reader := textproto.NewReader(bufio.NewReader(conn))
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	code, msg, err := reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("USER Bad")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("PASS Security")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	_, passReader, _ := openPassive(reader, writer)

	err = writer.PrintfLine("LIST")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)
	for {
		line, err := passReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Println(line, err)
	}
	//fmt.Println("Closing Passive")
	//passive.Close()
	fmt.Println("Closed")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)
}

func testConnect2(t *testing.T) {
	conn, _ := net.DialTimeout("tcp", "127.0.0.1:2121", 10000000)

	reader := textproto.NewReader(bufio.NewReader(conn))
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	code, msg, err := reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("USER Bad")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("PASS Security")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	err = writer.PrintfLine("CWD the_matrix")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	passive, passReader, passWriter := openPassive(reader, writer)

	err = writer.PrintfLine("LIST")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)
	for {
		line, err := passReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Println(line, err)
	}
	fmt.Println("Closing Passive")
	passive.Close()
	fmt.Println("Closed")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	passive, passReader, passWriter = openPassive(reader, writer)

	err = writer.PrintfLine("STOR big.mov")
	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	passWriter.Write(fileBytes)
	passWriter.Flush()

	passive.Close()

	code, msg, err = reader.ReadResponse(0)
	fmt.Println(code, msg, err)

	time.Sleep(1 * time.Minute)
}

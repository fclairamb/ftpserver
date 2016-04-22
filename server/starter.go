package server

import "fmt"
import "net"
import "time"
import "os"
import "github.com/andrewarrow/paradise_ftp/paradise"

func genClientID() string {
	random, _ := os.Open("/dev/urandom")
	b := make([]byte, 16)
	random.Read(b)
	random.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Start(fm *paradise.FileManager) {
	fmt.Println("starting...")
	url := fmt.Sprintf("localhost:%d", 2121) // change to 21 in production
	var listener net.Listener
	listener, err := net.Listen("tcp", url)

	if err != nil {
		fmt.Println("cannot listen on: ", url)
		return
	}
	fmt.Println("listening on: ", url)

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("listening error ", err)
			break
		}
		cid := genClientID()
		p := NewParadise(connection, cid, time.Now().Unix())
		ConnectionMap[cid] = p

		go p.HandleCommands()
	}
}

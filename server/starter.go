package server

import "fmt"
import "net"
import "time"
import "strconv"
import "math/rand"

func genClientID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	id := rand.Intn(Settings.MaxConnections)
	return strconv.FormatInt(int64(id), 16)
}

func Start() {
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

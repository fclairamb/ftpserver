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
	Settings = ReadSettings()
	fmt.Println(Settings)
	CommandMap = MakeCommandMap()
	ConnectionMap = make(map[int]*Paradise)

	url := fmt.Sprintf("localhost:%d", 2121) // change to 21 in production
	var listener net.Listener
	listener, err := net.Listen("tcp", url)

	if err != nil {
		fmt.Println("cannot listen on: ", url)
		return
	}
	fmt.Println("listening on: ", url)

	ids := 0
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("listening error ", err)
			break
		}
		ids++
		cid := genClientID()
		fmt.Println(cid)
		p := NewParadise(connection, ids, time.Now().Unix())
		ConnectionMap[ids] = p

		go p.HandleCommands()
	}
}

package server

import "fmt"
import "net"

func Start() {
	fmt.Println("starting...")
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
		p := NewParadise(connection, ids)
		ConnectionMap[ids] = p

		go p.HandleCommands()
	}
}

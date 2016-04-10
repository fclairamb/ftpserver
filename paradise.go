package main

import "paradise/server"
import "paradise/client"
import "os"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		go client.StressTest()
	}
	go server.Monitor()
	server.Start()
}

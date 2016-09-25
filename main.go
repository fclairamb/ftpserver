package main

import (
	"flag"
	"github.com/fclairamb/ftpserver/client"
	"github.com/fclairamb/ftpserver/sample"
	"github.com/fclairamb/ftpserver/server"
)

var (
	gracefulChild = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
	stressTest    = flag.Bool("stressTest", false, "start a client making connections")
)

func main() {
	flag.Parse()
	if *stressTest {
		go client.StressTest()
	}
	go server.Monitor()
	server.Start(sample.NewSampleDriver(), *gracefulChild)
}

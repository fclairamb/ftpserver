package main

import "github.com/andrewarrow/paradise_ftp/server"
import "github.com/andrewarrow/paradise_ftp/client"
import "github.com/andrewarrow/paradise_ftp/paradise"
import "flag"

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
	fm := paradise.NewDefaultFileSystem()
	am := paradise.NewDefaultAuthSystem()
	server.Start(fm, am, *gracefulChild)
}

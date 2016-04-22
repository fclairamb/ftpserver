package main

import "github.com/andrewarrow/paradise_ftp/server"
import "github.com/andrewarrow/paradise_ftp/client"
import "github.com/andrewarrow/paradise_ftp/paradise"
import "os"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		go client.StressTest()
	}
	go server.Monitor()
	fm := paradise.NewDefaultFileSystem()
	am := paradise.NewDefaultAuthSystem()
	server.Start(fm, am)
}

// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fclairamb/ftpserver/sample"
	"github.com/fclairamb/ftpserver/server"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	ftpServer *server.FtpServer
)

func main() {
	flag.Parse()
	ftpServer = server.NewFtpServer(sample.NewSampleDriver())

	go signalHandler()

	err := ftpServer.ListenAndServe()
	if err != nil {
		log15.Error("Problem listening", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	for {
		switch <-ch {
		case syscall.SIGTERM:
			ftpServer.Stop()
			break
		}
	}
}

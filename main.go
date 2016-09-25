package main

import (
	"flag"
	"github.com/fclairamb/ftpserver/client"
	"github.com/fclairamb/ftpserver/sample"
	"github.com/fclairamb/ftpserver/server"
	"os/signal"
	"syscall"
	"os"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	gracefulChild = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
	stressTest = flag.Bool("stressTest", false, "start a client making connections")
	ftpServer *server.FtpServer
)

func main() {
	flag.Parse()
	if *stressTest {
		go client.StressTest()
	}
	ftpServer = server.NewFtpServer(sample.NewSampleDriver())

	go signalHandler()

	err := ftpServer.ListenAndServe(*gracefulChild)
	if err != nil {
		log15.Error("Problem listening", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		ftpServer.HandleSignal(<- ch)
	}
}

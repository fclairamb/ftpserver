// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fclairamb/ftpserver/sample"
	"github.com/fclairamb/ftpserver/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var (
	ftpServer *server.FtpServer
)

func main() {
	flag.Parse()

	logger := log.With(
		log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	level.Info(logger).Log("msg", "Sample server")

	driver, err := sample.NewSampleDriver()
	if err != nil {
		level.Error(logger).Log("msg", "Could not load the driver", "err", err)
		return
	}
	driver.Logger = log.With(logger, "component", "driver")

	ftpServer = server.NewFtpServer(driver)
	ftpServer.Logger = log.With(logger, "component", "server")

	go signalHandler()

	err = ftpServer.ListenAndServe()
	if err != nil {
		level.Error(logger).Log("msg", "Problem listening", "err", err)
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

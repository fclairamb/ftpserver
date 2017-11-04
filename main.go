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
	// Parsing arguments
	confFile := flag.String("conf", "sample/conf/settings.toml", "Configuration file")
	dataDir := flag.String("data", "", "Data directory")
	flag.Parse()

	// Setting up the logger
	logger := log.With(
		log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	// Loading the driver
	driver, err := sample.NewSampleDriver(*dataDir, *confFile)

	if err != nil {
		level.Error(logger).Log("msg", "Could not load the driver", "err", err)
		return
	}

	// Overriding the driver default silent logger by a sub-logger (component: driver)
	driver.Logger = log.With(logger, "component", "driver")

	// Instantiating the server by passing our driver implementation
	ftpServer = server.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	ftpServer.Logger = log.With(logger, "component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if err := ftpServer.ListenAndServe(); err != nil {
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

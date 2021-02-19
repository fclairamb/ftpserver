// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"github.com/moovfinancial/ftpserver/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/moovfinancial/ftpserverlib"
	"github.com/moovfinancial/ftpserverlib/log/gokit"

	"github.com/moovfinancial/ftpserver/config"
	"github.com/moovfinancial/ftpserver/server"
)

var (
	ftpServer *ftpserver.FtpServer
	driver    *server.Server
)

func main() {
	// Arguments vars
	var confFile string
	var onlyConf bool

	// Parsing arguments
	flag.StringVar(&confFile, "conf", "", "Configuration file")
	flag.BoolVar(&onlyConf, "conf-only", false, "Only create the conf")
	flag.Parse()

	// Setting up the logger
	logger := gokit.NewGKLoggerStdout().With(
		"ts", gokit.GKDefaultTimestampUTC,
		"caller", gokit.GKDefaultCaller,
	)

	autoCreate := onlyConf

	// The general idea here is that if you start it without any arg, you're probably doing a local quick&dirty run
	// possibly on a windows machine, so we're better of just using a default file name and create the file.
	if confFile == "" {
		confFile = "ftpserver.json"
		autoCreate = true
	}

	if autoCreate {
		if _, err := os.Stat(confFile); err != nil && os.IsNotExist(err) {
			logger.Warn("No conf file, creating one", "confFile", confFile)

			if err := ioutil.WriteFile(confFile, confFileContent(), 0600); err != nil {
				logger.Warn("Couldn't create conf file", "confFile", confFile)
			}
		}
	}

	conf, errConfig := config.NewConfig(confFile, logger)
	if errConfig != nil {
		logger.Error("Can't load conf", "err", errConfig)
		return
	}

	// Setup folders
	for _, access := range conf.Content.Accesses {
		fileSys, errAccess := fs.LoadFs(access, logger)
		if errAccess != nil {
			logger.Error("Config: Invalid access !", "err", errAccess, "username", access.User, "fs", access.Fs)
			return
		}
		fileSys.Mkdir(conf.Content.Inbound, 0755)
	}

	// Loading the driver
	var errNewServer error
	driver, errNewServer = server.NewServer(conf, logger.With("component", "driver"))

	if errNewServer != nil {
		logger.Error("Could not load the driver", "err", errNewServer)
		return
	}

	// Instantiating the server by passing our driver implementation
	ftpServer = ftpserver.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	ftpServer.Logger = logger.With("component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if onlyConf {
		logger.Warn("Only creating conf")
		return
	}

	if err := ftpServer.ListenAndServe(); err != nil {
		logger.Error("Problem listening", "err", err)
	}

	// We wait at most 1 minutes for all clients to disconnect
	if err := driver.WaitGracefully(time.Minute); err != nil {
		ftpServer.Logger.Warn("Problem stopping server", "err", err)
	}
}

func stop() {
	driver.Stop()

	if err := ftpServer.Stop(); err != nil {
		ftpServer.Logger.Error("Problem stopping server", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		sig := <-ch

		if sig == syscall.SIGTERM {
			stop()
			break
		}
	}
}

func confFileContent() []byte {
	str := `{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "os",
      "params": {
        "basePath": "/tmp"
      }
    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  },
  "inbound": "inbound"
}`

	return []byte(str)
}

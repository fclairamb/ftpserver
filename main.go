// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	gkwrap "github.com/fclairamb/go-log/gokit"
	gokit "github.com/go-kit/log"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/server"
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
	logger := gkwrap.New()

	logger.Info("FTP server", "version", BuildVersion, "date", BuildDate, "commit", Commit)

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

			if err := ioutil.WriteFile(confFile, confFileContent(), 0600); err != nil { //nolint: gomnd
				logger.Warn("Couldn't create conf file", "confFile", confFile)
			}
		}
	}

	conf, errConfig := config.NewConfig(confFile, logger)
	if errConfig != nil {
		logger.Error("Can't load conf", "err", errConfig)

		return
	}

	// Now is a good time to open a logging file
	if conf.Content.Logging.File != "" {
		writer, err := os.OpenFile(conf.Content.Logging.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600) //nolint:gomnd

		if err != nil {
			logger.Error("Can't open log file", "err", err)

			return
		}

		logger = gkwrap.NewWrap(gokit.NewLogfmtLogger(io.MultiWriter(writer, os.Stdout))).With(
			"ts", gokit.DefaultTimestampUTC,
			"caller", gokit.DefaultCaller,
		)
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
  }
}`

	return []byte(str)
}

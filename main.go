// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	gklog "github.com/go-kit/kit/log"

	"github.com/fclairamb/ftpserver/sample"
	"github.com/fclairamb/ftpserver/server"
	"github.com/fclairamb/ftpserver/server/log"
)

var (
	ftpServer *server.FtpServer
)

func main() {
	// Arguments vars
	var confFile, dataDir string

	var onlyConf bool

	// Parsing arguments
	flag.StringVar(&confFile, "conf", "", "Configuration file")
	flag.StringVar(&dataDir, "data", "", "Data directory")
	flag.BoolVar(&onlyConf, "conf-only", false, "Only create the config")
	flag.Parse()

	// Setting up the logger
	logger := log.NewGKLogger(gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stdout))).With(
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	autoCreate := onlyConf

	// The general idea here is that if you start it without any arg, you're probably doing a local quick&dirty run
	// possibly on a windows machine, so we're better of just using a default file name and create the file.
	if confFile == "" {
		confFile = "settings.toml"
		autoCreate = true
	}

	if autoCreate {
		if _, err := os.Stat(confFile); err != nil && os.IsNotExist(err) {
			logger.Error("msg", "Not config file, creating one", "action", "conf_file.create", "confFile", confFile)

			if err := ioutil.WriteFile(confFile, confFileContent(), 0644); err != nil {
				logger.Error("msg", "Couldn't create config file", "action", "conf_file.could_not_create", "confFile", confFile)
			}
		}
	}

	// Loading the driver
	driver, err := sample.NewSampleDriver(dataDir, confFile)

	if err != nil {
		logger.Error("msg", "Could not load the driver", "err", err)
		return
	}

	// Overriding the driver default silent logger by a sub-logger (component: driver)
	driver.Logger = logger.With("component", "driver")

	// Instantiating the server by passing our driver implementation
	ftpServer = server.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	ftpServer.Logger = logger.With("component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if onlyConf {
		logger.Error("msg", "Only creating conf")
		return
	}

	if err := ftpServer.ListenAndServe(); err != nil {
		logger.Error("msg", "Problem listening", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		sig := <-ch

		if sig == syscall.SIGTERM {
			ftpServer.Stop()
		}
	}
}

func confFileContent() []byte {
	str := `# ftpserver configuration file
#
# These are all the config parameters with their default values. If not present,

# Max number of control connections to accept
# max_connections = 0
max_connections = 10

[server]
# Address and Port to listen on
# listen_addr="0.0.0.0:2121"

# Public host to expose in the passive connection
# public_host = ""

# Idle timeout time
# idle_timeout = 900

# Data port range from 10000 to 15000
# [passiveTransferPortRange]
# start = 2122
# end = 2200

[server.passiveTransferPortRange]
start = 2122
end = 2200

[[users]]
user="fclairamb"
pass="floflo"
dir="shared"

[[users]]
user="test"
pass="test"
dir="shared"

[[users]]
user="mcardon"
pass="marmar"
dir="marie"
`

	return []byte(str)
}

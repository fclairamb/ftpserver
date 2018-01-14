// ftpserver allows to create your own FTP(S) server
package main

import (
	"flag"
	"io/ioutil"
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
	// Arguments vars
	var confFile, dataDir string
	var onlyConf bool

	// Parsing arguments
	flag.StringVar(&confFile, "conf", "", "Configuration file")
	flag.StringVar(&dataDir, "data", "", "Data directory")
	flag.BoolVar(&onlyConf, "conf-only", false, "Only create the config")
	flag.Parse()

	// Setting up the logger
	logger := log.With(
		log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
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
			level.Info(logger).Log("msg", "Not config file, creating one", "action", "conf_file.create", "confFile", confFile)

			if err := ioutil.WriteFile(confFile, confFileContent(), 0644); err != nil {
				level.Error(logger).Log("msg", "Couldn't create config file", "action", "conf_file.could_not_create", "confFile", confFile)
			}
		}
	}

	// Loading the driver
	driver, err := sample.NewSampleDriver(dataDir, confFile)

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
	if onlyConf {
		level.Error(logger).Log("msg", "Only creating conf")
		return
	}

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

func confFileContent() []byte {
	str := `# ftpserver configuration file
#
# These are all the config parameters with their default values. If not present,

# Max number of control connections to accept
# max_connections = 0
max_connections = 10

[server]
# Address to listen on
# listen_host = "0.0.0.0"

# Port to listen on
# listen_port = 2121

# Public host to expose in the passive connection
# public_host = ""

# Idle timeout time
# idle_timeout = 900

# Data port range from 10000 to 15000
# [dataPortRange]
# start = 2122
# end = 2200

[server.dataPortRange]
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

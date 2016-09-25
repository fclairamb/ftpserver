package server

import (
	"os/exec"
	"crypto/rand"
	"os"
	"time"
	"fmt"
	"syscall"
	"net"
)

var Settings ParadiseSettings
var err error
var FinishAndStop bool

// TODO: Consider if we actually need it
func genClientID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (server *FtpServer) HandleSignal(sig os.Signal) {
	//ch := make(chan os.Signal, 10)
	//signal.Notify(ch, syscall.SIGTERM, syscall.SIGUSR2)

	// From now on, we can't just handle signals for the complete program, we would have to transfer them to us.
	switch sig {
	case syscall.SIGTERM:
		FinishAndStop = true
		return
	case syscall.SIGUSR2:
		file, _ := server.Listener.(*net.TCPListener).File()
		path := Settings.Exec
		args := []string{
			"-graceful"}
		cmd := exec.Command(path, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.ExtraFiles = []*os.File{file}
		err := cmd.Start()
		fmt.Println("forking err is ", err)
	}
}

func (server *FtpServer) ListenAndServe(gracefulChild bool) error {
	Settings = ReadSettings()
	FinishAndStop = false
	fmt.Println("starting...")

	if gracefulChild {
		f := os.NewFile(3, "") // FD 3 is a special file descriptor to get an already-opened socket
		server.Listener, err = net.FileListener(f)
	} else {
		url := fmt.Sprintf("%s:%d", Settings.Host, Settings.Port)
		server.Listener, err = net.Listen("tcp", url)
	}

	if err != nil {
		fmt.Println("cannot listen: ", err)
		return err
	}
	fmt.Println("listening...")

	if gracefulChild {
		parent := syscall.Getppid()
		syscall.Kill(parent, syscall.SIGTERM)
	}

	// The actual signal handler of the core program will do that (if he wants to)
	// go signalHandler()

	for {
		if FinishAndStop {
			break
		}
		server.Listener.(*net.TCPListener).SetDeadline(time.Now().Add(60 * time.Second))
		connection, err := server.Listener.Accept()
		if err != nil {
			if opError, ok := err.(*net.OpError); !ok || !opError.Timeout() {
				fmt.Println("listening error ", err)
			}
		} else {
			cid := genClientID()
			p := server.NewClientHandler(connection, cid, time.Now().Unix())
			server.ConnectionMap[cid] = p

			go p.HandleCommands()
		}
	}

	// TODO add wait group for still active connections to finish up
	// otherwise, this will just exit and kill them
	// defeating whole point of gracefulChild restart
	return nil
}

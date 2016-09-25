package server

import "strconv"
import "bytes"
import "fmt"
import "strings"
import "time"
import (
	"github.com/jehiah/go-strftime"
	"path/filepath"
)

func (p *ClientHandler) HandleCwd() {
	if p.param == ".." {
		p.HandleCdUp()
		return
	}

	path := p.Path()

	if strings.HasPrefix(p.param, "/") {
		path = p.param
	} else {
		if path != "/" {
			path += "/"
		}
		path += p.param
	}

	// TODO: Find something smarter, this is obviously quite limitating...
	if path == "/debug" {
		p.writeMessage(250, "Debug activated !")
		p.debug = true
		return
	}

	if err := p.daddy.driver.GoToDirectory(p, path); err == nil {
		p.SetPath(path)
		p.writeMessage(250, fmt.Sprintf("CD worked on %s", path))
	} else {
		p.writeMessage(550, fmt.Sprintf("CD issue: %s", err.Error()))
	}
}

func (p *ClientHandler) HandleCdUp() {
	dirs := filepath.SplitList(p.Path())
	dirs = dirs[0:len(dirs) - 1]
	path := filepath.Join(dirs...)
	if path == "" {
		path = "/"
	}
	if err == nil {
		if err = p.daddy.driver.GoToDirectory(p, path); err == nil {
			p.SetPath(path)
			p.writeMessage(250, fmt.Sprintf("CDUP worked on %s", path))
		}
	}

	if err != nil {
		p.writeMessage(550, fmt.Sprintf("CDUP issue: %s", err.Error()))
	}
}

func (p *ClientHandler) HandlePwd() {
	p.writeMessage(257, "\"" + p.Path() + "\" is the current directory")
}

func (p *ClientHandler) HandleList() {
	passive := p.lastPassive()
	if passive == nil {
		return
	}
	p.writeMessage(150, "Opening ASCII mode data connection for file list")

	bytes, err := p.dirList()
	if err != nil {
		p.writeMessage(550, err.Error())
	} else {
		if waitTimeout(&passive.waiter, time.Minute) {
			p.writeMessage(550, "Could not get passive connection.")
			p.closePassive(passive)
			return
		}
		if passive.listenFailedAt > 0 {
			p.writeMessage(550, "Could not get passive connection.")
			p.closePassive(passive)
			return
		}
		passive.connection.Write(bytes)
		message := "Closing data connection, sent some bytes"
		p.writeMessage(226, message)
	}
	p.closePassive(passive)
}

func (p *ClientHandler) dirList() ([]byte, error) {
	var buf bytes.Buffer

	files, err := p.daddy.driver.GetFiles(p)
	for _, file := range files {

		if file["isDir"] != "" {
			fmt.Fprintf(&buf, "drw-r--r--")
		} else {
			fmt.Fprintf(&buf, "-rw-r--r--")
		}
		fmt.Fprintf(&buf, " 1 %s %s ", "paradise", "ftp")
		fmt.Fprintf(&buf, lpad(file["size"], 12))
		ts, _ := strconv.ParseInt(file["modTime"], 10, 64)
		fmt.Fprintf(&buf, strftime.Format(" %b %d %H:%M ", time.Unix(ts, 0)))
		fmt.Fprintf(&buf, "%s\r\n", file["name"])
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.Bytes(), err
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length - len(input)) + input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}

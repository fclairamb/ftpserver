package server

import (
	"github.com/jehiah/go-strftime"
	"path/filepath"
	"strings"
	"fmt"
	"time"
	"strconv"
	"bytes"
)

func (c *ClientHandler) absPath(p string) string {
	path := c.Path()

	if strings.HasPrefix(p, "/") {
		path = p
	} else {
		if path != "/" {
			path += "/"
		}
		path += p
	}

	return path
}

func (c *ClientHandler) HandleCwd() {
	if c.param == ".." {
		c.HandleCdUp()
		return
	}

	path := c.absPath(c.param)

	// TODO: Find something smarter, this is obviously quite limitating...
	if path == "/debug" {
		c.writeMessage(250, "Debug activated !")
		c.debug = true
		return
	}

	if err := c.daddy.driver.ChangeDirectory(c, path); err == nil {
		c.SetPath(path)
		c.writeMessage(250, fmt.Sprintf("CD worked on %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("CD issue: %s", err.Error()))
	}
}

func (c *ClientHandler) HandleMkd() {
	path := c.absPath(c.param)
	if err := c.daddy.driver.MakeDirectory(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Created dir %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not create %s : %s", path, err.Error()))
	}
}

func (c *ClientHandler) HandleCdUp() {
	dirs := filepath.SplitList(c.Path())
	dirs = dirs[0:len(dirs) - 1]
	path := filepath.Join(dirs...)
	if path == "" {
		path = "/"
	}
	if err := c.daddy.driver.ChangeDirectory(c, path); err == nil {
		c.SetPath(path)
		c.writeMessage(250, fmt.Sprintf("CDUP worked on %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("CDUP issue: %s", err.Error()))
	}
}

func (c *ClientHandler) HandlePwd() {
	c.writeMessage(257, "\"" + c.Path() + "\" is the current directory")
}

func (c *ClientHandler) HandleList() {
	passive := c.lastPassive()
	if passive == nil {
		return
	}
	c.writeMessage(150, "Opening ASCII mode data connection for file list")

	bytes, err := c.dirList()
	if err != nil {
		c.writeMessage(550, err.Error())
	} else {
		if waitTimeout(&passive.waiter, time.Minute) {
			c.writeMessage(550, "Could not get passive connection.")
			c.closePassive(passive)
			return
		}
		if passive.listenFailedAt > 0 {
			c.writeMessage(550, "Could not get passive connection.")
			c.closePassive(passive)
			return
		}
		passive.connection.Write(bytes)
		message := "Closing data connection, sent some bytes"
		c.writeMessage(226, message)
	}
	c.closePassive(passive)
}

func (c *ClientHandler) dirList() ([]byte, error) {
	var buf bytes.Buffer

	files, err := c.daddy.driver.GetFiles(c)
	for _, file := range files {

		if file["isDir"] != "" {
			buf.WriteString("drw-r--r--")
		} else {
			buf.WriteString("-rw-r--r--")
		}
		fmt.Fprintf(&buf, " 1 %s %s ", "paradise", "ftp")
		fmt.Fprintf(&buf, "%12s", file["size"])
		ts, _ := strconv.ParseInt(file["modTime"], 10, 64)
		fmt.Fprintf(&buf, strftime.Format(" %b %d %H:%M ", time.Unix(ts, 0)))
		fmt.Fprintf(&buf, "%s\r\n", file["name"])
	}
	buf.WriteString("\r\n")
	return buf.Bytes(), err
}

// Useless, fmt.Sprintf("%12s") can do the same
/*
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
*/

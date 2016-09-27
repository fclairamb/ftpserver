package server

import (
	"github.com/jehiah/go-strftime"
	"path/filepath"
	"strings"
	"fmt"
	"time"
	"os"
	"io"
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

func (c *ClientHandler) HandleRmd() {
	path := c.absPath(c.param)
	if err := c.daddy.driver.DeleteFile(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Deleted dir %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not delete dir %s : %s", path, err.Error()))
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
	defer c.closePassive(passive)

	c.writeMessage(150, "Opening ASCII mode data connection for file list")

	files, err := c.daddy.driver.ListFiles(c)
	if err != nil {
		c.writeMessage(550, err.Error())
	} else {
		if waitTimeout(&passive.waiter, time.Minute) {
			c.writeMessage(550, "Could not get passive connection.")
			return
		}
		if passive.listenFailedAt > 0 {
			c.writeMessage(550, "Could not get passive connection.")
			return
		}
		c.dirList(passive.connection, files)
		c.writeMessage(226, "Closing data connection, sent some bytes")
	}

}

func (c *ClientHandler) dirList(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		fmt.Fprint(w, file.Mode().String())
		fmt.Fprintf(w, " 1 %s %s ", "ftp", "ftp")
		fmt.Fprintf(w, "%12d", file.Size())
		fmt.Fprintf(w, strftime.Format(" %b %d %H:%M ", file.ModTime()))
		fmt.Fprintf(w, "%s\r\n", file.Name())
	}
	fmt.Fprint(w, "\r\n")
	return nil
}

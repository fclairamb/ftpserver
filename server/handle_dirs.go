package server

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func (c *clientHandler) absPath(p string) string {
	p2 := c.Path()

	if p == "." {
		return p2
	}

	if strings.HasPrefix(p, "/") {
		p2 = p
	} else {
		if p2 != "/" {
			p2 += "/"
		}
		p2 += p
	}

	if p2 != "/" && strings.HasSuffix(p2, "/") {
		p2 = p2[0 : len(p2)-1]
	}

	return p2
}

func (c *clientHandler) handleCWD() {
	if c.param == ".." {
		c.handleCDUP()
		return
	}

	p := c.absPath(c.param)

	if err := c.driver.ChangeDirectory(c, p); err == nil {
		c.SetPath(p)
		c.writeMessage(250, fmt.Sprintf("CD worked on %s", p))
	} else {
		c.writeMessage(550, fmt.Sprintf("CD issue: %v", err))
	}
}

func (c *clientHandler) handleMKD() {
	p := c.absPath(c.param)
	if err := c.driver.MakeDirectory(c, p); err == nil {
		c.writeMessage(257, fmt.Sprintf("Created dir %s", p))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not create %s : %v", p, err))
	}
}

func (c *clientHandler) handleRMD() {
	p := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, p); err == nil {
		c.writeMessage(250, fmt.Sprintf("Deleted dir %s", p))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not delete dir %s: %v", p, err))
	}
}

func (c *clientHandler) handleCDUP() {
	parent, _ := path.Split(c.Path())
	if parent != "/" && strings.HasSuffix(parent, "/") {
		parent = parent[0 : len(parent)-1]
	}
	if err := c.driver.ChangeDirectory(c, parent); err == nil {
		c.SetPath(parent)
		c.writeMessage(250, fmt.Sprintf("CDUP worked on %s", parent))
	} else {
		c.writeMessage(550, fmt.Sprintf("CDUP issue: %v", err))
	}
}

func (c *clientHandler) handlePWD() {
	c.writeMessage(257, "\""+c.Path()+"\" is the current directory")
}

func (c *clientHandler) handleLIST() {
	if files, err := c.driver.ListFiles(c); err == nil {
		if tr, err := c.TransferOpen(); err == nil {
			defer c.TransferClose()
			c.dirTransferLIST(tr, files)
		}
	} else {
		c.writeMessage(500, fmt.Sprintf("Could not list: %v", err))
	}
}

func (c *clientHandler) handleMLSD() {
	if c.daddy.settings.DisableMLSD {
		c.writeMessage(500, "MLSD has been disabled")
		return
	}
	if files, err := c.driver.ListFiles(c); err == nil {
		if tr, err := c.TransferOpen(); err == nil {
			defer c.TransferClose()
			c.dirTransferMLSD(tr, files)
		}
	} else {
		c.writeMessage(500, fmt.Sprintf("Could not list: %v", err))
	}
}

const (
	dateFormatStatTime      = "Jan _2 15:04"          // LIST date formatting with hour and minute
	dateFormatStatYear      = "Jan _2  2006"          // LIST date formatting with year
	dateFormatStatOldSwitch = time.Hour * 24 * 30 * 6 // 6 months ago
	dateFormatMLSD          = "20060102150405"        // MLSD date formatting
)

func (c *clientHandler) fileStat(file os.FileInfo) string {

	modTime := file.ModTime()

	var dateFormat string

	if c.connectedAt.Sub(modTime) > dateFormatStatOldSwitch {
		dateFormat = dateFormatStatYear
	} else {
		dateFormat = dateFormatStatTime
	}

	return fmt.Sprintf(
		"%s 1 ftp ftp %12d %s %s",
		file.Mode(),
		file.Size(),
		file.ModTime().Format(dateFormat),
		file.Name(),
	)
}

// fclairamb (2018-02-13): #64: Removed extra empty line
func (c *clientHandler) dirTransferLIST(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		fmt.Fprintf(w, "%s\r\n", c.fileStat(file))
	}
	return nil
}

// fclairamb (2018-02-13): #64: Removed extra empty line
func (c *clientHandler) dirTransferMLSD(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		c.writeMLSxOutput(w, file)
	}
	return nil
}
func (c *clientHandler) writeMLSxOutput(w io.Writer, file os.FileInfo) {
	var listType string
	if file.IsDir() {
		listType = "dir"
	} else {
		listType = "file"
	}
	fmt.Fprintf(
		w,
		"Type=%s;Size=%d;Modify=%s; %s\r\n",
		listType,
		file.Size(),
		file.ModTime().Format(dateFormatMLSD),
		file.Name(),
	)
}

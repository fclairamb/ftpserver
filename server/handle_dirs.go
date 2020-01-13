// Package server provides all the tools to build your own FTP server: The core library and the driver.
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
	if strings.HasPrefix(p, "/") {
		return path.Clean(p)
	}

	return path.Clean(c.Path() + "/" + p)
}

func (c *clientHandler) handleCWD() error {
	p := c.absPath(c.param)

	if err := c.driver.ChangeDirectory(c, p); err == nil {
		c.SetPath(p)
		c.writeMessage(StatusFileOK, fmt.Sprintf("CD worked on %s", p))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("CD issue: %v", err))
	}

	return nil
}

func (c *clientHandler) handleMKD() error {
	p := c.absPath(c.param)
	if err := c.driver.MakeDirectory(c, p); err == nil {
		c.writeMessage(StatusPathCreated, fmt.Sprintf("Created dir %s", p))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Could not create %s : %v", p, err))
	}

	return nil
}

func (c *clientHandler) handleRMD() error {
	p := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, p); err == nil {
		c.writeMessage(StatusFileOK, fmt.Sprintf("Deleted dir %s", p))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Could not delete dir %s: %v", p, err))
	}

	return nil
}

func (c *clientHandler) handleCDUP() error {
	parent, _ := path.Split(c.Path())
	if parent != "/" && strings.HasSuffix(parent, "/") {
		parent = parent[0 : len(parent)-1]
	}

	if err := c.driver.ChangeDirectory(c, parent); err == nil {
		c.SetPath(parent)
		c.writeMessage(StatusFileOK, fmt.Sprintf("CDUP worked on %s", parent))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("CDUP issue: %v", err))
	}

	return nil
}

func (c *clientHandler) handlePWD() error {
	c.writeMessage(StatusPathCreated, "\""+c.Path()+"\" is the current directory")
	return nil
}

func (c *clientHandler) handleLIST() error {
	if files, err := c.driver.ListFiles(c, c.absPath(c.param)); err == nil {
		if tr, errTr := c.TransferOpen(); errTr == nil {
			defer c.TransferClose()
			return c.dirTransferLIST(tr, files)
		}
	} else {
		c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Could not list: %v", err))
	}

	return nil
}

func (c *clientHandler) handleNLST() error {
	if files, err := c.driver.ListFiles(c, c.absPath(c.param)); err == nil {
		if tr, errTrOpen := c.TransferOpen(); errTrOpen == nil {
			defer c.TransferClose()
			return c.dirTransferNLST(tr, files)
		}
	} else {
		c.writeMessage(500, fmt.Sprintf("Could not list: %v", err))
	}

	return nil
}

func (c *clientHandler) dirTransferNLST(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		fmt.Fprintf(w, "%s\r\n", file.Name())
	}

	return nil
}

func (c *clientHandler) handleMLSD() error {
	if c.server.settings.DisableMLSD {
		c.writeMessage(StatusSyntaxErrorNotRecognised, "MLSD has been disabled")
		return nil
	}

	if files, err := c.driver.ListFiles(c, c.absPath(c.param)); err == nil {
		if tr, errTr := c.TransferOpen(); errTr == nil {
			defer c.TransferClose()
			return c.dirTransferMLSD(tr, files)
		}
	} else {
		c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Could not list: %v", err))
	}

	return nil
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

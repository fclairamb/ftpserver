package server

import (
	"os"
	"fmt"
	"io"
	"net"
)

func (c *clientHandler) handleSTOR() {
	c.handleStoreAndAppend(false)
}

func (c *clientHandler) handleAPPE() {
	c.handleStoreAndAppend(true)
}

// Handles both the "STOR" and "APPE" commands
func (c *clientHandler) handleStoreAndAppend(append bool) {

	path := c.absPath(c.param)

	if tr, err := c.TransferOpen(); err == nil {
		defer c.TransferClose()
		if _, err := c.storeOrAppend(tr, path, append); err != nil && err != io.EOF {
			c.writeMessage(550, err.Error())
		}
	} else {
		c.writeMessage(550, err.Error())
	}
}

func (c *clientHandler) handleRETR() {

	path := c.absPath(c.param)

	if tr, err := c.TransferOpen(); err == nil {
		defer c.TransferClose()
		if _, err := c.download(tr, path); err != nil && err != io.EOF {
			c.writeMessage(550, err.Error())
		}
	} else {
		c.writeMessage(550, err.Error())
	}
}

func (c *clientHandler) download(conn net.Conn, name string) (int64, error) {
	if file, err := c.driver.OpenFile(c, name, os.O_RDONLY); err == nil {
		defer file.Close()
		return io.Copy(conn, file)
	} else {
		return 0, err
	}
}

func (c *clientHandler) storeOrAppend(conn net.Conn, name string, append bool) (int64, error) {
	flag := os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	}

	if file, err := c.driver.OpenFile(c, name, flag); err == nil {
		defer file.Close()
		return io.Copy(file, conn)
	} else {
		return 0, err
	}
}

func (c *clientHandler) handleDELE() {
	path := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Removed file %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't delete %s: %v", path, err))
	}
}

func (c *clientHandler) handleRNFR() {
	path := c.absPath(c.param)
	if _, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(350, "Sure, give me a target")
		c.UserInfo()["rnfr"] = path
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}
}

func (c *clientHandler) handleRNTO() {
	dst := c.absPath(c.param)
	if src := c.UserInfo()["rnfr"]; src != "" {
		if err := c.driver.RenameFile(c, src, dst); err == nil {
			c.writeMessage(250, "Done !")
			delete(c.UserInfo(), "rnfr")
		} else {
			c.writeMessage(550, fmt.Sprintf("Couldn't rename %s to %s: %s", src, dst, err.Error()))
		}
	}
}

func (c *clientHandler) handleSIZE() {
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(213, fmt.Sprintf("%d", info.Size()))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}
}

func (c *clientHandler) handleMDTM() {
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(250, info.ModTime().UTC().Format("20060102150405"))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}
}

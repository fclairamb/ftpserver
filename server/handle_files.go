package server

import (
	"os"
	"fmt"
	"io"
	"net"
)

func (c *ClientHandler) handleSTOR() {
	c.handleStoreAndAppend(false)
}

func (c *ClientHandler) handleAPPE() {
	c.handleStoreAndAppend(true)
}

// Handles both the "STOR" and "APPE" commands
func (c *ClientHandler) handleStoreAndAppend(append bool) {

	path := c.absPath(c.param)

	if tr, err := c.TransferOpen(); err == nil {
		defer c.TransferClose()
		if total, err := c.storeOrAppend(tr, path, append); err == nil || err == io.EOF {
			c.writeMessage(226, fmt.Sprintf("OK, received %d bytes", total))
		} else {
			c.writeMessage(550, err.Error())
		}
	} else {
		c.writeMessage(550, err.Error())
	}
}

func (c *ClientHandler) handleRETR() {

	path := c.absPath(c.param)

	if tr, err := c.TransferOpen(); err == nil {
		defer c.TransferClose()
		if total, err := c.download(tr, path); err == nil || err == io.EOF {
			c.writeMessage(226, fmt.Sprintf("OK, sent %d bytes", total))
		} else {
			c.writeMessage(550, err.Error())
		}
	} else {
		c.writeMessage(550, err.Error())
	}
}

func (c *ClientHandler) download(conn net.Conn, name string) (int64, error) {
	if file, err := c.daddy.driver.OpenFile(c, name, os.O_RDONLY); err == nil {
		defer file.Close()
		return io.Copy(conn, file)
	} else {
		return 0, err
	}
}

func (c *ClientHandler) storeOrAppend(conn net.Conn, name string, append bool) (int64, error) {
	flag := os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	}

	if file, err := c.daddy.driver.OpenFile(c, name, flag); err == nil {
		defer file.Close()
		return io.Copy(file, conn)
	} else {
		return 0, err
	}
}

func (c *ClientHandler) handleDELE() {
	path := c.absPath(c.param)
	if err := c.daddy.driver.DeleteFile(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Removed file %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't delete %s: %s", path, err.Error()))
	}
}

func (c *ClientHandler) handleRNFR() {
	path := c.absPath(c.param)
	if _, err := c.daddy.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(250, "Sure, give me a target")
		c.UserInfo()["rnfr"] = path
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}
}

func (c *ClientHandler) handleRNTO() {
	dst := c.absPath(c.param)
	if src := c.UserInfo()["rnfr"]; src != "" {
		if err := c.daddy.driver.RenameFile(c, src, dst); err == nil {
			c.writeMessage(250, "Done !")
			delete(c.UserInfo(), "rnfr")
		} else {
			c.writeMessage(550, fmt.Sprintf("Couldn't rename %s to %s: %s", src, dst, err.Error()))
		}
	}
}

func (c *ClientHandler) handleSIZE() {
	path := c.absPath(c.param)
	if info, err := c.daddy.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(213, fmt.Sprintf("%d", info.Size()))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}
}

func (c *ClientHandler) handleMDTM() {
	path := c.absPath(c.param)
	if info, err := c.daddy.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(250, info.ModTime().UTC().Format("20060102150405"))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}
}

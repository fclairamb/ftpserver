package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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
		if c.ctx_rest != 0 {
			file.Seek(c.ctx_rest, 0)
			c.ctx_rest = 0
		}
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
		if c.ctx_rest != 0 {
			file.Seek(c.ctx_rest, 0)
			c.ctx_rest = 0
		}
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
		c.ctx_rnfr = path
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}
}

func (c *clientHandler) handleRNTO() {
	dst := c.absPath(c.param)
	if c.ctx_rnfr != "" {
		if err := c.driver.RenameFile(c, c.ctx_rnfr, dst); err == nil {
			c.writeMessage(250, "Done !")
			c.ctx_rnfr = ""
		} else {
			c.writeMessage(550, fmt.Sprintf("Couldn't rename %s to %s: %s", c.ctx_rnfr, dst, err.Error()))
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

func (c *clientHandler) handleALLO() {
	// We should probably add a method in the driver
	if size, err := strconv.Atoi(c.param); err == nil {
		if ok, err := c.driver.CanAllocate(c, size); err == nil {
			if ok {
				c.writeMessage(202, "OK, we have the free space")
			} else {
				c.writeMessage(550, "NOT OK, we don't have the free space")
			}
		} else {
			c.writeMessage(500, fmt.Sprintf("Driver issue: %v", err))
		}
	} else {
		c.writeMessage(501, fmt.Sprintf("Couldn't parse size: %v", err))
	}
}

func (c *clientHandler) handleREST() {
	if size, err := strconv.ParseInt(c.param, 10, 0); err == nil {
		c.ctx_rest = size
		c.writeMessage(350, "OK")
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't parse size: %v", err))
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

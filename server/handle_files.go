package server

import "io"
import (
	"time"
	"os"
	"fmt"
)

func (c *ClientHandler) HandleStore() {
	c.handleStoreAndAppend(false)
}

func (c *ClientHandler) HandleAppend() {
	c.handleStoreAndAppend(true)
}

// Handles both the "STOR" and "APPE" commands
// TODO: Fix this passive connection handling, it is overly complex and strange
func (c *ClientHandler) handleStoreAndAppend(append bool) {
	passive := c.lastPassive()
	if passive == nil {
		return
	}
	defer c.closePassive(passive)

	c.writeMessage(150, "Data transfer starting")
	if waitTimeout(&passive.waiter, time.Minute) {
		c.writeMessage(550, "Could not get passive connection.")
		return
	}
	if passive.listenFailedAt > 0 {
		c.writeMessage(550, "Could not get passive connection.")
		return
	}

	path := c.absPath(c.param)

	if total, err := c.storeOrAppend(passive, path, append); err == nil {
		c.writeMessage(226, fmt.Sprintf("OK, received %d bytes", total))
	} else {
		c.writeMessage(550, "Error with upload: " + err.Error())
	}
}

func (c *ClientHandler) HandleRetr() {
	passive := c.lastPassive()
	if passive == nil {
		return
	}
	defer c.closePassive(passive)

	c.writeMessage(150, "Data transfer starting")
	if waitTimeout(&passive.waiter, time.Minute) {
		c.writeMessage(550, "Could not get passive connection.")
		return
	}
	if passive.listenFailedAt > 0 {
		c.writeMessage(550, "Could not get passive connection.")
		return
	}

	path := c.absPath(c.param)

	if total, err := c.download(passive, path); err == nil {
		c.writeMessage(226, fmt.Sprintf("OK, sent %d bytes", total))
	} else {
		c.writeMessage(551, "Error with download: " + err.Error())
	}
}

func (c *ClientHandler) download(passive *Passive, name string) (int64, error) {
	if file, err := c.daddy.driver.OpenFile(c, name, os.O_RDONLY); err == nil {
		defer file.Close()
		return io.Copy(passive.connection, file)
	} else {
		return 0, err
	}
}

func (c *ClientHandler) storeOrAppend(passive *Passive, name string, append bool) (int64, error) {
	flag := os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	}

	if file, err := c.daddy.driver.OpenFile(c, name, flag); err == nil {
		defer file.Close()
		// We copy 512 bytes for type identification
		if first, err := io.CopyN(file, passive.connection, 512); err == nil {
			// And then everything else
			total, err := io.Copy(file, passive.connection)
			total += first
			return total, err
		} else {
			return first, err
		}
	} else {
		return 0, err
	}
}

func (c *ClientHandler) HandleDele() {
	path := c.absPath(c.param)
	if err := c.daddy.driver.DeleteFile(c, path); err == nil {
		c.SetPath(path)
		c.writeMessage(250, fmt.Sprintf("Removed file %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't delete %s: %s", path, err.Error()))
	}
}

package server

import "io"
import (
	"time"
	"os"
	"fmt"
	"gopkg.in/inconshreveable/log15.v2"
)

func (c *ClientHandler) HandleStore() {
	c.handleStoreAndAppend(false)
}

func (c *ClientHandler) HandleAppend() {
	c.handleStoreAndAppend(true)
}

// Handles both the "STOR" and "APPE" commands
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

	name := c.Path() + "/" + c.param


	if total, err := c.storeOrAppend(passive, append, name); err == nil {
		c.writeMessage(226, fmt.Sprintf("OK, received %d bytes", total))
	} else {
		c.writeMessage(550, "Error with upload: "+err.Error())
	}
}

func (c *ClientHandler) storeOrAppend(passive *Passive, append bool, name string) (int64, error) {
	var err error

	flag := 0

	if append {
		flag |= os.O_APPEND
	}

	file, err := c.daddy.driver.StartFileUpload(c, name, flag)

	if err != nil {
		return 0, err
	}
	defer file.Close()

	total := int64(0)
	n := 0
	bytesToRead := 512 // We read 512B and then 4MB
	for {
		temp_buffer := make([]byte, bytesToRead)
		n, err = passive.connection.Read(temp_buffer)
		total += int64(n)

		if err != nil {
			log15.Error("Error while reading", "err", err)
			break
		}

		_, err := file.Write(temp_buffer[0:n])

		if err != nil {
			log15.Error("Error while writing", "err", err)
			break
		}

		bytesToRead = 4 * 1024 * 1024
	}

	if err == io.EOF {
		return total, nil
	} else {
		return total, err
	}
}

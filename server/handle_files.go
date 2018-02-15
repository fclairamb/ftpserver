package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func (c *clientHandler) handleSTOR() {
	c.transferFile(true, false)
}

func (c *clientHandler) handleAPPE() {
	c.transferFile(true, true)
}

func (c *clientHandler) handleRETR() {
	c.transferFile(false, false)
}

// File transfer, read or write, seek or not, is basically the same.
// To make sure we don't miss any step, we execute everything in order
func (c *clientHandler) transferFile(write bool, append bool) {

	var file FileStream
	var err error

	// We try to open the file
	{
		var fileFlag int
		if write {
			fileFlag = os.O_WRONLY
			if append {
				fileFlag |= os.O_APPEND
			}
		} else {
			fileFlag = os.O_RDONLY
		}

		// If this fail, can stop right here
		if file, err = c.driver.OpenFile(c, c.absPath(c.param), fileFlag); err != nil {
			c.writeMessage(550, "Could not access file: "+err.Error())
			return
		}
	}

	// Try to seek on it
	if c.ctxRest != 0 {
		if err == nil {
			if _, errSeek := file.Seek(c.ctxRest, 0); errSeek != nil {
				err = errSeek
			}
		}

		// Whatever happens we should reset the seek position
		c.ctxRest = 0
	}

	// Start the transfer
	if err == nil {
		var tr net.Conn
		if tr, err = c.TransferOpen(); err == nil {
			defer c.TransferClose()

			// Copy the data
			var in io.Reader
			var out io.Writer

			if write { // ... from the connection to the file
				in = tr
				out = file
			} else { // ... from the file to the connection
				in = file
				out = tr
			}

			if _, errCopy := io.Copy(out, in); errCopy != nil && errCopy != io.EOF {
				err = errCopy
			}
		}
	}

	// *ALWAYS* close the file but only save the error if there wasn't one before
	// Note: We could discard the error in read mode
	if errClose := file.Close(); errClose != nil && err == nil {
		err = errClose
	}

	if err != nil {
		c.writeMessage(550, "Could not transfer file: "+err.Error())
		return
	}
}

func (c *clientHandler) handleCHMOD(params string) {
	spl := strings.SplitN(params, " ", 2)
	modeNb, err := strconv.ParseUint(spl[0], 10, 32)

	mode := os.FileMode(modeNb)
	path := c.absPath(spl[1])

	if err == nil {
		err = c.driver.ChmodFile(c, path, mode)
	}

	if err != nil {
		c.writeMessage(550, err.Error())
		return
	}

	c.writeMessage(200, "SITE CHMOD command successful")
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
		c.ctxRnfr = path
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}
}

func (c *clientHandler) handleRNTO() {
	dst := c.absPath(c.param)
	if c.ctxRnfr != "" {
		if err := c.driver.RenameFile(c, c.ctxRnfr, dst); err == nil {
			c.writeMessage(250, "Done !")
			c.ctxRnfr = ""
		} else {
			c.writeMessage(550, fmt.Sprintf("Couldn't rename %s to %s: %s", c.ctxRnfr, dst, err.Error()))
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

func (c *clientHandler) handleSTATFile() {
	path := c.absPath(c.param)

	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeLine("211-Status follows:")
		if info.IsDir() {
			if files, err := c.driver.ListFiles(c); err == nil {
				for _, f := range files {
					c.writeLine(fmt.Sprintf(" %s", c.fileStat(f)))
				}
			}
		} else {
			c.writeLine(fmt.Sprintf(" %s", c.fileStat(info)))
		}
		c.writeLine("211 End of status")
	} else {
		c.writeMessage(450, fmt.Sprintf("Could not STAT: %v", err))
	}
}

func (c *clientHandler) handleMLST() {
	if c.daddy.settings.DisableMLST {
		c.writeMessage(500, "MLST has been disabled")
		return
	}
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writer.Write([]byte("250- File details\r\n "))
		c.writeMLSxOutput(c.writer, info)
		c.writeMessage(250, "End of file details")
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not list: %v", err))
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
		c.ctxRest = size
		c.writeMessage(350, "OK")
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't parse size: %v", err))
	}
}

func (c *clientHandler) handleMDTM() {
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(250, info.ModTime().UTC().Format(dateFormatMLSD))
	} else {
		c.writeMessage(550, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}
}

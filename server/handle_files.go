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
	c.handleStoreAndAppend(false)
}

func (c *clientHandler) handleAPPE() {
	c.handleStoreAndAppend(true)
}

// Handles both the "STOR" and "APPE" commands
func (c *clientHandler) handleStoreAndAppend(append bool) {
	if err := c.storeOrAppend(append); err != nil {
		c.writeMessage(550, err.Error())
	}
}

func (c *clientHandler) storeOrAppend(append bool) (err error) {
	file, err := c.openFile(c.absPath(c.param), append)
	if err != nil {
		return fmt.Errorf("Could not open file: %s", err)
	}

	defer func() {
		if errClose := file.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()

	tr, err := c.TransferOpen()
	if err != nil {
		return fmt.Errorf("Could not open transfer: %s", err)
	}
	defer c.TransferClose()

	if c.ctxRest != 0 {
		file.Seek(c.ctxRest, 0)
		c.ctxRest = 0
	}

	if _, err := io.Copy(file, tr); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (c *clientHandler) openFile(path string, append bool) (FileStream, error) {
	flag := os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	}

	return c.driver.OpenFile(c, path, flag)
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
	file, err := c.driver.OpenFile(c, name, os.O_RDONLY)

	if err != nil {
		return 0, err
	}

	if c.ctxRest != 0 {
		file.Seek(c.ctxRest, 0)
		c.ctxRest = 0
	}

	defer file.Close()
	return io.Copy(conn, file)
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

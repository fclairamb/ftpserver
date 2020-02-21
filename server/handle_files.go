// Package server provides all the tools to build your own FTP server: The core library and the driver.
package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func (c *clientHandler) handleSTOR() error {
	c.transferFile(true, false)
	return nil
}

func (c *clientHandler) handleAPPE() error {
	c.transferFile(true, true)
	return nil
}

func (c *clientHandler) handleRETR() error {
	c.transferFile(false, false)
	return nil
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
			c.writeMessage(StatusActionNotTaken, "Could not access file: "+err.Error())
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
	if errClose := file.Close(); errClose != nil && err == nil {
		err = errClose
	}

	if err != nil {
		c.writeMessage(StatusActionNotTaken, "Could not transfer file: "+err.Error())
		return
	}
}

func (c *clientHandler) handleCHMOD(params string) {
	spl := strings.SplitN(params, " ", 2)
	modeNb, err := strconv.ParseUint(spl[0], 8, 32)

	mode := os.FileMode(modeNb)
	path := c.absPath(spl[1])

	if err == nil {
		err = c.driver.ChmodFile(c, path, mode)
	}

	if err != nil {
		c.writeMessage(StatusActionNotTaken, err.Error())
		return
	}

	c.writeMessage(StatusOK, "SITE CHMOD command successful")
}

func (c *clientHandler) handleDELE() error {
	path := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, path); err == nil {
		c.writeMessage(StatusFileOK, fmt.Sprintf("Removed file %s", path))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't delete %s: %v", path, err))
	}

	return nil
}

func (c *clientHandler) handleRNFR() error {
	path := c.absPath(c.param)
	if _, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(StatusFileActionPending, "Sure, give me a target")
		c.ctxRnfr = path
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}

	return nil
}

func (c *clientHandler) handleRNTO() error {
	dst := c.absPath(c.param)

	if c.ctxRnfr != "" {
		if err := c.driver.RenameFile(c, c.ctxRnfr, dst); err == nil {
			c.writeMessage(StatusFileOK, "Done !")
			c.ctxRnfr = ""
		} else {
			c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't rename %s to %s: %s", c.ctxRnfr, dst, err.Error()))
		}
	}

	return nil
}

func (c *clientHandler) handleSIZE() error {
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(StatusFileStatus, fmt.Sprintf("%d", info.Size()))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %v", path, err))
	}

	return nil
}

func (c *clientHandler) handleSTATFile() error {
	path := c.absPath(c.param)

	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		m := c.multilineAnswer(StatusSystemStatus, "System status")
		defer m()
		// c.writeLine(fmt.Sprintf("%d-Status follows:", StatusSystemStatus))
		if info.IsDir() {
			if files, errList := c.driver.ListFiles(c, c.absPath(c.param)); errList == nil {
				for _, f := range files {
					c.writeLine(fmt.Sprintf(" %s", c.fileStat(f)))
				}
			}
		} else {
			c.writeLine(fmt.Sprintf(" %s", c.fileStat(info)))
		}
	} else {
		c.writeMessage(StatusFileActionNotTaken, fmt.Sprintf("Could not STAT: %v", err))
	}

	return nil
}

func (c *clientHandler) handleMLST() error {
	if c.server.settings.DisableMLST {
		c.writeMessage(StatusSyntaxErrorNotRecognised, "MLST has been disabled")
		return nil
	}

	path := c.absPath(c.param)

	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		m := c.multilineAnswer(StatusFileOK, "File details")
		defer m()

		c.writeMLSxOutput(c.writer, info)
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Could not list: %v", err))
	}

	return nil
}

func (c *clientHandler) handleALLO() error {
	// We should probably add a method in the driver
	if size, err := strconv.Atoi(c.param); err == nil {
		if ok, err2 := c.driver.CanAllocate(c, size); err2 == nil {
			if ok {
				c.writeMessage(StatusNotImplemented, "OK, we have the free space")
			} else {
				c.writeMessage(StatusActionNotTaken, "NOT OK, we don't have the free space")
			}
		} else {
			c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Driver issue: %v", err2))
		}
	} else {
		c.writeMessage(StatusSyntaxErrorParameters, fmt.Sprintf("Couldn't parse size: %v", err))
	}

	return nil
}

func (c *clientHandler) handleREST() error {
	if size, err := strconv.ParseInt(c.param, 10, 0); err == nil {
		c.ctxRest = size
		c.writeMessage(StatusFileActionPending, "OK")
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't parse size: %v", err))
	}

	return nil
}

func (c *clientHandler) handleMDTM() error {
	path := c.absPath(c.param)
	if info, err := c.driver.GetFileInfo(c, path); err == nil {
		c.writeMessage(StatusFileStatus, info.ModTime().UTC().Format(dateFormatMLSD))
	} else {
		c.writeMessage(StatusActionNotTaken, fmt.Sprintf("Couldn't access %s: %s", path, err.Error()))
	}

	return nil
}

// RFC draft: https://tools.ietf.org/html/draft-somers-ftp-mfxx-04#section-3.1
func (c *clientHandler) handleMFMT() error {
	params := strings.SplitN(c.param, " ", 2)
	if len(params) != 2 {
		c.writeMessage(StatusSyntaxErrorNotRecognised, fmt.Sprintf(
			"Couldn't set mtime, not enough params, given: %s", c.param))
	}

	mtime, err := time.Parse("20060102150405", params[0])
	if err != nil {
		c.writeMessage(StatusSyntaxErrorParameters, fmt.Sprintf(
			"Couldn't parse mtime, given: %s, err: %v", params[0], err))
	}

	if err := c.driver.SetFileMtime(c, params[1], mtime); err != nil {
		c.writeMessage(StatusSyntaxErrorParameters, fmt.Sprintf(
			"Couldn't set mtime %q for %q, err: %v", mtime.Format(time.RFC3339), params[0], err))
	}

	c.writeMessage(StatusFileStatus, fmt.Sprintf("Modify=%s; %s", params[0], params[1]))

	return nil
}

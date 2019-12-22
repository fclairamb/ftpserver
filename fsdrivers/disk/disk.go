// Package disk provides access to local files on the disk
package disk

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/fclairamb/ftpserver/server"
)

// Driver provides an implementation of driver for disk access
type Driver struct {
	baseDir string
}

// ChangeDirectory changes the current working directory
func (driver *Driver) ChangeDirectory(cc server.ClientContext, directory string) error {
	if strings.HasPrefix(directory, "/root") {
		return errors.New("this doesn't look good")
	} else if directory == "/virtual" {
		return nil
	}

	_, err := os.Stat(driver.baseDir + directory)

	return err
}

// MakeDirectory creates a directory
func (driver *Driver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0750)
}

// ListFiles lists the files of a directory
func (driver *Driver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {
	return ioutil.ReadDir(driver.baseDir + cc.Path())
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *Driver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	path = filepath.Join(driver.baseDir, path)

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			if _, errStat := os.Stat(path); errStat == nil {
				if errRemove := os.Remove(path); errRemove != nil {
					log15.Warn(
						"Could not remove file",
						"path", path,
						"err", errRemove,
					)
				}
			} else if !os.IsNotExist(errStat) {
				log15.Error("We had an error checking for a file",
					"path", path,
					"err", errStat,
				)
			}
		}
	}

	return os.OpenFile(path, flag, 0600)
}

// CanAllocate gives the approval to allocate some data
func (driver *Driver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

// GetFileInfo gets some info around a file or a directory
func (driver *Driver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = driver.baseDir + path

	return os.Stat(path)
}

// DeleteFile deletes a file or a directory
func (driver *Driver) DeleteFile(cc server.ClientContext, path string) error {
	return os.Remove(filepath.Join(driver.baseDir, path))
}

// RenameFile renames a file or a directory
func (driver *Driver) RenameFile(cc server.ClientContext, from, to string) error {
	return os.Rename(
		filepath.Join(driver.baseDir, from),
		filepath.Join(driver.baseDir, to),
	)
}

// ChmodFile changes the attributes of the file
func (driver *Driver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	path = driver.baseDir + path

	return os.Chmod(path, mode)
}

// NewDriver creates a new instance on a particular directory
func NewDriver(directory string) (*Driver, error) {
	return &Driver{baseDir: directory}, nil
}

// NewDriverTemp creates a new instance of this on a temporary directory
func NewDriverTemp() *Driver {
	dir := "/tmp/ftpisback"

	if errStat := os.MkdirAll(dir, 0750); errStat != nil {
		log15.Info("Couldn't get our preferred dir", "dir", dir, "err", errStat)
		dir, errStat = ioutil.TempDir("", "ftpserver")

		if errStat != nil {
			log15.Error("Could not find a temporary dir", "err", errStat)
		}
	}

	driver := &Driver{
		baseDir: dir,
	}

	return driver
}

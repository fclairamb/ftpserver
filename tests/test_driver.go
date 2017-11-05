package tests

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"os"

	"github.com/fclairamb/ftpserver/server"
)

// NewTestServer provides a test server with or without debugging
func NewTestServer(debug bool) *server.FtpServer {
	return NewTestServerWithDriver(&ServerDriver{Debug: debug})
}

// NewTestServerWithDriver provides a server instantiated with some settings
func NewTestServerWithDriver(driver *ServerDriver) *server.FtpServer {
	if driver.Settings == nil {
		driver.Settings = &server.Settings{}
	}

	if driver.Settings.ListenAddr == "" {
		driver.Settings.ListenAddr = "127.0.0.1:0"
	}

	s := server.NewFtpServer(driver)
	if err := s.Listen(); err != nil {
		return nil
	}
	go s.Serve()
	return s
}

// ServerDriver defines a minimal serverftp server driver
type ServerDriver struct {
	Debug    bool             // To display connection logs information
	Settings *server.Settings // Settings
}

// ClientDriver defines a minimal serverftp client driver
type ClientDriver struct {
	baseDir string
}

// NewClientDriver creates a client driver
func NewClientDriver() *ClientDriver {
	dir, _ := ioutil.TempDir("", "example")
	os.MkdirAll(dir, 0777)
	return &ClientDriver{baseDir: dir}
}

// WelcomeUser is the very first message people will see
func (driver *ServerDriver) WelcomeUser(cc server.ClientContext) (string, error) {
	cc.SetDebug(driver.Debug)
	// This will remain the official name for now
	return "TEST Server", nil
}

// AuthUser with authenticate users
func (driver *ServerDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {
	if user == "test" && pass == "test" {
		return NewClientDriver(), nil
	}
	return nil, errors.New("bad username or password")
}

// UserLeft is called when the user disconnects
func (driver *ServerDriver) UserLeft(cc server.ClientContext) {

}

// GetSettings fetches the basic server settings
func (driver *ServerDriver) GetSettings() (*server.Settings, error) {
	return driver.Settings, nil
}

// GetTLSConfig fetches the TLS config
func (driver *ServerDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, nil
}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	_, err := os.Stat(driver.baseDir + directory)
	return err
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0777)
}

// ListFiles lists the files of a directory
func (driver *ClientDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {
	path := driver.baseDir + cc.Path()
	files, err := ioutil.ReadDir(path)
	return files, err
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	path = driver.baseDir + path

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			os.Remove(path)
		}
	}

	return os.OpenFile(path, flag, 0666)
}

// GetFileInfo gets some info around a file or a directory
func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = driver.baseDir + path

	return os.Stat(path)
}

// CanAllocate gives the approval to allocate some data
func (driver *ClientDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

// ChmodFile changes the attributes of the file
func (driver *ClientDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	path = driver.baseDir + path
	return os.Chmod(path, mode)
}

// DeleteFile deletes a file or a directory
func (driver *ClientDriver) DeleteFile(cc server.ClientContext, path string) error {
	path = driver.baseDir + path
	return os.Remove(path)
}

// RenameFile renames a file or a directory
func (driver *ClientDriver) RenameFile(cc server.ClientContext, from, to string) error {
	from = driver.baseDir + from
	to = driver.baseDir + to
	return os.Rename(from, to)
}

package tests

import (
	"crypto/tls"
	"errors"
	"github.com/fclairamb/ftpserver/server"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func getServer(debug bool) *server.FtpServer {
	s := server.NewFtpServer(NewServerDriver(debug))
	if err := s.Listen(); err != nil {
		return nil
	}
	go s.Serve()
	return s
}

// NewServerDriver creates a server driver
func NewServerDriver(debug bool) *ServerDriver {
	return &ServerDriver{Debug: debug}
}

// ServerDriver defines a minimal serverftp server driver
type ServerDriver struct {
	Debug bool // To display connection logs information
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
	return nil, errors.New("Bad username or password")
}

// UserLeft is called when the user disconnects
func (driver *ServerDriver) UserLeft(cc server.ClientContext) {

}

// GetSettings fetches the basic server settings
func (driver *ServerDriver) GetSettings() *server.Settings {
	return &server.Settings{ListenHost: "127.0.0.1", ListenPort: -1}
}

// GetTLSConfig fetches the TLS config
func (driver *ServerDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, nil
}

func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	_, err := os.Stat(driver.baseDir + directory)
	return err
}

func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0777)
}

func (driver *ClientDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {
	path := driver.baseDir + cc.Path()
	files, err := ioutil.ReadDir(path)
	return files, err
}

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

func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = driver.baseDir + path

	return os.Stat(path)
}

func (driver *ClientDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

func (driver *ClientDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	path = driver.baseDir + path
	return os.Chmod(path, mode)
}

func (driver *ClientDriver) DeleteFile(cc server.ClientContext, path string) error {
	path = driver.baseDir + path
	return os.Remove(path)
}

func (driver *ClientDriver) RenameFile(cc server.ClientContext, from, to string) error {
	from = driver.baseDir + from
	to = driver.baseDir + to
	return os.Rename(from, to)
}

type VirtualFile struct {
	content    []byte // Content of the file
	readOffset int    // Reading offset
}

func (f *VirtualFile) Close() error {
	return nil
}

func (f *VirtualFile) Read(buffer []byte) (int, error) {
	n := copy(buffer, f.content[f.readOffset:])
	f.readOffset += n
	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func (f *VirtualFile) Seek(n int64, w int) (int64, error) {
	return 0, nil
}

func (f *VirtualFile) Write(buffer []byte) (int, error) {
	return 0, nil
}

type VirtualFileInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (f VirtualFileInfo) Name() string {
	return f.name
}

func (f VirtualFileInfo) Size() int64 {
	return f.size
}

func (f VirtualFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f VirtualFileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f VirtualFileInfo) ModTime() time.Time {
	return time.Now().UTC()
}

func (f VirtualFileInfo) Sys() interface{} {
	return nil
}

// Package sample is a sample server driver
package sample

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"sync/atomic"

	"github.com/fclairamb/ftpserver/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/naoina/toml"
)

// MainDriver defines a very basic ftpserver driver
type MainDriver struct {
	Logger       log.Logger  // Logger
	SettingsFile string      // Settings file
	BaseDir      string      // Base directory from which to serve file
	tlsConfig    *tls.Config // TLS config (if applies)
	config       OurSettings // Our settings
	nbClients    int32       // Number of clients
}

// ClientDriver defines a very basic client driver
type ClientDriver struct {
	BaseDir string // Base directory from which to server file
}

// Account defines a user/pass password
type Account struct {
	User string // Username
	Pass string // Password
	Dir  string // Directory
}

// OurSettings defines our settings
type OurSettings struct {
	Server         server.Settings // Server settings (shouldn't need to be filled)
	Users          []Account       // Credentials
	MaxConnections int32           // Maximum number of clients that are allowed to connect at the same time
}

// GetSettings returns some general settings around the server setup
func (driver *MainDriver) GetSettings() (*server.Settings, error) {
	f, err := os.Open(driver.SettingsFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	//var config OurSettings
	if err := toml.Unmarshal(buf, &driver.config); err != nil {
		return nil, fmt.Errorf("problem loading \"%s\": %v", driver.SettingsFile, err)
	}

	// This is the new IP loading change coming from Ray
	if driver.config.Server.PublicHost == "" {
		level.Debug(driver.Logger).Log("msg", "Fetching our external IP address...")
		if driver.config.Server.PublicHost, err = externalIP(); err != nil {
			level.Warn(driver.Logger).Log("msg", "Couldn't fetch an external IP", "err", err)
		} else {
			level.Debug(driver.Logger).Log("msg", "Fetched our external IP address", "ipAddress", driver.config.Server.PublicHost)
		}
	}

	if len(driver.config.Users) == 0 {
		return nil, errors.New("you must have at least one user defined")
	}

	return &driver.config.Server, nil
}

// GetTLSConfig returns a TLS Certificate to use
func (driver *MainDriver) GetTLSConfig() (*tls.Config, error) {
	if driver.tlsConfig == nil {
		level.Info(driver.Logger).Log("msg", "Loading certificate")
		if cert, err := tls.LoadX509KeyPair("sample/certs/mycert.crt", "sample/certs/mycert.key"); err == nil {
			driver.tlsConfig = &tls.Config{
				NextProtos:   []string{"ftp"},
				Certificates: []tls.Certificate{cert},
			}
		} else {
			return nil, err
		}
	}
	return driver.tlsConfig, nil
}

// WelcomeUser is called to send the very first welcome message
func (driver *MainDriver) WelcomeUser(cc server.ClientContext) (string, error) {
	nbClients := atomic.AddInt32(&driver.nbClients, 1)
	if nbClients > driver.config.MaxConnections {
		return "Cannot accept any additional client", fmt.Errorf("too many clients: %d > % d", driver.nbClients, driver.config.MaxConnections)
	}

	cc.SetDebug(true)
	// This will remain the official name for now
	return fmt.Sprintf(
			"Welcome on ftpserver, you're on dir %s, your ID is %d, your IP:port is %s, we currently have %d clients connected",
			driver.BaseDir,
			cc.ID(),
			cc.RemoteAddr(),
			nbClients),
		nil
}

// AuthUser authenticates the user and selects an handling driver
func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {

	for _, act := range driver.config.Users {
		if act.User == user && act.Pass == pass {
			// If we are authenticated, we can return a client driver containing *our* basedir
			baseDir := driver.BaseDir + string(os.PathSeparator) + act.Dir
			os.MkdirAll(baseDir, 0777)
			return &ClientDriver{BaseDir: baseDir}, nil
		}
	}

	return nil, fmt.Errorf("could not authenticate you")
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {
	atomic.AddInt32(&driver.nbClients, -1)
}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	if directory == "/debug" {
		cc.SetDebug(!cc.Debug())
		return nil
	} else if directory == "/virtual" {
		return nil
	}
	_, err := os.Stat(driver.BaseDir + directory)
	return err
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.BaseDir+directory, 0777)
}

// ListFiles lists the files of a directory
func (driver *ClientDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {

	if cc.Path() == "/virtual" {
		files := make([]os.FileInfo, 0)
		files = append(files,
			virtualFileInfo{
				name: "localpath.txt",
				mode: os.FileMode(0666),
				size: 1024,
			},
			virtualFileInfo{
				name: "file2.txt",
				mode: os.FileMode(0666),
				size: 2048,
			},
		)
		return files, nil
	}

	path := driver.BaseDir + cc.Path()

	files, err := ioutil.ReadDir(path)

	// We add a virtual dir
	if cc.Path() == "/" && err == nil {
		files = append(files, virtualFileInfo{
			name: "virtual",
			mode: os.FileMode(0666) | os.ModeDir,
			size: 4096,
		})
	}

	return files, err
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {

	if path == "/virtual/localpath.txt" {
		return &virtualFile{content: []byte(driver.BaseDir)}, nil
	}

	path = driver.BaseDir + path

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
	path = driver.BaseDir + path

	return os.Stat(path)
}

// CanAllocate gives the approval to allocate some data
func (driver *ClientDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

// ChmodFile changes the attributes of the file
func (driver *ClientDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	path = driver.BaseDir + path

	return os.Chmod(path, mode)
}

// DeleteFile deletes a file or a directory
func (driver *ClientDriver) DeleteFile(cc server.ClientContext, path string) error {
	path = driver.BaseDir + path

	return os.Remove(path)
}

// RenameFile renames a file or a directory
func (driver *ClientDriver) RenameFile(cc server.ClientContext, from, to string) error {
	from = driver.BaseDir + from
	to = driver.BaseDir + to

	return os.Rename(from, to)
}

// NewSampleDriver creates a sample driver
func NewSampleDriver(dir string, settingsFile string) (*MainDriver, error) {
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "ftpserver")
		if err != nil {
			return nil, fmt.Errorf("could not find a temporary dir, err: %v", err)
		}
	}

	drv := &MainDriver{
		Logger:       log.NewNopLogger(),
		SettingsFile: settingsFile,
		BaseDir:      dir,
	}

	return drv, nil
}

// The virtual file is an example of how you can implement a purely virtual file
type virtualFile struct {
	content    []byte // Content of the file
	readOffset int    // Reading offset
}

func (f *virtualFile) Close() error {
	return nil
}

func (f *virtualFile) Read(buffer []byte) (int, error) {
	n := copy(buffer, f.content[f.readOffset:])
	f.readOffset += n
	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func (f *virtualFile) Seek(n int64, w int) (int64, error) {
	return 0, nil
}

func (f *virtualFile) Write(buffer []byte) (int, error) {
	return 0, nil
}

type virtualFileInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (f virtualFileInfo) Name() string {
	return f.name
}

func (f virtualFileInfo) Size() int64 {
	return f.size
}

func (f virtualFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f virtualFileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f virtualFileInfo) ModTime() time.Time {
	return time.Now().UTC()
}

func (f virtualFileInfo) Sys() interface{} {
	return nil
}

func externalIP() (string, error) {
	// If you need to take a bet, amazon is about as reliable & sustainable a service as you can get
	rsp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}

// Package tests brings all the logic to test the server without messing up the main code
package tests

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	gklog "github.com/go-kit/kit/log"

	"github.com/fclairamb/ftpserver/server"
	"github.com/fclairamb/ftpserver/server/log"
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

	// If we are in debug mode, we should log things
	if driver.Debug {
		s.Logger = log.NewGKLogger(gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stdout))).With(
			"ts", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	if err := s.Listen(); err != nil {
		return nil
	}

	go s.Serve()

	return s
}

// ServerDriver defines a minimal serverftp server driver
type ServerDriver struct {
	Debug bool // To display connection logs information
	TLS   bool

	Settings *server.Settings // Settings
	server.FileStream
}

// ClientDriver defines a minimal serverftp client driver
type ClientDriver struct {
	baseDir string
	server.FileStream
}

// NewClientDriver creates a client driver
func NewClientDriver() *ClientDriver {
	dir, _ := ioutil.TempDir("", "example")
	if err := os.MkdirAll(dir, 0750); err != nil {
		panic(err)
	}

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
		clientdriver := NewClientDriver()
		if driver.FileStream != nil {
			clientdriver.FileStream = driver.FileStream
		}

		return clientdriver, nil
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
	if driver.TLS {
		keypair, err := tls.X509KeyPair(localhostCert, localhostKey)
		if err != nil {
			return nil, err
		}

		return &tls.Config{Certificates: []tls.Certificate{keypair}}, nil
	}

	return nil, nil
}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	_, err := os.Stat(driver.baseDir + directory)
	return err
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0750)
}

// ListFiles lists the files of a directory
func (driver *ClientDriver) ListFiles(cc server.ClientContext, directory string) ([]os.FileInfo, error) {
	path := path.Join(driver.baseDir, directory)
	if directory == "" {
		path = driver.baseDir + cc.Path()
	}

	return ioutil.ReadDir(path)
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	path = driver.baseDir + path

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			if _, err := os.Stat(path); err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("error accessing file %s: %v", path, err)
				}
			} else {
				if err := os.Remove(path); err != nil {
					return nil, fmt.Errorf("error deleting %s: %v", path, err)
				}
			}
		}
	}

	if driver.FileStream != nil {
		return driver.FileStream, nil
	}

	return os.OpenFile(path, flag, 0600)
}

// GetFileInfo gets some info around a file or a directory
func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = driver.baseDir + path

	return os.Stat(path)
}

// SetFileMtime changes file mtime
func (driver *ClientDriver) SetFileMtime(cc server.ClientContext, path string, mtime time.Time) error {
	path = driver.baseDir + path
	return os.Chtimes(path, mtime, mtime)
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

// (copied from net/http/httptest)
// localhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at the last second of 2049 (the end
// of ASN.1 time).
// generated from src/crypto/tls:
// go run "$(go env GOROOT)/src/crypto/tls/generate_cert.go" \
//   --rsa-bits 2048 \
//   --host 127.0.0.1,::1,example.com \
//   --ca --start-date "Jan 1 00:00:00 1970" \
//   --duration=1000000h
// The initial 512 bits key caused this error:
// "tls: failed to sign handshake: crypto/rsa: key size too small for PSS signature"
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDGTCCAgGgAwIBAgIRAJ5VaFcqzaSMmEpeZc33uuowDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAMmv1cldip1/97VnNpPElc5Msa69Cx9l2LmtPubok3pcQy4lS/uF
1zlMDwFseRYuYMOy+lafmYsCO1OFQvt4dginlSZ9yUNq7qSv+dvOUpn6bWQdrLto
d+fDWS4KWiiFsyS78ITozMyRS3G9mJS8YSbGuV4O50UpOJQd6yN5pMQEnp/wHfRI
6y9OYOjWe2snw3rXq1wN7wkj4iVKgrqkJJUHe7Heq4uD7uGfABfOyACmzYFxexXN
f+++L/DesKyMH2At+nKmBtF3JixViIyVKpsCz6Lce1P90n39lYuQDHbV/N2P7ww8
fiwH7fA30yfDqxWKUXhxu7eHGxD1GCFBpgcCAwEAAaNoMGYwDgYDVR0PAQH/BAQD
AgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wLgYDVR0R
BCcwJYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJKoZI
hvcNAQELBQADggEBAHwPqImQQHsqao5MSmhd3S8XQiH6V95T2qOiF9pri7TfMF8h
JWhkaFKQ0yMi37reSZ5+cBiHO9z3UIAmpAWLdqbtVOnh2bchVMO8nSnKHkrOqV2E
IK0Fq5QVW2wyHlYaOMLLQ2sA4I3J/yHl6W4rigetEzY8OtQtPFbg/S1YMqFV8mRz
8PAxtrOWK+ARJP9tqgbylcL6/6cZc6lBQcs0BuwXjcI6fxi+YBXEqbpah9tGRivD
X/k0l93dx/zfNc1Yz06VrCpko6W2Kqa76F6tDIa+WpfZba7t7jNFZTPB4dymUS9L
ICoGMF9k6xscqAURRx8RoSiELemGE5kYUsyvqSE=
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDJr9XJXYqdf/e1
ZzaTxJXOTLGuvQsfZdi5rT7m6JN6XEMuJUv7hdc5TA8BbHkWLmDDsvpWn5mLAjtT
hUL7eHYIp5UmfclDau6kr/nbzlKZ+m1kHay7aHfnw1kuCloohbMku/CE6MzMkUtx
vZiUvGEmxrleDudFKTiUHesjeaTEBJ6f8B30SOsvTmDo1ntrJ8N616tcDe8JI+Il
SoK6pCSVB3ux3quLg+7hnwAXzsgAps2BcXsVzX/vvi/w3rCsjB9gLfpypgbRdyYs
VYiMlSqbAs+i3HtT/dJ9/ZWLkAx21fzdj+8MPH4sB+3wN9Mnw6sVilF4cbu3hxsQ
9RghQaYHAgMBAAECggEADmJN+viC5Ey2G+fqiotgq7/ohC/TVT/sPwHOFKXNrtJZ
sDbUvnGDMgDsqQtVb3GLUSm4lOj5CGL2XDSK3Ghw8pkRGBesfPRpZLFwPm7ukTC9
EIDVSuBefNb/yzrNx0oRxrLoqnH3+Tb7jHcbJLBytVNC8SRa9iHEeTvRA0yvpZMW
WriTbAELv+Zcjal2fPYYtTE9HnRpJX7kHvnCRzlGza0MIs8Q4QgmBE20GRCEXaRi
4jPYjlBx/N4mdD1MTz9jAq+WCHQNJS6aic6l5jidemsSDjtLkSIy8mpTSbA83BTe
qkjAbxtSQ5FKYYH6zDhNbGKwmyqaF1g5gMPSFaDjUQKBgQDfsXamsiyf2Er1+WyI
WyFxwRJmDJ8y3IlH5z5d8Gw+b3JFt72Cy6MVYsO564UALQgUwJmxFIjnVh5F/ZA5
nwsfyyfUkpwQtiqZOTzeTnMd1wt4MPmGwaxfGwVhG5fUgYKnt1GTyF4zHz653RoL
AA0hhsiVmd4hb53PfVHEMVPEewKBgQDm0LzTkgz4zYgocRxiqa4n62VngRS2l7vs
oBgf6o7Dssp1aOucM5uryqzOZsAB/BwCVVeVTnC5nCL2os59YFWbBLlt15l7ykBo
HvUwfmf0R+81onMDqjYPj1+9CSKw4BbTD0WMBOUehvMpL6/k9CsAC2jXQ0oH735V
7dQHEZ1s5QKBgGNbGn1eBE4XLuxkFd3WxFsXS4nCL2/S3rLuNhhZcmqk65elzenr
cwtLq+3He3KhjcZR6bHqkghWiunBfy7ownMjtBRJ7kHJ98/IyY1gQOdPHcwLzLkb
CunPQatpKx37TEIcPYKra5O/XAgH+cpLAooSqMMx7aTiQ7DmU8wVsMRDAoGAQHcM
RgsElHjTDniI9QVvHrcgG0hyAI1gbzZHhqJ8PSwyX5huNbI0SEbS/NK1zdgb+orb
a1f9I9n36eqOwXWmcyVepM8SjwBt/Kao1GJ5pkBxDwnQFbX0Y2Qn2SQ0DDKKLWiW
hATZ+Sy3vUkUV13apKiLH5QrmQvKvTUvgsnorgECgYEAsuf9V7HXDVL3Za1imywP
B8waIgXRIjSWT4Fje7RTMT948qhguVhpoAgVzwzMqizzq6YIQbL7MHwXj7oZNUoQ
CARLpnYLaeWP2nxQyzwGx5pn9TJwg79Yknr8PbSjeym1BSbE5C9ruqar4PfiIzYx
di02m2YJAvRsG9VDpXogi+c=
-----END PRIVATE KEY-----`)

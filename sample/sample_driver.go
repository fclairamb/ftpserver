package sample

import (
	"github.com/fclairamb/ftpserver/server"
	"errors"
	"fmt"
	"strings"
	"os"
	"io/ioutil"
	"github.com/naoina/toml"
)

// SampleDriver defines a very basic serverftp driver
type SampleDriver struct {

}

func (driver SampleDriver) WelcomeUser(cc server.ClientContext) (string, error) {
	// This will remain the official name for now
	return "Welcome on PARADISE FTP !", nil
}

func (driver SampleDriver) CheckUser(cc server.ClientContext, user, pass string) error {
	if user == "bad" || pass == "bad" {
		return errors.New("BAD username or password !")
	} else {
		return nil
	}
}

func (driver SampleDriver) GoToDirectory(cc server.ClientContext, directory string) error {
	if strings.HasPrefix(directory, "/root") {
		return errors.New("This doesn't look good !")
	}
	return nil
}

func (driver SampleDriver) GetFiles(cc server.ClientContext) ([]map[string]string, error) {
	files := make([]map[string]string, 0)

	path := cc.Path()

	if path == "/" {
		{
			file := make(map[string]string)
			file["size"] = "4096"
			file["isDir"] = "true"
			file["name"] = "home"
			files = append(files, file)
		}
		{
			file := make(map[string]string)
			file["size"] = "4096"
			file["isDir"] = "true"
			file["name"] = "root"
			files = append(files, file)
		}
	}

	if path == "/home" {
		for i := 0; i < 5; i++ {
			file := make(map[string]string)
			file["size"] = "90210"
			file["name"] = fmt.Sprintf("paradise_%d.txt", i)
			files = append(files, file)
		}
	}

	return files, nil
}

func (driver SampleDriver) UserLeft(cc server.ClientContext) {

}

func (driver SampleDriver) StartFileUpload (cc server.ClientContext, path string) (server.FileContext, error) {
	return os.Create("/tmp/"+path)
}


// We actually only need this for a more complex implementation.
/*
type FileWriter struct {
	Name string
}

func (fw FileWriter) Write(buf []byte) error {
	return nil
}

func (fw FileWriter) Close() error {
	return nil
}
*/

func (driver SampleDriver) GetSettings() *server.Settings {
	f, err := os.Open("conf/settings.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var config server.Settings
	if err := toml.Unmarshal(buf, &config); err != nil {
		panic(err)
	}
	return &config
}


// Note: This is not a mistake. Interface can be pointers. There seems to be a lot of confusion around this in the
//       server_ftp original code.
func NewSampleDriver() server.Driver {
	return new(SampleDriver)
}

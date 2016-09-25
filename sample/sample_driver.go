package sample

import (
	"github.com/fclairamb/ftpserver/server"
	"errors"
	"fmt"
	"strings"
)

// SampleDriver defines a very basic serverftp driver
type SampleDriver struct {

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
	cc.UserInfo()["path"] = directory
	return nil
}

func (driver SampleDriver) GetFiles(cc server.ClientContext) ([]map[string]string, error) {
	files := make([]map[string]string, 0)

	userInfo := cc.UserInfo()

	if userInfo["path"] == "/" {
		file := make(map[string]string)
		file["size"] = "1024"
		file["isDir"] = "true"
		file["name"] = "home"
		files = append(files, file)
	}

	for i := 0; i < 5; i++ {
		file := make(map[string]string)
		file["size"] = "90210"
		file["name"] = fmt.Sprintf("paradise_%d.txt", i)
		files = append(files, file)
	}

	return files, nil
}

// Note: This is not a mistake. Interface can be pointers. There seems to be a lot of confusion around this in the
//       server_ftp original code.
func NewSampleDriver() server.Driver {
	return new(SampleDriver)
}

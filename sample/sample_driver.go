package sample

import "github.com/fclairamb/ftpserver/server"

// SampleDriver defines a very basic serverftp driver
type SampleDriver struct {

}

func (driver SampleDriver) CheckUser(user, pass string, userInfo *map[string]string) bool {
	return true
}

func (driver SampleDriver) GetFiles(userInfo *map[string]string) ([]map[string]string, error) {
	files := make([]map[string]string, 0)

	//if p.user == "test" {
	// no op just to use p.user as example
	//}

	for i := 0; i < 5; i++ {
		file := make(map[string]string)
		file["size"] = "90210"
		file["name"] = "paradise.txt"
		files = append(files, file)
	}

	return files, nil
}

// Note: This is not a mistake. Interface can be pointers. There seems to be a lot of confusion around this in the
//       server_ftp original code.
func NewSampleDriver() server.Driver {
	return new(SampleDriver)
}

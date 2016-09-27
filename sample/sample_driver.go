package sample

import (
	"github.com/fclairamb/ftpserver/server"
	"errors"
	"strings"
	"os"
	"io/ioutil"
	"github.com/naoina/toml"
	"time"
)

var BASE_DIR = "/tmp/ftpserver"

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

func (driver SampleDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	if strings.HasPrefix(directory, "/root") {
		return errors.New("This doesn't look good !")
	} else if directory == "/virtual" {
		return nil
	}
	_, err := os.Stat(BASE_DIR + directory)
	return err
}

func (driver SampleDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(BASE_DIR + directory, 0777)
}

func (driver SampleDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {

	if ( cc.Path() == "/virtual") {
		files := make([]os.FileInfo, 0)
		files = append(files,
			VirtualFile{
				name: "file1.txt",
				mode: os.FileMode(0666),
				size: 1024,
			},
			VirtualFile{
				name: "file2.txt",
				mode: os.FileMode(0666),
				size: 2048,
			},
		)
		return files, nil
	}

	path := BASE_DIR + cc.Path()

	files, err := ioutil.ReadDir(path)

	// We add a virtual dir
	if cc.Path() == "/" && err == nil {
		files = append(files, VirtualFile{
			name: "virtual",
			mode: os.FileMode(0666) | os.ModeDir,
			size: 4096,
		})
	}

	return files, err
}

func (driver SampleDriver) UserLeft(cc server.ClientContext) {

}

func (driver SampleDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileContext, error) {

	if path == "/virtual/file1.txt" {
		path = "/.file1.txt"
	}

	path = BASE_DIR + path

	// If we are writing and we are not in append mode, we should remove the file
	if ( flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			os.Remove(path)
		}
	}

	return os.OpenFile(path, flag, 0666)
}

func (driver SampleDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = BASE_DIR + path

	return os.Stat(path)
}

func (driver SampleDriver) DeleteFile(cc server.ClientContext, path string) error {
	path = BASE_DIR + path

	return os.Remove(path)
}

func (driver SampleDriver) RenameFile(cc server.ClientContext, from, to string) error {
	from = BASE_DIR + from
	to = BASE_DIR + to

	return os.Rename(from, to)
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
	os.MkdirAll(BASE_DIR, 0777)
	return new(SampleDriver)
}


type VirtualFile struct {
	name string
	size int64
	mode os.FileMode
}

func (f VirtualFile) Name() string {
	return f.name
}

func (f VirtualFile) Size() int64 {
	return f.size
}

func (f VirtualFile) Mode() os.FileMode {
	return f.mode
}

func (f VirtualFile) IsDir() bool {
	return f.mode.IsDir()
}

func (f VirtualFile) ModTime() time.Time {
	return time.Now().UTC()
}

func (f VirtualFile) Sys() interface{} {
	return nil
}
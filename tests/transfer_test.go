package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/secsy/goftp"

	"github.com/fclairamb/ftpserver/server"
)

func createTemporaryFile(t *testing.T, targetSize int) *os.File {
	var file *os.File

	var fileErr error

	if file, fileErr = ioutil.TempFile("", "ftpserver"); fileErr != nil {
		t.Fatal("Temporary creation error:", fileErr)
		return nil
	}

	src := rand.New(rand.NewSource(0))
	if _, err := io.CopyN(file, src, int64(targetSize)); err != nil {
		t.Fatal("Couldn't copy:", err)
		return nil
	}

	return file
}

func hashFile(t *testing.T, file *os.File) string {
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal("Couldn't seek:", err)
	}

	hashser := sha256.New()

	if _, err := io.Copy(hashser, file); err != nil {
		t.Fatal("Couldn't hashUpload:", err)
	}

	hash := hex.EncodeToString(hashser.Sum(nil))

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal("Couldn't seek:", err)
	}

	return hash
}

func ftpUpload(t *testing.T, ftp *goftp.Client, file *os.File, filename string) {
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal("Couldn't seek:", err)
	}

	if err := ftp.Store(filename+".tmp", file); err != nil {
		t.Fatal("Couldn't upload bin:", err)
	}

	if err := ftp.Rename(filename+".tmp", filename); err != nil {
		t.Fatal("Can't rename file:", err)
	}

	if _, err := ftp.Stat(filename); err != nil {
		t.Fatal("Couldn't get the size of file1.bin:", err)
	}

	if stats, err := ftp.Stat(filename); err != nil {
		// That's acceptable for now
		t.Log("Couldn't stat file:", err)
	} else {
		found := false
		if strings.HasSuffix(stats.Name(), filename) {
			found = true
		}
		if !found {
			t.Fatal("STAT: Couldn't find file !")
		}
	}
}

func ftpDownloadAndHash(t *testing.T, ftp *goftp.Client, filename string) string {
	hasher := sha256.New()
	if err := ftp.Retrieve(filename, hasher); err != nil {
		t.Fatal("Couldn't fetch file:", err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func ftpDelete(t *testing.T, ftp *goftp.Client, filename string) {
	if err := ftp.Delete(filename); err != nil {
		t.Fatal("Couldn't delete file "+filename+":", err)
	}

	if err := ftp.Delete(filename); err == nil {
		t.Fatal("Should have had a problem deleting " + filename)
	}
}

// TestTransfer validates the upload of file in both active and passive mode
func TestTransfer(t *testing.T) {
	s := NewTestServerWithDriver(&ServerDriver{Debug: true, Settings: &server.Settings{ActiveTransferPortNon20: true}})
	defer s.Stop()

	testTransferOnConnection(t, s, false)
	testTransferOnConnection(t, s, true)
}

func testTransferOnConnection(t *testing.T, server *server.FtpServer, active bool) {
	conf := goftp.Config{
		User:            "test",
		Password:        "test",
		ActiveTransfers: active,
	}

	var err error

	var c *goftp.Client

	if c, err = goftp.DialConfig(conf, server.Addr()); err != nil {
		t.Fatal("Couldn't connect", err)
	}

	defer func() { panicOnError(c.Close()) }()

	var hashUpload, hashDownload string
	{ // We create a 10MB file and upload it
		file := createTemporaryFile(t, 10*1024*1024)
		hashUpload = hashFile(t, file)
		ftpUpload(t, c, file, "file.bin")
	}

	{ // We download the file we just uploaded
		hashDownload = ftpDownloadAndHash(t, c, "file.bin")
		ftpDelete(t, c, "file.bin")
	}

	// We make sure the hashes of the two files match
	if hashUpload != hashDownload {
		t.Fatal("The two files don't have the same hash:", hashUpload, "!=", hashDownload)
	}
}

// TestFailedTransfer validates the handling of failed transfer caused by file access issues
func TestFailedTransfer(t *testing.T) {
	s := NewTestServer(true)
	defer s.Stop()

	conf := goftp.Config{
		User:     "test",
		Password: "test",
	}

	var err error

	var c *goftp.Client

	if c, err = goftp.DialConfig(conf, s.Addr()); err != nil {
		t.Fatal("Couldn't connect", err)
	}

	defer func() { panicOnError(c.Close()) }()

	// We create a 1KB file and upload it
	file := createTemporaryFile(t, 1*1024)
	if err = c.Store("/non/existing/path/file.bin", file); err == nil {
		t.Fatal("This upload should have failed")
	}

	if err = c.Store("file.bin", file); err != nil {
		t.Fatal("This upload should have succeeded", err)
	}
}

func TestFailedFileClose(t *testing.T) {
	driver := &ServerDriver{
		Debug:      true,
		FileStream: &failingCloser{},
	}

	s := NewTestServerWithDriver(driver)
	defer s.Stop()

	conf := goftp.Config{
		User:     "test",
		Password: "test",
	}

	var err error

	var c *goftp.Client

	if c, err = goftp.DialConfig(conf, s.Addr()); err != nil {
		t.Fatal("Couldn't connect", err)
	}

	defer func() { panicOnError(c.Close()) }()

	file := createTemporaryFile(t, 1*1024)
	err = c.Store("file.bin", file)

	if err == nil {
		t.Fatal("this upload should not succeed", err)
	}

	if !strings.Contains(err.Error(), errFailingCloser.Error()) {
		t.Errorf("got %s as the error message, want it to contain %s", err, errFailingCloser)
	}
}

type failingCloser struct {
	os.File
}

var errFailingCloser = errors.New("the hard disk crashed")

func (f *failingCloser) Close() error { return errFailingCloser }

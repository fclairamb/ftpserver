package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"gopkg.in/dutchcoders/goftp.v1"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func TestTransfer(t *testing.T) {
	s := getServer(true)
	defer s.Stop()

	var connErr error
	var ftp *goftp.FTP

	if ftp, connErr = goftp.Connect(s.Listener.Addr().String()); connErr != nil {
		t.Fatal("Couldn't connect", connErr)
	}
	defer ftp.Quit()

	if err := ftp.Login("test", "test"); err != nil {
		t.Fatal("Failed to login:", err)
	}

	var hashUpload, hashDownload string
	{ // We create a 20MB file and upload it
		var file *os.File
		var fileErr error
		if file, fileErr = ioutil.TempFile("", "ftpserver"); fileErr != nil {
			t.Fatal("Temporary creation error:", fileErr)
		}

		src := rand.New(rand.NewSource(0))
		if _, err := io.CopyN(file, src, 10*1024*1024); err != nil {
			t.Fatal("Couldn't copy:", err)
		}

		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal("Couldn't seek:", err)
		}

		hashser := sha256.New()
		if _, err := io.Copy(hashser, file); err != nil {
			t.Fatal("Couldn't hashUpload:", err)
		}

		hashUpload = hex.EncodeToString(hashser.Sum(nil))
		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal("Couldn't seek:", err)
		}

		if err := ftp.Stor("file.bin", file); err != nil {
			t.Fatal("Couldn't upload bin:", err)
		}
	}

	{ // We download the file we just uploaded
		readFunc := func(r io.Reader) error {
			var hasher = sha256.New()
			if _, err := io.Copy(hasher, r); err != nil {
				return err
			}

			hashDownload = hex.EncodeToString(hasher.Sum(nil))

			return nil
		}

		if _, err := ftp.Retr("file.bin", readFunc); err != nil {
			t.Fatal("Couldn't fetch file:", err)
		}
	}

	// We make sure the hashes of the two files match
	if hashUpload != hashDownload {
		t.Fatal("The two files don't have the same hash:", hashUpload, "!=", hashDownload)
	}
}

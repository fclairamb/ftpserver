package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"gopkg.in/dutchcoders/goftp.v1"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func TestTransfer(t *testing.T) {
	s := NewTestServer(true)
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

		targetSize := 10 * 1024 * 1024

		src := rand.New(rand.NewSource(0))
		if _, err := io.CopyN(file, src, int64(targetSize)); err != nil {
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

		if err := ftp.Stor("file1.bin", file); err != nil {
			t.Fatal("Couldn't upload bin:", err)
		}

		if size, err := ftp.Size("file1.bin"); err != nil {
			t.Fatal("Couldn't get the size of file1.bin:", err)
		} else {
			if size != targetSize {
				t.Fatalf("Size is %d instead of %d", size, targetSize)
			}
		}

		if err := ftp.Rename("file1.bin", "file2.bin"); err != nil {
			t.Fatal("Can't rename file:", err)
		}

		if stats, err := ftp.Stat("file2.bin"); err != nil {
			// That's acceptable for now
			t.Log("Couldn't stat file:", err)
		} else {
			found := false
			for _, l := range stats {
				if strings.HasSuffix(l, "file2.bin") {
					found = true
				}
			}
			if !found {
				t.Fatal("STAT: Couldn't find file !")
			}
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

		if _, err := ftp.Retr("file2.bin", readFunc); err != nil {
			t.Fatal("Couldn't fetch file:", err)
		}

		if err := ftp.Dele("file2.bin"); err != nil {
			t.Fatal("Couldn't delete file", err)
		}

		if err := ftp.Dele("file2.bin"); err == nil {
			t.Fatal("Should have had a problem deleting file2.bin")
		}
	}

	// We make sure the hashes of the two files match
	if hashUpload != hashDownload {
		t.Fatal("The two files don't have the same hash:", hashUpload, "!=", hashDownload)
	}
}

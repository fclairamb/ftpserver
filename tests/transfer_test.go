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

func ftpUpload(t *testing.T, ftp *goftp.FTP, file *os.File, filename string) {
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal("Couldn't seek:", err)
	}
	if err := ftp.Stor(filename+".tmp", file); err != nil {
		t.Fatal("Couldn't ftpUpload bin:", err)
	}

	if err := ftp.Rename(filename+".tmp", filename); err != nil {
		t.Fatal("Can't rename file:", err)
	}

	if _, err := ftp.Size(filename); err != nil {
		t.Fatal("Couldn't get the size of file1.bin:", err)
	}

	if stats, err := ftp.Stat(filename); err != nil {
		// That's acceptable for now
		t.Log("Couldn't stat file:", err)
	} else {
		found := false
		for _, l := range stats {
			if strings.HasSuffix(l, filename) {
				found = true
			}
		}
		if !found {
			t.Fatal("STAT: Couldn't find file !")
		}
	}
}

func ftpDownloadAndHash(t *testing.T, ftp *goftp.FTP, filename string) string {
	var hash string
	readFunc := func(r io.Reader) error {
		var hasher = sha256.New()
		if _, err := io.Copy(hasher, r); err != nil {
			return err
		}

		hash = hex.EncodeToString(hasher.Sum(nil))

		return nil
	}

	if _, err := ftp.Retr(filename, readFunc); err != nil {
		t.Fatal("Couldn't fetch file:", err)
	}

	return hash
}

func ftpDelete(t *testing.T, ftp *goftp.FTP, filename string) {
	if err := ftp.Dele(filename); err != nil {
		t.Fatal("Couldn't ftpDelete file", err)
	}

	if err := ftp.Dele(filename); err == nil {
		t.Fatal("Should have had a problem deleting file2.bin")
	}
}

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
	{ // We create a 10MB file and upload it
		file := createTemporaryFile(t, 10*1024*1024)
		hashUpload = hashFile(t, file)
		ftpUpload(t, ftp, file, "file.bin")
	}

	{ // We download the file we just uploaded
		hashDownload = ftpDownloadAndHash(t, ftp, "file.bin")
		ftpDelete(t, ftp, "file.bin")
	}

	// We make sure the hashes of the two files match
	if hashUpload != hashDownload {
		t.Fatal("The two files don't have the same hash:", hashUpload, "!=", hashDownload)
	}
}

package tests

import (
	"gopkg.in/dutchcoders/goftp.v1"
	"strings"
	"testing"
)

func TestDirAccess(t *testing.T) {
	s := NewTestServer(true)
	defer s.Stop()

	var connErr error
	var ftp *goftp.FTP

	if ftp, connErr = goftp.Connect(s.Listener.Addr().String()); connErr != nil {
		t.Fatal("Couldn't connect", connErr)
	}
	defer ftp.Quit()

	if _, err := ftp.List("/"); err == nil {
		t.Fatal("We could list files before login")
	}

	if err := ftp.Login("test", "test"); err != nil {
		t.Fatal("Failed to login:", err)
	}

	if path, err := ftp.Pwd(); err != nil {
		t.Fatal("Couldn't test PWD", err)
	} else if path != "/" {
		t.Fatal("Bad path:", path)
	}

	if err := ftp.Cwd("/unknown"); err == nil {
		t.Fatal("We should have had an error")
	}

	if err := ftp.Mkd("/known"); err != nil {
		t.Fatal("Couldn't create dir:", err)
	}

	if files, err := ftp.List("/"); err != nil {
		t.Fatal("Couldn't list files:", err)
	} else {
		found := false
		for _, f := range files {
			if strings.HasSuffix(strings.TrimSpace(f), "known") {
				found = true
			}
		}
		if !found {
			t.Fatal("Couldn't find the known dir")
		}
	}

	if err := ftp.Cwd("/known"); err != nil {
		t.Fatal("Couldn't access the known dir:", err)
	}

	if err := ftp.Rmd("/known"); err != nil {
		t.Fatal("Couldn't delete the known dir:", err)
	}

	if err := ftp.Rmd("/known"); err == nil {
		t.Fatal("We shouldn't have been able to delete known again")
	}
}

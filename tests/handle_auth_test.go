package tests

import (
	"gopkg.in/dutchcoders/goftp.v1"
	"testing"
)

func TestLoginSuccess(t *testing.T) {
	s := getServer(true)
	defer s.Stop()

	var err error
	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(s.Listener.Addr().String()); err != nil {
		t.Fatal("Couldn't connect", err)
	}
	defer ftp.Quit()

	if err = ftp.Login("test", "test"); err != nil {
		t.Fatal("Failed to login:", err)
	}

	if err := ftp.Noop(); err != nil {
		t.Fatal("Couldn't NOOP:", err)
	}

	if line, err := ftp.Syst(); err != nil {
		t.Fatal("Couldn't SYST:", err)
	} else {
		if line != "UNIX Type: L8" {
			t.Fatal("SYST:", line)
		}
	}
}

func TestLoginFailure(t *testing.T) {
	s := getServer(true)
	defer s.Stop()

	var err error
	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(s.Listener.Addr().String()); err != nil {
		t.Fatal("Couldn't connect:", err)
	}

	defer ftp.Quit()

	if err = ftp.Login("test", "test2"); err == nil {
		t.Fatal("We should have failed to login")
	}
}

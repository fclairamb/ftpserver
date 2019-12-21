package tests

import (
	"crypto/tls"
	"fmt"
	"testing"

	"gopkg.in/dutchcoders/goftp.v1"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func reportError(err error) {
	if err != nil {
		fmt.Println("Error reporting:", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	s := NewTestServer(true)
	defer s.Stop()

	var err error

	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(s.Addr()); err != nil {
		t.Fatal("Couldn't connect", err)
	}

	defer func() { panicOnError(ftp.Quit()) }()

	if err = ftp.Noop(); err != nil {
		t.Fatal("Couldn't NOOP before login:", err)
	}

	if err = ftp.Login("test", "test"); err != nil {
		t.Fatal("Failed to login:", err)
	}

	if err := ftp.Noop(); err != nil {
		t.Fatal("Couldn't NOOP:", err)
	}

	if line, err := ftp.Syst(); err != nil {
		t.Fatal("Couldn't SYST:", err)
	} else if line != "UNIX Type: L8" {
		t.Fatal("SYST:", line)
	}
}

func TestLoginFailure(t *testing.T) {
	s := NewTestServer(true)
	defer s.Stop()

	var err error

	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(s.Addr()); err != nil {
		t.Fatal("Couldn't connect:", err)
	}

	defer func() { reportError(ftp.Quit()) }()

	if err = ftp.Login("test", "test2"); err == nil {
		t.Fatal("We should have failed to login")
	}
}

func TestAuthTLS(t *testing.T) {
	s := NewTestServerWithDriver(&ServerDriver{
		Debug: true,
		TLS:   true,
	})
	defer s.Stop()

	ftp, err := goftp.Connect(s.Addr())
	if err != nil {
		t.Fatal("Couldn't connect:", err)
	}

	defer func() { reportError(ftp.Quit()) }()

	config := &tls.Config{
		// nolint:gosec
		InsecureSkipVerify: true,
		ClientAuth:         tls.RequestClientCert,
	}
	if err := ftp.AuthTLS(config); err != nil {
		t.Fatal("Couldn't upgrade connection to TLS:", err)
	}
}

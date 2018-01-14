package tests

import (
	"testing"

	"time"

	"github.com/fclairamb/ftpserver/server"
	"github.com/secsy/goftp"
)

func TestSiteCommand(t *testing.T) {
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
	defer c.Close()

	var raw goftp.RawConn

	if raw, err = c.OpenRawConn(); err != nil {
		t.Fatal("Couldn't open raw connection")
	}

	if rc, response, err := raw.SendCommand("SITE HELP"); err != nil {
		t.Fatal("Command not accepted", err)
	} else {
		if rc != 500 {
			t.Fatal("Are we supporting it now ?", rc)
		}
		if response != "Not understood SITE subcommand" {
			t.Fatal("Are we supporting it now ?", response)
		}
	}
}

// florent(2018-01-14): #58: IDLE timeout: Testing timeout
func TestIdleTimeout(t *testing.T) {
	s := NewTestServerWithDriver(&ServerDriver{Debug: true, Settings: &server.Settings{IdleTimeout: 2}})
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
	defer c.Close()

	var raw goftp.RawConn

	if raw, err = c.OpenRawConn(); err != nil {
		t.Fatal("Couldn't open raw connection")
	}

	time.Sleep(time.Second * 1) // < 2s : OK

	if rc, _, err := raw.SendCommand("NOOP"); err != nil || rc != 200 {
		t.Fatal("Command not accepted", rc, err)
	}

	time.Sleep(time.Second * 3) // > 2s : Timeout

	if rc, _, err := raw.SendCommand("NOOP"); err != nil || rc != 421 {
		t.Fatal("Command should have failed !")
	}
}

package tests

import (
	"testing"

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

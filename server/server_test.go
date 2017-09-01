package server

import "testing"

func TestPortCommandFormatOK(t *testing.T) {

	net, err := parseRemoteAddr("127,0,0,1,239,163")
	if err != nil {
		t.Fatal("Problem parsing", err)
	}
	if net.IP.String() != "127.0.0.1" {
		t.Fatal("Problem parsing IP", net.IP.String())
	}
	if net.Port != 239<<8+163 {
		t.Fatal("Problem parsing port", net.Port)
	}
}

func TestPortCommandFormatInvalid(t *testing.T) {
	badFormats := []string{
		"127,0,0,1,239,",
		"127,0,0,1,1,1,1",
	}
	for _, f := range badFormats {
		_, err := parseRemoteAddr(f)
		if err == nil {
			t.Fatal("This should have failed", f)
		}
	}
}

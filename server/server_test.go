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

func Test_qoutedoubling(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"1", args{" white space"}, " white space"},
		{"1", args{` one" quote`}, ` one"" quote`},
		{"1", args{` two"" quote`}, ` two"""" quote`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := quoteDoubling(tt.args.s); got != tt.want {
				t.Errorf("quoteDoubling() = %v, want %v", got, tt.want)
			}
		})
	}
}

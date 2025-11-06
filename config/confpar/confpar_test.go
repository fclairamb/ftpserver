package confpar

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestDurationMarshalJSON(t *testing.T) {
	for _, tt := range []struct {
		input    Duration
		expected []byte
	}{
		{Duration{0}, []byte(`"0s"`)},
		{Duration{31 * time.Second}, []byte(`"31s"`)},
		{Duration{5 * time.Minute}, []byte(`"5m0s"`)},
		{Duration{time.Hour}, []byte(`"1h0m0s"`)},
	} {

		b, err := json.Marshal(&tt.input)

		if err != nil {
			t.Fatalf("json.Marshal(): %v", err)
		}

		if !bytes.Equal(b, tt.expected) {
			t.Fatalf("have:%q want:%q", b, tt.expected)
		}
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	for _, tt := range []struct {
		input    []byte
		expected time.Duration
		wantErr  bool
	}{
		{[]byte(`"5m"`), 5 * time.Minute, false},
		{[]byte(`"30s"`), 30 * time.Second, false},
		{[]byte(`"1h"`), time.Hour, false},
		{[]byte(`"0s"`), 0, false},
		{[]byte(`"invalid"`), 0, true},
	} {
		var d Duration
		err := json.Unmarshal(tt.input, &d)

		if err == nil && tt.wantErr {
			t.Fatalf("expecting error for invalid input")
		}

		if err != nil && !tt.wantErr {
			t.Fatalf("json.Unmarshal(): %v", err)
		}

		if d.Duration != tt.expected {
			t.Fatalf("have:%v want:%v", d.Duration, tt.expected)
		}
	}
}

package confpar

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestDurationMarshalJSON(t *testing.T) {
	for _, item := range []struct {
		duration time.Duration
		want     string
	}{
		{0, `0s`},
		{31 * time.Second, `31s`},
		{5 * time.Minute, `5m0s`},
		{1 * time.Hour, `1h0m0s`},
	} {
		have, err := json.Marshal(&Duration{item.duration})
		if err != nil {
			t.Fatalf("json.Marshal(): %v", err)
		}
		if !bytes.Equal(have, []byte(`"`+item.want+`"`)) {
			t.Fatalf("have:%s want:%q", string(have), item.want)
		}
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	for _, item := range []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{`5m`, 5 * time.Minute, false},
		{`30s`, 30 * time.Second, false},
		{`1h`, time.Hour, false},
		{`0s`, 0, false},
		{`invalid`, 0, true},
	} {
		var have Duration
		err := json.Unmarshal([]byte(`"`+item.input+`"`), &have)
		if err == nil && item.wantErr {
			t.Fatalf("expecting error for invalid input")
		}
		if err != nil && !item.wantErr {
			t.Fatalf("json.Unmarshal(): %v", err)
		}
		if have.Duration != item.want {
			t.Fatalf("have:%v want:%v", have.Duration, item.want)
		}
	}
}

package confpar

import (
	"encoding/json"
	"testing"
	"time"
)

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
		val := &Duration{}
		err := json.Unmarshal(tt.input, &val)

		if err == nil && tt.wantErr {
			t.Fatalf("expecting error for invalid input")
		}

		if err != nil && !tt.wantErr {
			t.Fatalf("json.Unmarshal(): %v", err)
		}

		if val.Duration != tt.expected {
			t.Fatalf("have:%v want:%v", val.Duration, tt.expected)
		}
	}
}

package utils

import "testing"

func TestEnvReplace(t *testing.T) {
	t.Setenv("TEST", "abc")
	value := ReplaceEnvVars("/tmp/$TEST/def")
	if value != "/tmp/abc/def" {
		t.Error("EnvReplace failed", value)
	}
}

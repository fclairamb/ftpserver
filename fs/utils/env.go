package utils

import (
	"os"
	"regexp"
)

// Replace all environment variables in a string by their actual values

var envVar = regexp.MustCompile(`\$[a-zA-Z0-9_]+`)

func ReplaceEnvVars(s string) string {
	return envVar.ReplaceAllStringFunc(s, func(s string) string {
		return os.Getenv(s[1:])
	})
}

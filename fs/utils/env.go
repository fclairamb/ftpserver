// Package utils provides helpers shared by the file systems
package utils

import (
	"os"
	"regexp"
)

var envVar = regexp.MustCompile(`\$[a-zA-Z0-9_]+`)

// ReplaceEnvVars replaces all the environment variables ($NAME) in a string by their actual values
func ReplaceEnvVars(s string) string {
	return envVar.ReplaceAllStringFunc(s, func(s string) string {
		return os.Getenv(s[1:])
	})
}

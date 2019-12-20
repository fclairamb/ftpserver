package log

import (
	"os"
	"testing"

	gklog "github.com/go-kit/kit/log"
)

func getLogger() Logger {
	return NewGKLogger(gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stdout))).With(
		"ts", gklog.DefaultTimestampUTC,
		"caller", gklog.DefaultCaller,
	)
}

func TestLogSimple(t *testing.T) {
	logger := getLogger()
	logger.Info("msg", "Hello !")
}

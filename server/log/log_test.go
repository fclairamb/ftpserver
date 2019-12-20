package log

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	gklog "github.com/go-kit/kit/log"
)

func getLogger() Logger {
	return NewGKLogger(gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stdout))).With(
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
}

func TestLogSimple(t *testing.T) {
	logger := getLogger()
	logger.Info("msg", "Hello !")
}

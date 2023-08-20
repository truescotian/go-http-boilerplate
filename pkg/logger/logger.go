package logger

import (
	"time"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	logger.SetLevel(logrus.DebugLevel)
}

type Event struct {
	id      int
	message string
}

var (
	invalidArg = Event{1, "Invalid arg: %s"}
)

func InvalidArg(name string) {
	logger.Errorf(invalidArg.message, name)
}

func Infof(args ...interface{}) {
	logger.Info(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

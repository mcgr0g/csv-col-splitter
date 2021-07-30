package logging

import (
	"io"
	logging "log"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		logging.Fatalf("error opening file: %v", err)
	}

	log = logrus.New()
	// log.Level = logrus.DebugLevel // uncomment to see debug messages

	// log.SetReportCaller(true)
	log.SetFormatter(&logrus.TextFormatter{
		// DisableColors: false,
		// ForceColors:   true,
		FullTimestamp: true,
	})

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}

// Info ...
func Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Warn ...
func Warn(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

// Debug ...
func Debug(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Error ...
func Error(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

var (
	ConfigError = "%v type=config.error"

	CmdError = "%v type=cmd.error"

	CmdInfo = "%v type=cmd.info"
)

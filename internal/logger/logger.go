package logger

import (
	"github.com/gookit/color"
	"github.com/lfcypo/viperx"
	"github.com/sirupsen/logrus"
)

func init() {
	color.ForceOpenColor()
}

func DisableColor() {
	color.Disable()
}

func New(name string) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(Output())
	logger.SetFormatter(NewFormatter(name))

	logLevel := viperx.GetString("log.level", "info")

	switch logLevel {
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

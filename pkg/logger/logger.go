package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new configured logger
func NewLogger() *logrus.Logger {
	log := logrus.New()
	
	// Set logger level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel) // Default log level
	}

	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// Output to stdout
	log.SetOutput(os.Stdout)

	return log
}

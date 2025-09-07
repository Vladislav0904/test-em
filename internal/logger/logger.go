package logger

import (
	"os"
	"test-em/internal/config"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(cfg *config.Config) {
	Log = logrus.New()
	logLevel := getEnv("LOG_LEVEL", "info")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
	logFormat := getEnv("LOG_FORMAT", "json")
	if logFormat == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	Log.SetOutput(os.Stdout)

	Log = Log.WithFields(logrus.Fields{
		"service": "subscriptions-api",
	}).Logger
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

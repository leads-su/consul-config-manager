package log

import (
	"fmt"
	"github.com/leads-su/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

// Log describes structure for `log` configuration section
type Log struct {
	Level   string `mapstructure:"level"`
	WriteTo string `mapstructure:"write_to"`
}

// InitializeDefaults create new log config instance with default values
func InitializeDefaults() *Log {
	return &Log{
		Level:   "DEBUG",
		WriteTo: "/var/log/ccm",
	}
}

// SetLogLevel updates current log level for application
func (log *Log) SetLogLevel() {
	logLevel := logrus.ErrorLevel
	switch log.Level {
	case "t", "trace", "Trace", "TRACE":
		logLevel = logrus.TraceLevel
	case "d", "debug", "Debug", "DEBUG":
		logLevel = logrus.DebugLevel
	case "i", "info", "Info", "INFO":
		logLevel = logrus.InfoLevel
	case "w", "warn", "Warn", "WARN":
		logLevel = logrus.WarnLevel
	case "e", "error", "Error", "ERROR":
		logLevel = logrus.ErrorLevel
	case "f", "fatal", "Fatal", "FATAL":
		logLevel = logrus.FatalLevel
	case "p", "panic", "Panic", "PANIC":
		logLevel = logrus.PanicLevel
	default:
		logLevel = logrus.ErrorLevel
	}
	log.bootstrapLogger(logLevel)
}

// CollectLogsLocation collects locations for application logs and stores them in Viper for later usage
func (log *Log) CollectLogsLocation() {
	viper.SetDefault("application.log_level", log.Level)
	viper.SetDefault("application.log_path", log.WriteTo)
}

// bootstrapLogger bootstraps logger functionality
func (log *Log) bootstrapLogger(logLevel logrus.Level) {
	logger.RegisterRotators(logger.RotatorFileConfig{
		FileName:   log.generateLogFilePath("info"),
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     1,
		Level:      logrus.TraceLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		},
		Compress: false,
	}, logger.RotatorFileConfig{
		FileName:   log.generateLogFilePath("error"),
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     1,
		Level:      logrus.ErrorLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		},
		Compress: false,
	})
	logger.SetLogLevel(logLevel)
}

// generateLogFilePath creates OS independent path to log file
func (log *Log) generateLogFilePath(fileName string) string {
	return fmt.Sprintf("%s%s%s.log", log.WriteTo, string(os.PathSeparator), fileName)
}

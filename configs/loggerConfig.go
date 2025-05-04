package configs

import (
	"fmt"
	"os"
)

type loggerConfig struct {
	logPath string
}

type LoggerConfig interface {
	Configure() error
	LogPath() string
}

func (l *loggerConfig) getLogPath() error {
	logPath := os.Getenv("AISERVER_LOG")
	if logPath == "" {
		return fmt.Errorf("not found $AISERVER_LOG")
	}
	l.logPath = logPath
	return nil
}

func (l *loggerConfig) Configure() error {
	err := l.getLogPath()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (l *loggerConfig) LogPath() string {
	return l.logPath
}

func NewLoggerConfig() LoggerConfig {
	return &loggerConfig{}
}

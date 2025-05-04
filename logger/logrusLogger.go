package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gitverse.ru/udonetsm/aiserver/configs"
)

type logrusLogger struct {
	logger  *logrus.Logger
	storage io.WriteCloser
	config  configs.LoggerConfig
}

type Logger interface {
	Info(message ...any)
	Fatal(message ...any)
	Infof(format string, items ...any)
	Configure() error
	CloseLogger() error
}

func (l *logrusLogger) Info(message ...any) {
	l.logger.Info(message...)
}

func (l *logrusLogger) Fatal(message ...any) {
	l.logger.Fatal(message...)
}

func (l *logrusLogger) Infof(format string, items ...any) {
	l.logger.Infof(format, items...)
}

func (l *logrusLogger) Configure() error {
	if l.config.LogPath() == "" {
		l.storage = os.Stderr
	} else {
		file, err := os.OpenFile(l.config.LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("%w<---", err)
		}
		l.storage = file
	}
	l.logger = logrus.New()
	l.logger.SetOutput(l.storage)
	return nil
}

func (l *logrusLogger) CloseLogger() error {
	err := l.storage.Close()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	logrus.Info("see logs")
	return nil
}

func NewLogger(loggerConfig configs.LoggerConfig) Logger {
	return &logrusLogger{config: loggerConfig}
}

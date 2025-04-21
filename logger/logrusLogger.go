package logger

import "github.com/sirupsen/logrus"

type logrusLogger struct {
	logger *logrus.Logger
}

type Logger interface {
	Info(message ...any)
	Fatal(message ...any)
}

func (l *logrusLogger) Info(message ...any) {
	l.logger.Info(message...)
}

func (l *logrusLogger) Fatal(message ...any) {
	l.logger.Fatal(message...)
}

func NewLogger() Logger {
	return &logrusLogger{logger: logrus.New()}
}

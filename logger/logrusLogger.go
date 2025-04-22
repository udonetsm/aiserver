package logger

import "github.com/sirupsen/logrus"

type logrusLogger struct {
	logger *logrus.Logger
}

type Logger interface {
	Info(message ...any)
	Fatal(message ...any)
	Infof(format string, items ...any)
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

func NewLogger() Logger {
	return &logrusLogger{logger: logrus.New()}
}

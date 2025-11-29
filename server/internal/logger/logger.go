package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var Log *LoggerWrapper

type LoggerWrapper struct {
	logger *slog.Logger
}

func init() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	Log = &LoggerWrapper{logger: slog.New(handler)}
}

// WithFields создает новый логгер с дополнительными полями
func (l *LoggerWrapper) WithFields(fields map[string]interface{}) *LoggerWrapper {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &LoggerWrapper{logger: l.logger.With(args...)}
}

// Info логирует сообщение уровня Info
func (l *LoggerWrapper) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Error логирует сообщение уровня Error
func (l *LoggerWrapper) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Warn логирует сообщение уровня Warn
func (l *LoggerWrapper) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Debug логирует сообщение уровня Debug
func (l *LoggerWrapper) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info логирует сообщение уровня Info
func Info(args ...interface{}) {
	Log.logger.Info(fmt.Sprint(args...))
}

// Infof логирует форматированное сообщение уровня Info
func Infof(format string, args ...interface{}) {
	Log.logger.Info(fmt.Sprintf(format, args...))
}

// Error логирует сообщение уровня Error
func Error(args ...interface{}) {
	Log.logger.Error(fmt.Sprint(args...))
}

// Errorf логирует форматированное сообщение уровня Error
func Errorf(format string, args ...interface{}) {
	Log.logger.Error(fmt.Sprintf(format, args...))
}

// Warn логирует сообщение уровня Warn
func Warn(args ...interface{}) {
	Log.logger.Warn(fmt.Sprint(args...))
}

// Warnf логирует форматированное сообщение уровня Warn
func Warnf(format string, args ...interface{}) {
	Log.logger.Warn(fmt.Sprintf(format, args...))
}

// Debug логирует сообщение уровня Debug
func Debug(args ...interface{}) {
	Log.logger.Debug(fmt.Sprint(args...))
}

// Debugf логирует форматированное сообщение уровня Debug
func Debugf(format string, args ...interface{}) {
	Log.logger.Debug(fmt.Sprintf(format, args...))
}

// Fatal логирует сообщение уровня Error и вызывает os.Exit(1)
func Fatal(args ...interface{}) {
	Log.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf логирует форматированное сообщение уровня Error и вызывает os.Exit(1)
func Fatalf(format string, args ...interface{}) {
	Log.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	assert.NotNil(t, Log)
	assert.NotNil(t, Log.logger)
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Info("test info message")

	output := buf.String()
	assert.Contains(t, output, "test info message")
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Infof("test info message with %s", "formatting")

	output := buf.String()
	assert.Contains(t, output, "test info message with formatting")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Error("test error message")

	output := buf.String()
	assert.Contains(t, output, "test error message")
	assert.Contains(t, strings.ToLower(output), "error")
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Errorf("test error: %s", "something went wrong")

	output := buf.String()
	assert.Contains(t, output, "test error: something went wrong")
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Warn("test warning message")

	output := buf.String()
	assert.Contains(t, output, "test warning message")
}

func TestWarnf(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Warnf("test warning: %d items", 5)

	output := buf.String()
	assert.Contains(t, output, "test warning: 5 items")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Debug("test debug message")

	output := buf.String()
	assert.Contains(t, output, "test debug message")
}

func TestDebugf(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Debugf("test debug: %v", true)

	output := buf.String()
	assert.Contains(t, output, "test debug: true")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(args ...interface{})
		message string
	}{
		{
			name:    "Info level",
			logFunc: Info,
			message: "info test",
		},
		{
			name:    "Error level",
			logFunc: Error,
			message: "error test",
		},
		{
			name:    "Warn level",
			logFunc: Warn,
			message: "warn test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
			Log = &LoggerWrapper{logger: slog.New(handler)}
			defer func() {
				handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
				Log = &LoggerWrapper{logger: slog.New(handler)}
			}()

			tt.logFunc(tt.message)

			output := buf.String()
			assert.Contains(t, output, tt.message)
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Log.WithFields(map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}).Info("User logged in")

	output := buf.String()
	assert.Contains(t, output, "User logged in")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "login")
}

func TestMultipleLoggingCalls(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	Log = &LoggerWrapper{logger: slog.New(handler)}
	defer func() {
		handler := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})
		Log = &LoggerWrapper{logger: slog.New(handler)}
	}()

	Info("First message")
	Warn("Second message")
	Error("Third message")

	output := buf.String()
	assert.Contains(t, output, "First message")
	assert.Contains(t, output, "Second message")
	assert.Contains(t, output, "Third message")
}

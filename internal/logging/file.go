package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"clever.secure-onboard.com/internal/config"
)

type FileLogger struct {
	slog.Logger
}

func NewFileLogger(cfg config.LoggingInfo) FileLogger {
	os.WriteFile("/tmp/dcfagent-logs", []byte("---Start---"), os.ModeAppend)
	f, err := os.OpenFile("/tmp/dcfagent-logs", os.O_APPEND, os.ModeAppend) // TODO: Close
	if err != nil {
		fmt.Println(err.Error())
	}
	level, err := strconv.Atoi(cfg.MinLogLevel)
	var typedLevel slog.Level
	if err != nil {
		typedLevel = slog.LevelDebug
	} else {
		typedLevel = slog.Level(level)
	}

	return FileLogger{
		Logger: *slog.New(
			slog.NewJSONHandler(
				f,
				&slog.HandlerOptions{
					AddSource: false,
					Level:     typedLevel,
				},
			),
		),
	}
}

func (l FileLogger) Write(level slog.Level, message string, args ...any) {
	switch level {
	case slog.LevelDebug:
		l.Debug(message, args...)
	case slog.LevelInfo:
		l.Info(message, args...)
	case slog.LevelError:
		l.Logger.Error(message, args...)
	case slog.LevelWarn:
		l.Warn(message, args...)
	}
}

func (l FileLogger) Error(message string, args ...any) {
	l.Write(slog.LevelError, message, args)
}

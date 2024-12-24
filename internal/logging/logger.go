package logging

import (
	"log/slog"
	"os"
	"strconv"

	"clever.secure-onboard.com/internal/config"
)

type DefaultLogger struct {
	slog.Logger
}

func NewDefaultLogger(cfg config.LoggingInfo) DefaultLogger {
	level, err := strconv.Atoi(cfg.MinLogLevel)
	var typedLevel slog.Level
	if err != nil {
		typedLevel = slog.LevelDebug
	} else {
		typedLevel = slog.Level(level)
	}

	return DefaultLogger{
		Logger: *slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					AddSource: false,
					Level:     typedLevel,
				},
			),
		),
	}
}

func (l DefaultLogger) Write(level slog.Level, message string, args ...any) {
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

func (l DefaultLogger) Error(message string, args ...any) {
	l.Write(slog.LevelError, message, args)
}

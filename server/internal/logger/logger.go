package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	log *slog.Logger
}

func New(level string) *Logger {
	options := &slog.HandlerOptions{
		Level: parseLevel(level),
	}
	return &Logger{
		log: slog.New(slog.NewJSONHandler(os.Stdout, options)),
	}
}

func (l *Logger) Debug(msg string, attrs ...any) {
	l.log.Debug(msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...any) {
	l.log.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.log.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.log.Error(msg, attrs...)
}

func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{log: l.log.With(attrs...)}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

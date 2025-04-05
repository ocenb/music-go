package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/ocenb/music-go/user-service/internal/config"
)

type HandlerType string

const (
	TextHandler  HandlerType = "text"
	JSONHandler  HandlerType = "json"
	DefaultLevel slog.Level  = slog.LevelInfo
)

func Setup(cfg *config.Config) *slog.Logger {
	handlerType := cfg.LogHandler
	logLevel := cfg.LogLevel

	level := slog.LevelInfo

	if logLevel == int(slog.LevelDebug) || logLevel == int(slog.LevelInfo) || logLevel == int(slog.LevelWarn) || logLevel == int(slog.LevelError) {
		level = slog.Level(logLevel)
	} else {
		slog.Error("Invalid log level, using default level Info")
	}

	opts := &slog.HandlerOptions{Level: level}
	var log *slog.Logger

	switch handlerType {
	case string(TextHandler):
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case string(JSONHandler):
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	default:
		slog.Error("Invalid log handler type, using default TextHandler")
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	log = log.With(slog.String("env", cfg.Environment))

	return log
}

func NewForTest() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelError + 1}
	handler := slog.NewTextHandler(io.Discard, opts)

	return slog.New(handler)
}

package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

func New(level int, handlerType, environment string) *slog.Logger {
	logLevel := slog.Level(level)

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: environment == "local" || logLevel == slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if !ok {
					return a
				}
				if strings.HasSuffix(source.File, "internal/logger/adapter.go") ||
					strings.HasSuffix(source.File, "internal/middlewares/logging.go") {
					return slog.Attr{}
				}
			}
			return a
		},
	}

	var handler slog.Handler

	if handlerType == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler).With(slog.String("env", environment))
}

type loggerKeyType string

const loggerKey loggerKeyType = "logger"

func IntoContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}

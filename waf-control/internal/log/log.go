package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

// New creates a JSON slog logger at the given level.
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}

// WithRequestID returns a child logger with request_id attribute.
func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With("request_id", requestID)
}

// IntoContext stores the logger in context.
func IntoContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext retrieves the logger from context, or returns slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && v != nil {
		return v
	}
	return slog.Default()
}

package common

import (
	"context"
	"log/slog"
)

type slogContextKey struct{}

func InjectLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogContextKey{}, logger)
}

func ExtractLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(slogContextKey{}).(*slog.Logger)
	if !ok {
		return slog.New(new(disabledLoggerHandler))
	}

	return logger
}

package common

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

type slogContextKey struct{}

func InjectLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogContextKey{}, logger)
}

func GetLoggerInjectionUnaryInterceptor(
	logger *slog.Logger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = InjectLogger(ctx, logger)

		return handler(ctx, req)
	}
}

func ExtractLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(slogContextKey{}).(*slog.Logger)
	if !ok {
		return slog.New(new(disabledLoggerHandler))
	}

	return logger
}

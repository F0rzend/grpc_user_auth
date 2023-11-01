package common

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*disabledLoggerHandler)(nil)

type disabledLoggerHandler struct{}

func (*disabledLoggerHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (*disabledLoggerHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (*disabledLoggerHandler) WithAttrs(_ []slog.Attr) slog.Handler { //nolint:ireturn
	return new(disabledLoggerHandler)
}

func (*disabledLoggerHandler) WithGroup(_ string) slog.Handler { //nolint:ireturn
	return new(disabledLoggerHandler)
}

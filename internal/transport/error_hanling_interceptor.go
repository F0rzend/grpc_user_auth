package transport

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/F0rzend/grpc_user_auth/internal/common"
)

func ErrorHandlingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	resp, err := handler(ctx, req)
	switch {
	case err == nil:
		return resp, nil
	case status.Code(err) == codes.Unknown || status.Code(err) == codes.Internal:
		return nil, InternalError(ctx, err)
	default:
		return nil, err
	}
}

func InternalError(ctx context.Context, err error) error {
	logger := common.ExtractLogger(ctx)

	logger.ErrorContext(ctx, "Internal Error", slog.String("error", err.Error()))

	return status.Error(codes.Internal, "Internal Error")
}

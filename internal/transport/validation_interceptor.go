package transport

import (
	"context"

	"google.golang.org/grpc"

	"github.com/F0rzend/grpc_user_auth/proto"
)

func ValidationUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if validatable, ok := req.(proto.Validatable); ok {
		if err := validatable.Validate(); err != nil {
			//nolint:wrapcheck
			// In this interceptor, we intentionally do not wrap the error returned by the 'Validate' method
			// with additional context. This decision is made to shield users from the implementation details
			// of this interceptor and to avoid adding unnecessary context such as "error from the interceptor."
			// Users of this code should focus on the fact that the 'Validate' method's result indicates
			// whether the request is valid or not, without needing to worry about how the validation error
			// is handled internally.
			return nil, err
		}
	}

	return handler(ctx, req)
}

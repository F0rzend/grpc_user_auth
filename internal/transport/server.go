package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/proto"
)

type GRPCServer struct {
	server *grpc.Server
}

func NewGRPCServer(logger *slog.Logger, handlers *GRPCHandlers) *GRPCServer {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(common.GetLoggerInjectionUnaryInterceptor(logger)),
		grpc.ChainUnaryInterceptor(ErrorHandlingUnaryInterceptor),
		grpc.ChainUnaryInterceptor(handlers.BasicAuthUnaryInterceptor),
	)
	proto.RegisterUserServiceServer(server, handlers)

	return &GRPCServer{
		server: server,
	}
}

func (s *GRPCServer) ListenAndServe(ctx context.Context, address string) error {
	logger := common.ExtractLogger(ctx)

	listenerConfig := new(net.ListenConfig)
	listener, err := listenerConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.InfoContext(
		ctx,
		"grpc server is listening",
		slog.String("address", listener.Addr().String()),
		slog.String("network", listener.Addr().Network()),
	)
	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *GRPCServer) ShutdownOnContextDone(ctx context.Context) {
	<-ctx.Done()

	s.server.GracefulStop()
}

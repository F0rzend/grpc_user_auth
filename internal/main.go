package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/F0rzend/grpc_user_auth/internal/application"
	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/transport"
)

const ServerAddressEnv = "SERVER_ADDRESS"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx = common.InjectLogger(ctx, logger)

	repo := infrastructure.NewMemoryRepository()
	app := application.NewApplication(repo)
	handlers := transport.NewGRPCHandlers(app)
	grpcServer := transport.NewGRPCServer(handlers)

	address := os.Getenv(ServerAddressEnv)

	go grpcServer.ShutdownOnContextDone(ctx)
	if err := grpcServer.ListenAndServe(ctx, address); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}

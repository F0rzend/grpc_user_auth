package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/transport"
	"github.com/F0rzend/grpc_user_auth/internal/usecases"
)

const (
	ServerAddressEnv = "SERVER_ADDRESS"
	AdminUsernameEnv = "ADMIN_USERNAME"
	AdminEmailEnv    = "ADMIN_EMAIL"
	AdminPasswordEnv = "ADMIN_PASSWORD"
)

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
	app := usecases.NewUserUseCases(repo)
	handlers := transport.NewGRPCHandlers(app)
	grpcServer := transport.NewGRPCServer(logger, handlers)

	if err := createAdmin(app); err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	address := os.Getenv(ServerAddressEnv)

	go grpcServer.ShutdownOnContextDone(ctx)
	if err := grpcServer.ListenAndServe(ctx, address); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
}

func createAdmin(app *usecases.UserUseCases) error {
	adminUsername := os.Getenv(AdminUsernameEnv)
	adminEmail := os.Getenv(AdminEmailEnv)
	adminPassword := os.Getenv(AdminPasswordEnv)

	if adminUsername == "" || adminEmail == "" || adminPassword == "" {
		return fmt.Errorf("admin username, email and password must be set")
	}

	_, err := app.CreateUser(
		adminUsername,
		adminEmail,
		adminPassword,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	return nil
}

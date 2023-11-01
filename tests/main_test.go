package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DockerContext = "../"
	Dockerfile    = "Dockerfile"
	ContainerName = "app_test_container"

	ServerAddressEnv = "SERVER_ADDRESS"
	AdminUsernameEnv = "ADMIN_USERNAME"
	AdminEmailEnv    = "ADMIN_EMAIL"
	AdminPasswordEnv = "ADMIN_PASSWORD"
)

var appURL string //nolint:gochecknoglobals

type Config struct {
	ServerAddress string
	AdminUsername string
	AdminEmail    string
	AdminPassword string
}

func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()

	config, err := getConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	container, err := startTestEnvironment(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	appURL, err = getContainerURL(ctx, container, config.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Fatal(err)
	}

	os.Exit(exitCode)
}

func startTestEnvironment(ctx context.Context, config *Config) (testcontainers.Container, error) { //nolint:ireturn
	genericContainerRequest, err := getGenericContainerRequest(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get generic container request: %w", err)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get generic container: %w", err)
	}

	return container, nil
}

func getGenericContainerRequest(
	config *Config,
) (testcontainers.GenericContainerRequest, error) {
	_, port, err := net.SplitHostPort(config.ServerAddress)
	if err != nil {
		return testcontainers.GenericContainerRequest{}, fmt.Errorf("failed to split server address: %w", err)
	}

	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       DockerContext,
				Dockerfile:    Dockerfile,
				PrintBuildLog: true,
			},
			ExposedPorts: []string{port},
			Env: map[string]string{
				ServerAddressEnv: config.ServerAddress,
				AdminUsernameEnv: config.AdminUsername,
				AdminEmailEnv:    config.AdminEmail,
				AdminPasswordEnv: config.AdminPassword,
			},
			WaitingFor: wait.ForLog("grpc server is listening").WithStartupTimeout(5 * time.Minute),
			Name:       ContainerName,
		},
		Started: true,
		Reuse:   true,
	}, nil
}

func getConfigFromEnv() (*Config, error) {
	serverAddress, ok := os.LookupEnv(ServerAddressEnv)
	if !ok {
		return nil, fmt.Errorf("server address is not set")
	}

	adminUsername, ok := os.LookupEnv(AdminUsernameEnv)
	if !ok {
		return nil, fmt.Errorf("admin username is not set")
	}

	adminEmail, ok := os.LookupEnv(AdminEmailEnv)
	if !ok {
		return nil, fmt.Errorf("admin email is not set")
	}

	adminPassword, ok := os.LookupEnv(AdminPasswordEnv)
	if !ok {
		return nil, fmt.Errorf("admin password is not set")
	}

	return &Config{
		ServerAddress: serverAddress,
		AdminUsername: adminUsername,
		AdminEmail:    adminEmail,
		AdminPassword: adminPassword,
	}, nil
}

func getContainerURL(
	ctx context.Context,
	container testcontainers.Container,
	serverAddress string,
) (string, error) {
	_, rawServerPort, err := net.SplitHostPort(serverAddress)
	if err != nil {
		return "", fmt.Errorf("failed to split server address: %w", err)
	}

	serverPort := nat.Port(rawServerPort)

	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, serverPort)
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	addr := net.JoinHostPort(host, port.Port())

	return addr, nil
}

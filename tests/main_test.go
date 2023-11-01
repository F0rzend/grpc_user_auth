package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DockerContext = "../"
	Dockerfile    = "Dockerfile"
	ContainerName = "app_test_container"

	ServerAddressEnv = "SERVER_ADDRESS"
)

func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()
	container, err := startTestEnvironment(ctx)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Fatal(err)
	}

	os.Exit(exitCode)
}

func startTestEnvironment(ctx context.Context) (testcontainers.Container, error) { //nolint:ireturn
	serverAddress, ok := os.LookupEnv(ServerAddressEnv)
	if !ok {
		return nil, fmt.Errorf("server address is not set")
	}

	genericContainerRequest, err := getGenericContainerRequest(serverAddress)
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
	serverAddress string,
) (testcontainers.GenericContainerRequest, error) {
	_, port, err := net.SplitHostPort(serverAddress)
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
				ServerAddressEnv: serverAddress,
			},
			WaitingFor: wait.ForLog("grpc server is listening").WithStartupTimeout(5 * time.Minute),
			Name:       ContainerName,
		},
		Started: true,
		Reuse:   true,
	}, nil
}

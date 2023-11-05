package transport

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/F0rzend/grpc_user_auth/internal/common"
)

type Authenticator[UserModel any] struct {
	authFn func(username, password string) (UserModel, error)
}

type AuthFn[UserModel any] func(username, password string) (UserModel, error)

func NewAuthenticator[UserModel any](authFn AuthFn[UserModel]) *Authenticator[UserModel] {
	return &Authenticator[UserModel]{
		authFn: authFn,
	}
}

type authContextKey struct{}

type basicAuthCredentials struct {
	username string
	password string
}

func (a *Authenticator[UserModel]) BasicAuthUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	const expectedSchema = "Basic"

	token, err := extractAuthToken(ctx, expectedSchema)
	if err != nil {
		return nil, err
	}

	credentials, err := getBasicAuthCredentialsFromToken(token)
	if err != nil {
		return nil, err
	}

	user, err := a.authFn(credentials.username, credentials.password)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagNotFound):
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	default:
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	ctx = context.WithValue(ctx, authContextKey{}, user)

	return handler(ctx, req)
}

func extractAuthToken(ctx context.Context, expectedSchema string) (string, error) {
	const authorizationHeaderKey = "authorization"
	const authorizationHeaderSeparator = " "
	const authorizationHeaderComponentsNumber = 2

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Missing metadata")
	}

	authHeader, ok := md[authorizationHeaderKey]
	if !ok || len(authHeader) != 1 {
		return "", status.Error(codes.Unauthenticated, "Missing authorization token")
	}

	components := strings.SplitN(authHeader[0], authorizationHeaderSeparator, authorizationHeaderComponentsNumber)
	schema, token := components[0], components[1]

	if schema != expectedSchema {
		return "", status.Error(codes.Unauthenticated, "Invalid authorization schema")
	}

	return token, nil
}

func getBasicAuthCredentialsFromToken(token string) (*basicAuthCredentials, error) {
	const credentialsNumber = 2

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization token")
	}

	credentials := strings.SplitN(string(decoded), ":", credentialsNumber)
	if len(credentials) != credentialsNumber {
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization token")
	}

	return &basicAuthCredentials{
		username: credentials[0],
		password: credentials[1],
	}, nil
}

func (*Authenticator[UserModel]) GetAuthenticatedUser(ctx context.Context) (UserModel, error) { //nolint:ireturn
	var zero UserModel

	user, ok := ctx.Value(authContextKey{}).(UserModel)
	if !ok {
		return zero, status.Error(codes.Internal, "Failed to get authenticated user")
	}

	return user, nil
}

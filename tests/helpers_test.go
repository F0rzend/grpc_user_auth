package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/F0rzend/grpc_user_auth/proto"
)

var empty = new(emptypb.Empty) //nolint:gochecknoglobals

type CloseFn func()

func NewClient(t *testing.T, opts ...grpc.DialOption) (proto.UserServiceClient, CloseFn) { //nolint:ireturn
	t.Helper()

	conn, err := grpc.Dial(appURL, opts...)
	require.NoError(t, err)

	closeConnectionFn := func() {
		require.NoError(t, conn.Close())
	}

	return proto.NewUserServiceClient(conn), closeConnectionFn
}

func WithUnsecure() grpc.DialOption { //nolint:ireturn
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

func WithBasicAuth(username, password string) grpc.DialOption { //nolint:ireturn
	return grpc.WithPerRPCCredentials(&basicAuth{username, password})
}

type basicAuth struct {
	username string
	password string
}

func (b *basicAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	const AuthorizationHeaderKey = "authorization"
	const AuthorizationScheme = "Basic"

	joinedCredits := fmt.Sprintf("%s:%s", b.username, b.password)
	encodedCredits := base64.StdEncoding.EncodeToString([]byte(joinedCredits))

	return map[string]string{
		AuthorizationHeaderKey: fmt.Sprintf("%s %s", AuthorizationScheme, encodedCredits),
	}, nil
}

func (b *basicAuth) RequireTransportSecurity() bool {
	return false
}

func AssertErrorCode(t *testing.T, expected codes.Code, err error) {
	t.Helper()

	actual := status.Code(err)
	errorStatus, ok := status.FromError(err)
	require.True(t, ok, "error must be a status error")

	assert.Equal(
		t,
		expected,
		actual,
		fmt.Sprintf(
			"%q expected, but got %q with message: %q",
			expected,
			actual,
			errorStatus.Message(),
		),
	)
}

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/F0rzend/grpc_user_auth/proto"
)

func TestUnauthorized(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		invoke func(context.Context, proto.UserServiceClient) (any, error)
	}{
		{
			name: "GetAllUsers",
			invoke: func(ctx context.Context, client proto.UserServiceClient) (any, error) {
				return client.GetAllUsers(ctx, empty)
			},
		},
		{
			name: "GetUserByID",
			invoke: func(ctx context.Context, client proto.UserServiceClient) (any, error) {
				return client.GetUserByID(ctx, new(proto.GetUserRequest))
			},
		},
		{
			name: "CreateUser",
			invoke: func(ctx context.Context, client proto.UserServiceClient) (any, error) {
				return client.CreateUser(ctx, new(proto.CreateUserRequest))
			},
		},
		{
			name: "UpdateUser",
			invoke: func(ctx context.Context, client proto.UserServiceClient) (any, error) {
				return client.UpdateUser(ctx, new(proto.UpdateUserRequest))
			},
		},
		{
			name: "DeleteUser",
			invoke: func(ctx context.Context, client proto.UserServiceClient) (any, error) {
				return client.DeleteUser(ctx, new(proto.DeleteUserRequest))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client, closeConnection := NewClient(t, WithUnsecure())
			defer closeConnection()

			response, err := tc.invoke(ctx, client)

			assert.Nil(t, response)
			AssertErrorCode(t, codes.Unauthenticated, err)
		})
	}
}

func TestAdminWorkflow(t *testing.T) {
	t.Parallel()

	const (
		adminUsername = "admin"
		adminPassword = "admin"
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, closeConnection := NewClient(t, WithUnsecure(), WithBasicAuth(adminUsername, adminPassword))
	defer closeConnection()

	getAllUsersResponse, err := client.GetAllUsers(ctx, empty)
	assert.NoError(t, err)
	assert.NotEmpty(t, getAllUsersResponse.Users)

	user := struct {
		email    string
		username string
		password string
		admin    bool
	}{
		email:    "email@email.com",
		username: "username",
		password: "password",
		admin:    false,
	}

	createUserResponse, err := client.CreateUser(ctx, &proto.CreateUserRequest{
		Email:    user.email,
		Username: user.username,
		Password: user.password,
		Admin:    user.admin,
	})
	assert.NoError(t, err)
	assert.NotNil(t, createUserResponse)

	createdUserID := createUserResponse.Id

	getUserByIDResponse, err := client.GetUserByID(ctx, &proto.GetUserRequest{
		Id: createdUserID,
	})

	assert.NoError(t, err)
	assert.Equal(t, createdUserID, getUserByIDResponse.User.Id)
	assert.Equal(t, user.email, getUserByIDResponse.User.Email)
	assert.Equal(t, user.username, getUserByIDResponse.User.Username)
	assert.Equal(t, user.admin, getUserByIDResponse.User.Admin)

	newUsername := "newUsername"

	updateUserResponse, err := client.UpdateUser(ctx, &proto.UpdateUserRequest{
		Id:       createdUserID,
		Email:    user.email,
		Username: newUsername,
		Password: user.password,
		Admin:    user.admin,
	})
	assert.NoError(t, err)
	assert.Equal(t, empty, updateUserResponse)

	getUserByIDResponse, err = client.GetUserByID(ctx, &proto.GetUserRequest{
		Id: createdUserID,
	})
	assert.NoError(t, err)
	assert.Equal(t, newUsername, getUserByIDResponse.User.Username)

	deleteUserResponse, err := client.DeleteUser(ctx, &proto.DeleteUserRequest{
		Id: createdUserID,
	})
	assert.NoError(t, err)
	assert.Equal(t, empty, deleteUserResponse)

	getUserByIDResponse, err = client.GetUserByID(ctx, &proto.GetUserRequest{
		Id: createdUserID,
	})
	assert.Nil(t, getUserByIDResponse)
	AssertErrorCode(t, codes.NotFound, err)
}

func TestUserWorkflow(t *testing.T) {
	t.Parallel()

	const (
		email    = "user@email.com"
		username = "user"
		password = "user"
	)

	createUser(t, email, username, password)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, closeConnection := NewClient(t, WithUnsecure(), WithBasicAuth(username, password))
	defer closeConnection()

	response, err := client.GetAllUsers(ctx, empty)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Users)
}

func createUser(t *testing.T, email, username, password string) {
	t.Helper()

	const (
		adminUsername = "admin"
		adminPassword = "admin"
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	admin, closeAdminConnection := NewClient(t, WithUnsecure(), WithBasicAuth(adminUsername, adminPassword))
	defer closeAdminConnection()

	resp, err := admin.CreateUser(ctx, &proto.CreateUserRequest{
		Email:    email,
		Username: username,
		Password: password,
		Admin:    false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

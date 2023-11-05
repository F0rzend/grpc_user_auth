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
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/models"
	"github.com/F0rzend/grpc_user_auth/internal/usecases"
	"github.com/F0rzend/grpc_user_auth/proto"
)

var _ proto.UserServiceServer = (*GRPCHandlers)(nil)

type GRPCHandlers struct {
	proto.UnimplementedUserServiceServer

	userUseCases *usecases.UserUseCases
}

func NewGRPCHandlers(userUseCases *usecases.UserUseCases) *GRPCHandlers {
	return &GRPCHandlers{
		UnimplementedUserServiceServer: proto.UnimplementedUserServiceServer{},

		userUseCases: userUseCases,
	}
}

var empty = new(emptypb.Empty) //nolint:gochecknoglobals

func (h *GRPCHandlers) CreateUser(
	ctx context.Context,
	request *proto.CreateUserRequest,
) (*proto.CreateUserResponse, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

	cmd, err := usecases.NewCreateUserCommand(
		request.Username,
		request.Email,
		request.Password,
		request.Admin,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation Error")
	}

	id, err := h.userUseCases.CreateUser(cmd)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	default:
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &proto.CreateUserResponse{
		Id: id.String(),
	}, nil
}

func (h *GRPCHandlers) GetAllUsers(_ context.Context, _ *emptypb.Empty) (*proto.GetAllUsersResponse, error) {
	users, err := h.userUseCases.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	response := &proto.GetAllUsersResponse{
		Users: make([]*proto.User, len(users)),
	}
	for i, user := range users {
		response.Users[i] = &proto.User{
			Id:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Admin:    user.Admin,
		}
	}

	return response, nil
}

func (h *GRPCHandlers) GetUserByID(ctx context.Context, request *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

	query, err := usecases.NewGetUserByIDQuery(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation Error")
	}

	user, err := h.userUseCases.GetUserByID(query)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagNotFound):
		return nil, status.Error(codes.NotFound, "User not found")
	default:
		return nil, fmt.Errorf("failed to get user by id %q: %w", request.Id, err)
	}

	response := &proto.GetUserResponse{
		User: &proto.User{
			Id:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Admin:    user.Admin,
		},
	}

	return response, nil
}

func (h *GRPCHandlers) UpdateUser(ctx context.Context, request *proto.UpdateUserRequest) (*emptypb.Empty, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

	cmd, err := usecases.NewUpdateUserCommand(
		request.Id,
		request.Username,
		request.Email,
		request.Password,
		request.Admin,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation Error")
	}

	err = h.userUseCases.UpdateUser(cmd)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagNotFound):
		return nil, status.Error(codes.NotFound, "User not found")
	default:
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return empty, nil
}

func (h *GRPCHandlers) DeleteUser(ctx context.Context, request *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

	cmd, err := usecases.NewDeleteUserCommand(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation Error")
	}

	err = h.userUseCases.DeleteUser(cmd)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagNotFound):
		return nil, status.Error(codes.NotFound, "User not found")
	default:
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return empty, nil
}

type authContextKey struct{}

func (h *GRPCHandlers) BasicAuthUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	const expectedSchema = "Basic"
	const authorizationHeaderKey = "authorization"
	const credentialsNumber = 2

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Missing metadata")
	}

	authHeader, ok := md[authorizationHeaderKey]
	if !ok || len(authHeader) != 1 {
		return nil, status.Error(codes.Unauthenticated, "Missing authorization token")
	}

	components := strings.Split(authHeader[0], " ")
	schema, token := components[0], components[1]

	if schema != expectedSchema {
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization schema")
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization token")
	}

	credentials := strings.SplitN(string(decoded), ":", credentialsNumber)
	username, password := credentials[0], credentials[1]

	user, err := h.userUseCases.AuthenticateUser(username, password)
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	ctx = context.WithValue(ctx, authContextKey{}, user)

	return handler(ctx, req)
}

func (h *GRPCHandlers) GetAuthenticatedUser(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(authContextKey{}).(*models.User)
	if !ok {
		return nil, status.Error(codes.Internal, "Failed to get authenticated user")
	}

	return user, nil
}

func (h *GRPCHandlers) AdminOnly(ctx context.Context) error {
	user, err := h.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated user: %w", err)
	}

	if !user.Admin {
		return status.Error(codes.PermissionDenied, "Only admins can perform this action")
	}

	return nil
}

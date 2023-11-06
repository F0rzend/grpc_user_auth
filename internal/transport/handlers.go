package transport

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
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

	userUseCases  *usecases.UserUseCases
	authenticator *Authenticator[*models.User]
}

func NewGRPCHandlers(
	userUseCases *usecases.UserUseCases,
	authenticator *Authenticator[*models.User],
) *GRPCHandlers {
	return &GRPCHandlers{
		UnimplementedUserServiceServer: proto.UnimplementedUserServiceServer{},

		userUseCases:  userUseCases,
		authenticator: authenticator,
	}
}

var empty = new(emptypb.Empty) //nolint:gochecknoglobals

func (h *GRPCHandlers) CreateUser(
	ctx context.Context,
	request *proto.CreateUserRequest,
) (*proto.CreateUserResponse, error) {
	err := h.adminOnly(ctx)
	if err != nil {
		return nil, err
	}

	cmd, err := usecases.NewCreateUserCommand(
		request.Username,
		request.Email,
		request.Password,
		request.Admin,
	)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagInvalidArgument):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, fmt.Errorf("failed to create CreateUser command: %w", err)
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
	err := h.adminOnly(ctx)
	if err != nil {
		return nil, err
	}

	query, err := usecases.NewGetUserByIDQuery(request.Id)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagInvalidArgument):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, fmt.Errorf("failed to create GetUserByID query: %w", err)
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
	err := h.adminOnly(ctx)
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
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagInvalidArgument):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, fmt.Errorf("failed to create UpdateUser command: %w", err)
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
	err := h.adminOnly(ctx)
	if err != nil {
		return nil, err
	}

	cmd, err := usecases.NewDeleteUserCommand(request.Id)
	switch {
	case err == nil:
	case common.IsFlaggedError(err, common.FlagInvalidArgument):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, fmt.Errorf("failed to create DeleteUser command: %w", err)
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

func (h *GRPCHandlers) adminOnly(ctx context.Context) error {
	user, err := h.authenticator.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated user: %w", err)
	}

	if !user.Admin {
		return status.Error(codes.PermissionDenied, "Only admins can perform this action")
	}

	return nil
}

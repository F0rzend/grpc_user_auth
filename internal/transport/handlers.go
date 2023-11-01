package transport

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/F0rzend/grpc_user_auth/internal/application"
	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/proto"
)

var _ proto.UserServiceServer = (*GRPCHandlers)(nil)

type GRPCHandlers struct {
	proto.UnimplementedUserServiceServer

	app *application.Application
}

func NewGRPCHandlers(app *application.Application) *GRPCHandlers {
	return &GRPCHandlers{
		UnimplementedUserServiceServer: proto.UnimplementedUserServiceServer{},

		app: app,
	}
}

var empty = new(emptypb.Empty) //nolint:gochecknoglobals

func (h *GRPCHandlers) CreateUser(_ context.Context, request *proto.CreateUserRequest) (*emptypb.Empty, error) {
	parsedUUID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid UUID")
	}

	_, err = mail.ParseAddress(request.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid email")
	}

	err = h.app.CreateUser(
		parsedUUID,
		request.Username,
		request.Email,
		request.Password,
		request.Admin,
	)
	if common.IsFlaggedError(err, common.FlagAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return empty, nil
}

func (h *GRPCHandlers) GetAllUsers(_ context.Context, _ *emptypb.Empty) (*proto.GetAllUsersResponse, error) {
	users, err := h.app.GetAllUsers()
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

func (h *GRPCHandlers) GetUserByID(_ context.Context, request *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	parsedUUID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid UUID")
	}

	user, err := h.app.GetUserByID(parsedUUID)
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %q: %w", parsedUUID, err)
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

func (h *GRPCHandlers) UpdateUser(_ context.Context, request *proto.UpdateUserRequest) (*emptypb.Empty, error) {
	parsedUUID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid UUID")
	}

	_, err = mail.ParseAddress(request.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid email")
	}

	err = h.app.UpdateUser(
		parsedUUID,
		request.Username,
		request.Email,
		request.Password,
		request.Admin,
	)
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return empty, nil
}

func (h *GRPCHandlers) DeleteUser(_ context.Context, request *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	parsedUUID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid UUID")
	}

	err = h.app.DeleteUser(parsedUUID)
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return empty, nil
}

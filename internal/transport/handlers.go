package transport

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/F0rzend/grpc_user_auth/internal/application"
	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/models"
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

func (h *GRPCHandlers) CreateUser(ctx context.Context, request *proto.CreateUserRequest) (*emptypb.Empty, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

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

func (h *GRPCHandlers) GetUserByID(ctx context.Context, request *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

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

func (h *GRPCHandlers) UpdateUser(ctx context.Context, request *proto.UpdateUserRequest) (*emptypb.Empty, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

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

func (h *GRPCHandlers) DeleteUser(ctx context.Context, request *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	err := h.AdminOnly(ctx)
	if err != nil {
		return nil, err
	}

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

	logger := common.ExtractLogger(ctx)

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

	user, err := h.app.AuthenticateUser(username, password)
	if common.IsFlaggedError(err, common.FlagNotFound) {
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}
	if err != nil {
		logger.ErrorContext(ctx, "Failed to authenticate user", slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, "Failed to authenticate user")
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

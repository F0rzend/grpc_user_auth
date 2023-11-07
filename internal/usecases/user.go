package usecases

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/models"
)

type UserUseCases struct {
	repo *infrastructure.Repository
}

func NewUserUseCases(repo *infrastructure.Repository) *UserUseCases {
	return &UserUseCases{
		repo: repo,
	}
}

type CreateUserCommand struct {
	username string
	email    string
	password string
	admin    bool
}

func NewCreateUserCommand(
	username string,
	email string,
	password string,
	admin bool,
) *CreateUserCommand {
	return &CreateUserCommand{
		username: username,
		email:    email,
		password: password,
		admin:    admin,
	}
}

func (u *UserUseCases) CreateUser(cmd *CreateUserCommand) (uuid.UUID, error) {
	_, err := u.repo.GetByUsername(cmd.username)
	switch {
	case errors.Is(err, common.ErrNotFound):
	case err == nil:
		return uuid.UUID{}, fmt.Errorf("%w: user with username %q already exists", common.ErrAlreadyExists, cmd.username)
	default:
		return uuid.UUID{}, fmt.Errorf("failed to get user by username %q: %w", cmd.username, err)
	}

	id := uuid.New()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           id,
		Username:     cmd.username,
		Email:        cmd.email,
		PasswordHash: passwordHash,
		Admin:        cmd.admin,
	}

	err = u.repo.Save(user)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

type GetUserByIDQuery struct {
	id uuid.UUID
}

func NewGetUserByIDQuery(id string) (*GetUserByIDQuery, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id %q: %w", id, err)
	}

	return &GetUserByIDQuery{
		id: userUUID,
	}, nil
}

func (u *UserUseCases) GetUserByID(query *GetUserByIDQuery) (*models.User, error) {
	user, err := u.repo.GetByID(query.id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %q: %w", query.id, err)
	}

	return user, nil
}

func (u *UserUseCases) GetAllUsers() ([]*models.User, error) {
	users, err := u.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

type UpdateUserCommand struct {
	id       uuid.UUID
	username string
	email    string
	password string
	admin    bool
}

func NewUpdateUserCommand(
	id string,
	username string,
	email string,
	password string,
	admin bool,
) (*UpdateUserCommand, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id %q: %w", id, err)
	}

	return &UpdateUserCommand{
		id:       userUUID,
		username: username,
		email:    email,
		password: password,
		admin:    admin,
	}, nil
}

func (u *UserUseCases) UpdateUser(cmd *UpdateUserCommand) error {
	_, err := u.repo.GetByID(cmd.id)
	if err != nil {
		return fmt.Errorf("failed to get user by id %q: %w", cmd.id, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           cmd.id,
		Username:     cmd.username,
		Email:        cmd.email,
		PasswordHash: passwordHash,
		Admin:        cmd.admin,
	}

	err = u.repo.Save(user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

type DeleteUserCommand struct {
	id uuid.UUID
}

func NewDeleteUserCommand(id string) (*DeleteUserCommand, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id %q: %w", id, err)
	}

	return &DeleteUserCommand{
		id: userUUID,
	}, nil
}

func (u *UserUseCases) DeleteUser(cmd *DeleteUserCommand) error {
	err := u.repo.Delete(cmd.id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (u *UserUseCases) AuthenticateUser(username string, rawPassword string) (*models.User, error) {
	user, err := u.repo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username %q: %w", username, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(rawPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
	}

	return user, nil
}

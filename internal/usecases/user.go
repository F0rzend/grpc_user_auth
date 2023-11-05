package usecases

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/models"
)

type UserUseCases struct {
	repo *infrastructure.MemoryRepository
}

func NewUserUseCases(repo *infrastructure.MemoryRepository) *UserUseCases {
	return &UserUseCases{
		repo: repo,
	}
}

func (u *UserUseCases) CreateUser(
	username string,
	email string,
	rawPassword string,
	admin bool,
) (uuid.UUID, error) {
	_, err := u.repo.GetByUsername(username)
	switch {
	case common.IsFlaggedError(err, common.FlagNotFound):
	case err == nil:
		return uuid.UUID{}, common.FlagError(
			fmt.Errorf("user with username %q already exists", username),
			common.FlagAlreadyExists,
		)
	default:
		return uuid.UUID{}, fmt.Errorf("failed to get user by username %q: %w", username, err)
	}

	id := uuid.New()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Admin:        admin,
	}

	err = u.repo.Save(user)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

func (u *UserUseCases) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %q: %w", id, err)
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

func (u *UserUseCases) UpdateUser(
	id uuid.UUID,
	username string,
	email string,
	rawPassword string,
	admin bool,
) error {
	_, err := u.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user by id %q: %w", id, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Admin:        admin,
	}

	err = u.repo.Save(user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (u *UserUseCases) DeleteUser(id uuid.UUID) error {
	err := u.repo.Delete(id)
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

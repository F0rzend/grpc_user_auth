package application

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/models"
)

type Application struct {
	repo *infrastructure.MemoryRepository
}

func NewApplication(repo *infrastructure.MemoryRepository) *Application {
	return &Application{
		repo: repo,
	}
}

func (a *Application) CreateUser(
	id uuid.UUID,
	username string,
	email string,
	rawPassword string,
	admin bool,
) error {
	_, err := a.repo.GetByID(id)
	if err == nil {
		return common.FlagError(fmt.Errorf("user with id %q already exists", id), common.FlagAlreadyExists)
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

	err = a.repo.Save(user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (a *Application) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := a.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %q: %w", id, err)
	}

	return user, nil
}

func (a *Application) GetAllUsers() ([]*models.User, error) {
	users, err := a.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

func (a *Application) UpdateUser(
	id uuid.UUID,
	username string,
	email string,
	rawPassword string,
	admin bool,
) error {
	_, err := a.repo.GetByID(id)
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

	err = a.repo.Save(user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (a *Application) DeleteUser(id uuid.UUID) error {
	err := a.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (a *Application) AuthenticateUser(username string, rawPassword string) (*models.User, error) {
	user, err := a.repo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username %q: %w", username, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(rawPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
	}

	return user, nil
}

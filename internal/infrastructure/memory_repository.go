package infrastructure

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/models"
)

type MemoryRepository struct {
	users sync.Map
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: sync.Map{},
	}
}

func (r *MemoryRepository) Save(user *models.User) error {
	r.users.Store(user.ID.String(), user)

	return nil
}

func (r *MemoryRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if value, ok := r.users.Load(id.String()); ok {
		user, ok := value.(*models.User)
		if !ok {
			return nil, fmt.Errorf("unexpected value %+#v in users map", value)
		}

		return user, nil
	}

	return nil, common.FlagError(fmt.Errorf("user with id %q not found", id), common.FlagNotFound)
}

func (r *MemoryRepository) GetByUsername(username string) (*models.User, error) {
	var found *models.User

	r.users.Range(func(key, value interface{}) bool {
		user, ok := value.(*models.User)
		if !ok {
			return false
		}

		if user.Username == username {
			found = user

			return false
		}

		return true
	})

	if found == nil {
		return nil, common.FlagError(fmt.Errorf("user with username %q not found", username), common.FlagNotFound)
	}

	return found, nil
}

func (r *MemoryRepository) GetAll() ([]*models.User, error) {
	users := make([]*models.User, 0)

	r.users.Range(func(key, value interface{}) bool {
		user, ok := value.(*models.User)
		if !ok {
			return false
		}

		users = append(users, user)

		return true
	})

	return users, nil
}

func (r *MemoryRepository) Delete(id uuid.UUID) error {
	existing, err := r.GetByID(id)
	if err != nil {
		return err
	}

	r.users.Delete(existing.ID.String())

	return nil
}

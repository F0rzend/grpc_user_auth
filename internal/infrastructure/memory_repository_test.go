package infrastructure_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/F0rzend/grpc_user_auth/internal/common"
	"github.com/F0rzend/grpc_user_auth/internal/infrastructure"
	"github.com/F0rzend/grpc_user_auth/internal/models"
)

func TestMemoryRepository_Save(t *testing.T) {
	t.Parallel()

	testUser := createTestUser(t)
	sut := infrastructure.NewMemoryRepository()

	err := sut.Save(testUser)
	assert.NoError(t, err)
}

func TestMemoryRepository_GetByID(t *testing.T) {
	t.Parallel()

	sut := infrastructure.NewMemoryRepository()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testUser := createTestUser(t)
		err := sut.Save(testUser)
		require.NoError(t, err)

		user, err := sut.GetByID(testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		user, err := sut.GetByID(uuid.New())

		assertErrorFlag(t, err, common.FlagNotFound)
		assert.Nil(t, user)
	})
}

func TestMemoryRepository_GetAll(t *testing.T) {
	t.Parallel()

	const usersCount = 10
	savedIDs := make([]uuid.UUID, usersCount)

	sut := infrastructure.NewMemoryRepository()

	for i := 0; i < usersCount; i++ {
		user := createTestUser(t)
		err := sut.Save(user)
		require.NoError(t, err)
		savedIDs[i] = user.ID
	}

	users, err := sut.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, usersCount)

	for _, user := range users {
		assert.Contains(t, savedIDs, user.ID)
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	t.Parallel()

	sut := infrastructure.NewMemoryRepository()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testUser := createTestUser(t)
		err := sut.Save(testUser)
		require.NoError(t, err)

		actual, err := sut.GetByID(testUser.ID)
		require.NoError(t, err)
		require.Equal(t, testUser, actual)

		err = sut.Delete(testUser.ID)
		assert.NoError(t, err)

		actual, err = sut.GetByID(testUser.ID)
		assertErrorFlag(t, err, common.FlagNotFound)
		assert.Nil(t, actual)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		err := sut.Delete(uuid.New())

		assertErrorFlag(t, err, common.FlagNotFound)
	})
}

func createTestUser(t *testing.T) *models.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)

	return &models.User{
		ID:           uuid.New(),
		Username:     "test",
		Email:        "test.email@gmail.com",
		PasswordHash: passwordHash,
		Admin:        false,
	}
}

func assertErrorFlag(t *testing.T, err error, expectedFlag common.Flag) {
	t.Helper()

	flagged := new(common.FlaggedError)
	assert.ErrorAs(t, err, flagged)
	assert.Equal(t, expectedFlag, flagged.Flag())
}

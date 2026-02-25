package sqlite3_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/auth"
	"github.com/nasermirzaei89/scribble/db/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	ctx, db := newTestDB(t)

	repo := sqlite3.NewUserRepository(db)

	t.Run("Find not found", func(t *testing.T) {
		userID := uuid.NewString()

		_, err := repo.Find(ctx, userID)

		var userNotFoundErr *auth.UserNotFoundError

		assert.ErrorAs(t, err, &userNotFoundErr)
		assert.Equal(t, userID, userNotFoundErr.ID)
	})

	t.Run("FindByUsername not found", func(t *testing.T) {
		username := "missing-username"
		_, err := repo.FindByUsername(ctx, username)

		var userByUsernameNotFoundErr *auth.UserByUsernameNotFoundError

		assert.ErrorAs(t, err, &userByUsernameNotFoundErr)
		assert.Equal(t, username, userByUsernameNotFoundErr.Username)
	})

	t.Run("Insert and find by ID and username", func(t *testing.T) {
		registeredAt := time.Date(2026, 2, 24, 10, 30, 0, 0, time.UTC)

		user := &auth.User{
			ID:           uuid.NewString(),
			Username:     "johndoe",
			PasswordHash: "password-hash",
			RegisteredAt: registeredAt,
		}

		err := repo.Insert(ctx, user)
		require.NoError(t, err)

		foundByID, err := repo.Find(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, foundByID.ID)
		assert.Equal(t, user.Username, foundByID.Username)
		assert.Equal(t, user.PasswordHash, foundByID.PasswordHash)
		assert.True(t, foundByID.RegisteredAt.Equal(user.RegisteredAt))

		foundByUsername, err := repo.FindByUsername(ctx, user.Username)
		require.NoError(t, err)
		assert.Equal(t, user.ID, foundByUsername.ID)
		assert.Equal(t, user.Username, foundByUsername.Username)
	})

	t.Run("Insert duplicate username", func(t *testing.T) {
		user := &auth.User{
			ID:           uuid.NewString(),
			Username:     "johndoe",
			PasswordHash: "another-hash",
			RegisteredAt: time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
		}

		err := repo.Insert(ctx, user)
		require.Error(t, err)
	})
}

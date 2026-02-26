package sqlite3_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/authentication"
	"github.com/nasermirzaei89/scribble/database/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository(t *testing.T) {
	ctx, db := newTestDB(t)

	userRepo := sqlite3.NewUserRepository(db)
	sessionRepo := sqlite3.NewSessionRepository(db)

	user := &authentication.User{
		ID:           uuid.NewString(),
		Username:     "session-user-" + uuid.NewString(),
		PasswordHash: "password-hash",
		RegisteredAt: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
	}

	err := userRepo.Insert(ctx, user)
	require.NoError(t, err)

	t.Run("Find not found", func(t *testing.T) {
		sessionID := uuid.NewString()

		_, err := sessionRepo.Find(ctx, sessionID)

		var sessionNotFoundErr *authentication.SessionNotFoundError

		require.ErrorAs(t, err, &sessionNotFoundErr)
		assert.Equal(t, sessionID, sessionNotFoundErr.ID)
	})

	t.Run("Insert and find", func(t *testing.T) {
		session := &authentication.Session{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			CreatedAt: time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
			ExpiresAt: time.Date(2026, 2, 25, 11, 0, 0, 0, time.UTC),
		}

		err := sessionRepo.Insert(ctx, session)
		require.NoError(t, err)

		found, err := sessionRepo.Find(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, found.ID)
		assert.Equal(t, session.UserID, found.UserID)
		assert.True(t, found.CreatedAt.Equal(session.CreatedAt))
		assert.True(t, found.ExpiresAt.Equal(session.ExpiresAt))
	})

	t.Run("Delete existing", func(t *testing.T) {
		session := &authentication.Session{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			CreatedAt: time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC),
			ExpiresAt: time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC),
		}

		err := sessionRepo.Insert(ctx, session)
		require.NoError(t, err)

		err = sessionRepo.Delete(ctx, session.ID)
		require.NoError(t, err)

		_, err = sessionRepo.Find(ctx, session.ID)

		var sessionNotFoundErr *authentication.SessionNotFoundError

		require.ErrorAs(t, err, &sessionNotFoundErr)
		assert.Equal(t, session.ID, sessionNotFoundErr.ID)
	})

	t.Run("Delete not found", func(t *testing.T) {
		sessionID := uuid.NewString()

		err := sessionRepo.Delete(ctx, sessionID)

		var sessionNotFoundErr *authentication.SessionNotFoundError

		require.ErrorAs(t, err, &sessionNotFoundErr)
		assert.Equal(t, sessionID, sessionNotFoundErr.ID)
	})
}

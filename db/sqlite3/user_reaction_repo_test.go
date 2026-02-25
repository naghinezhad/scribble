package sqlite3_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/auth"
	"github.com/nasermirzaei89/scribble/db/sqlite3"
	"github.com/nasermirzaei89/scribble/reactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserReactionRepository(t *testing.T) {
	ctx, db := newTestDB(t)

	userRepo := sqlite3.NewUserRepository(db)
	repo := sqlite3.NewUserReactionRepository(db)

	user1 := &auth.User{
		ID:           uuid.NewString(),
		Username:     "reaction-user-" + uuid.NewString(),
		PasswordHash: "password-hash",
		RegisteredAt: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
	}

	user2 := &auth.User{
		ID:           uuid.NewString(),
		Username:     "reaction-user-" + uuid.NewString(),
		PasswordHash: "password-hash",
		RegisteredAt: time.Date(2026, 2, 24, 10, 1, 0, 0, time.UTC),
	}

	err := userRepo.Insert(ctx, user1)
	require.NoError(t, err)

	err = userRepo.Insert(ctx, user2)
	require.NoError(t, err)

	targetID := uuid.NewString()

	t.Run("FindByUserTarget not found", func(t *testing.T) {
		_, err := repo.FindByUserTarget(ctx, reactions.TargetTypePost, targetID, user1.ID)

		var userReactionNotFoundErr *reactions.UserReactionNotFoundError

		assert.ErrorAs(t, err, &userReactionNotFoundErr)
		assert.Equal(t, reactions.TargetTypePost, userReactionNotFoundErr.TargetType)
		assert.Equal(t, targetID, userReactionNotFoundErr.TargetID)
		assert.Equal(t, user1.ID, userReactionNotFoundErr.UserID)
	})

	t.Run("Upsert and find", func(t *testing.T) {
		reaction := &reactions.UserReaction{
			TargetType: reactions.TargetTypePost,
			TargetID:   targetID,
			UserID:     user1.ID,
			Emoji:      "🔥",
			CreatedAt:  time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
		}

		err := repo.Upsert(ctx, reaction)
		require.NoError(t, err)

		found, err := repo.FindByUserTarget(ctx, reaction.TargetType, reaction.TargetID, reaction.UserID)
		require.NoError(t, err)
		assert.Equal(t, reaction.TargetType, found.TargetType)
		assert.Equal(t, reaction.TargetID, found.TargetID)
		assert.Equal(t, reaction.UserID, found.UserID)
		assert.Equal(t, reaction.Emoji, found.Emoji)
		assert.True(t, found.CreatedAt.Equal(reaction.CreatedAt))
	})

	t.Run("Upsert updates existing reaction", func(t *testing.T) {
		reaction := &reactions.UserReaction{
			TargetType: reactions.TargetTypePost,
			TargetID:   targetID,
			UserID:     user1.ID,
			Emoji:      "❤️",
			CreatedAt:  time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC),
		}

		err := repo.Upsert(ctx, reaction)
		require.NoError(t, err)

		found, err := repo.FindByUserTarget(ctx, reaction.TargetType, reaction.TargetID, reaction.UserID)
		require.NoError(t, err)
		assert.Equal(t, "❤️", found.Emoji)
		assert.True(t, found.CreatedAt.Equal(reaction.CreatedAt))
	})

	t.Run("CountByTarget groups by emoji", func(t *testing.T) {
		err := repo.Upsert(ctx, &reactions.UserReaction{
			TargetType: reactions.TargetTypePost,
			TargetID:   targetID,
			UserID:     user2.ID,
			Emoji:      "❤️",
			CreatedAt:  time.Date(2026, 2, 24, 13, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)

		otherTargetID := uuid.NewString()

		err = repo.Upsert(ctx, &reactions.UserReaction{
			TargetType: reactions.TargetTypePost,
			TargetID:   otherTargetID,
			UserID:     user1.ID,
			Emoji:      "🔥",
			CreatedAt:  time.Date(2026, 2, 24, 13, 1, 0, 0, time.UTC),
		})
		require.NoError(t, err)

		counts, err := repo.CountByTarget(ctx, reactions.TargetTypePost, targetID)
		require.NoError(t, err)
		assert.Equal(t, 2, counts["❤️"])
		assert.Equal(t, 0, counts["🔥"])
	})

	t.Run("DeleteByUserTarget", func(t *testing.T) {
		err := repo.DeleteByUserTarget(ctx, reactions.TargetTypePost, targetID, user1.ID)
		require.NoError(t, err)

		_, err = repo.FindByUserTarget(ctx, reactions.TargetTypePost, targetID, user1.ID)

		var userReactionNotFoundErr *reactions.UserReactionNotFoundError

		assert.ErrorAs(t, err, &userReactionNotFoundErr)

		err = repo.DeleteByUserTarget(ctx, reactions.TargetTypePost, targetID, user1.ID)
		require.NoError(t, err)
	})
}
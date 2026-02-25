package sqlite3_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/auth"
	"github.com/nasermirzaei89/scribble/contents"
	"github.com/nasermirzaei89/scribble/db/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRepository(t *testing.T) {
	ctx, db := newTestDB(t)

	userRepo := sqlite3.NewUserRepository(db)
	postRepo := sqlite3.NewPostRepository(db)

	user := &auth.User{
		ID:           uuid.NewString(),
		Username:     "post-user-" + uuid.NewString(),
		PasswordHash: "password-hash",
		RegisteredAt: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
	}

	err := userRepo.Insert(ctx, user)
	require.NoError(t, err)

	t.Run("List empty", func(t *testing.T) {
		posts, err := postRepo.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("Find not found", func(t *testing.T) {
		postID := uuid.NewString()

		_, err := postRepo.Find(ctx, postID)

		var postNotFoundErr contents.PostNotFoundError

		assert.ErrorAs(t, err, &postNotFoundErr)
		assert.Equal(t, postID, postNotFoundErr.ID)
	})

	t.Run("Insert find and list", func(t *testing.T) {
		post1 := &contents.Post{
			ID:        uuid.NewString(),
			AuthorID:  user.ID,
			Content:   "post content 1",
			CreatedAt: time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
		}

		post2 := &contents.Post{
			ID:        uuid.NewString(),
			AuthorID:  user.ID,
			Content:   "post content 2",
			CreatedAt: time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC),
		}

		err := postRepo.Insert(ctx, post1)
		require.NoError(t, err)

		err = postRepo.Insert(ctx, post2)
		require.NoError(t, err)

		found, err := postRepo.Find(ctx, post1.ID)
		require.NoError(t, err)
		assert.Equal(t, post1.ID, found.ID)
		assert.Equal(t, post1.AuthorID, found.AuthorID)
		assert.Equal(t, post1.Content, found.Content)
		assert.True(t, found.CreatedAt.Equal(post1.CreatedAt))

		posts, err := postRepo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, posts, 2)

		postIDs := []string{posts[0].ID, posts[1].ID}
		assert.Contains(t, postIDs, post1.ID)
		assert.Contains(t, postIDs, post2.ID)
	})
}
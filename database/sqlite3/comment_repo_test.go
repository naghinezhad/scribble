package sqlite3_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/authentication"
	"github.com/nasermirzaei89/scribble/contents"
	"github.com/nasermirzaei89/scribble/database/sqlite3"
	"github.com/nasermirzaei89/scribble/discuss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentRepository(t *testing.T) {
	ctx, db := newTestDB(t)

	userRepo := sqlite3.NewUserRepository(db)
	postRepo := sqlite3.NewPostRepository(db)
	commentRepo := sqlite3.NewCommentRepository(db)

	user := &authentication.User{
		ID:           uuid.NewString(),
		Username:     "comment-user-" + uuid.NewString(),
		PasswordHash: "password-hash",
		RegisteredAt: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
	}

	err := userRepo.Insert(ctx, user)
	require.NoError(t, err)

	post1 := &contents.Post{
		ID:        uuid.NewString(),
		AuthorID:  user.ID,
		Content:   "post for comments 1",
		CreatedAt: time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
	}

	post2 := &contents.Post{
		ID:        uuid.NewString(),
		AuthorID:  user.ID,
		Content:   "post for comments 2",
		CreatedAt: time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC),
	}

	err = postRepo.Insert(ctx, post1)
	require.NoError(t, err)

	err = postRepo.Insert(ctx, post2)
	require.NoError(t, err)

	t.Run("List and count empty", func(t *testing.T) {
		comments, err := commentRepo.List(ctx, &discuss.ListCommentsParams{})
		require.NoError(t, err)
		assert.Empty(t, comments)

		count, err := commentRepo.Count(ctx, &discuss.CountCommentsParams{})
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Insert list and count with filter", func(t *testing.T) {
		comment1 := &discuss.Comment{
			ID:        uuid.NewString(),
			PostID:    post1.ID,
			AuthorID:  user.ID,
			Content:   "first comment",
			CreatedAt: time.Date(2026, 2, 24, 13, 0, 0, 0, time.UTC),
		}

		replyTo := comment1.ID
		comment2 := &discuss.Comment{
			ID:        uuid.NewString(),
			PostID:    post1.ID,
			AuthorID:  user.ID,
			ReplyTo:   &replyTo,
			Content:   "reply comment",
			CreatedAt: time.Date(2026, 2, 24, 14, 0, 0, 0, time.UTC),
		}

		comment3 := &discuss.Comment{
			ID:        uuid.NewString(),
			PostID:    post2.ID,
			AuthorID:  user.ID,
			Content:   "comment on other post",
			CreatedAt: time.Date(2026, 2, 24, 15, 0, 0, 0, time.UTC),
		}

		err := commentRepo.Insert(ctx, comment1)
		require.NoError(t, err)

		err = commentRepo.Insert(ctx, comment2)
		require.NoError(t, err)

		err = commentRepo.Insert(ctx, comment3)
		require.NoError(t, err)

		comments, err := commentRepo.List(ctx, &discuss.ListCommentsParams{})
		require.NoError(t, err)
		assert.Len(t, comments, 3)
		assert.Equal(t, comment1.ID, comments[0].ID)
		assert.Equal(t, comment2.ID, comments[1].ID)
		assert.Equal(t, comment3.ID, comments[2].ID)

		post1Comments, err := commentRepo.List(ctx, &discuss.ListCommentsParams{PostID: post1.ID})
		require.NoError(t, err)
		assert.Len(t, post1Comments, 2)
		assert.Equal(t, comment1.ID, post1Comments[0].ID)
		assert.Equal(t, comment2.ID, post1Comments[1].ID)

		countAll, err := commentRepo.Count(ctx, &discuss.CountCommentsParams{})
		require.NoError(t, err)
		assert.Equal(t, 3, countAll)

		countPost1, err := commentRepo.Count(ctx, &discuss.CountCommentsParams{PostID: post1.ID})
		require.NoError(t, err)
		assert.Equal(t, 2, countPost1)
	})
}

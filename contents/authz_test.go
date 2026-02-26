package contents_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/google/uuid"
	authcontext "github.com/nasermirzaei89/scribble/authentication/context"
	"github.com/nasermirzaei89/scribble/authorization"
	"github.com/nasermirzaei89/scribble/authorization/casbin"
	"github.com/nasermirzaei89/scribble/contents"
	"github.com/stretchr/testify/require"
)

type stubService struct{}

func (s *stubService) CreatePost(ctx context.Context, req contents.CreatePostRequest) (*contents.Post, error) {
	return &contents.Post{ID: "post1", AuthorID: req.AuthorID, Content: req.Content}, nil
}

func (s *stubService) ListPosts(ctx context.Context) ([]*contents.Post, error) {
	return []*contents.Post{}, nil
}

func (s *stubService) GetPost(ctx context.Context, postID string) (*contents.Post, error) {
	return &contents.Post{ID: postID, AuthorID: "author1", Content: "test"}, nil
}

func TestAuthorizationMiddleware(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policy.csv")
	content := []byte(`g, system:anonymous, system:unauthenticated

p, system:authenticated, github.com/nasermirzaei89/scribble/contents, -, createPost
p, system:authenticated, github.com/nasermirzaei89/scribble/contents, -, listPosts
p, system:unauthenticated, github.com/nasermirzaei89/scribble/contents, -, listPosts
p, system:authenticated, github.com/nasermirzaei89/scribble/contents, *, getPost
p, system:unauthenticated, github.com/nasermirzaei89/scribble/contents, *, getPost
`)

	err := os.WriteFile(tmpFile, content, 0o600)
	require.NoError(t, err)

	adapter := fileadapter.NewAdapter(tmpFile)

	provider, err := casbin.NewAuthorizationProvider(adapter)
	require.NoError(t, err)

	authzSvc, err := authorization.NewService(provider)
	require.NoError(t, err)

	client := authorization.NewClient(authzSvc)
	svc := contents.NewAuthorizationMiddleware(client, &stubService{})

	userID := uuid.NewString()
	err = client.AddToGroup(ctx, userID, authcontext.Authenticated)
	require.NoError(t, err)

	authorID := uuid.NewString()

	anonymousCtx := ctx
	authenticatedCtx := authcontext.WithSubject(ctx, userID)

	t.Run("anonymous", func(t *testing.T) {
		_, err := svc.CreatePost(anonymousCtx, contents.CreatePostRequest{AuthorID: authorID, Content: "post"})
		require.Error(t, err)

		accessDeniedErr := &authorization.AccessDeniedError{}
		require.ErrorAs(t, err, &accessDeniedErr)

		_, err = svc.ListPosts(anonymousCtx)
		require.NoError(t, err)

		_, err = svc.GetPost(anonymousCtx, "post1")
		require.NoError(t, err)
	})

	t.Run("authenticated", func(t *testing.T) {
		_, err := svc.CreatePost(authenticatedCtx, contents.CreatePostRequest{AuthorID: authorID, Content: "post"})
		require.NoError(t, err)

		_, err = svc.ListPosts(authenticatedCtx)
		require.NoError(t, err)

		_, err = svc.GetPost(authenticatedCtx, "post1")
		require.NoError(t, err)
	})
}

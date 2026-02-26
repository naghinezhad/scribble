package reactions_test

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
	"github.com/nasermirzaei89/scribble/reactions"
	"github.com/stretchr/testify/require"
)

type stubService struct{}

func (s *stubService) AllowedEmojis(
	ctx context.Context,
	targetType reactions.TargetType,
	targetID string,
) ([]string, error) {
	return []string{"👍", "👎", "😂"}, nil
}

func (s *stubService) ToggleReaction(
	ctx context.Context,
	targetType reactions.TargetType,
	targetID string,
	userID string,
	emoji string,
) error {
	return nil
}

func (s *stubService) GetTargetReactions(
	ctx context.Context,
	targetType reactions.TargetType,
	targetID string,
	currentUserID *string,
) (*reactions.TargetReactions, error) {
	return &reactions.TargetReactions{
		TargetType: targetType,
		TargetID:   targetID,
		Options:    []reactions.ReactionOption{},
	}, nil
}

func TestAuthorizationMiddleware(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "policy.csv")
	content := []byte(`g, system:anonymous, system:unauthenticated

p, system:authenticated, github.com/nasermirzaei89/scribble/reactions, *, toggleReaction
p, system:authenticated, github.com/nasermirzaei89/scribble/reactions, *, getTargetReactions
p, system:unauthenticated, github.com/nasermirzaei89/scribble/reactions, *, getTargetReactions
`)

	err := os.WriteFile(tmpFile, content, 0o600)
	require.NoError(t, err)

	adapter := fileadapter.NewAdapter(tmpFile)
	provider, err := casbin.NewAuthorizationProvider(adapter)
	require.NoError(t, err)

	authzSvc, err := authorization.NewService(provider)
	require.NoError(t, err)

	client := authorization.NewClient(authzSvc)
	svc := reactions.NewAuthorizationMiddleware(client, &stubService{})

	userID := uuid.NewString()
	err = client.AddToGroup(ctx, userID, authcontext.Authenticated)
	require.NoError(t, err)

	targetType := reactions.TargetTypePost
	targetID := uuid.NewString()
	emoji := "👍"

	anonymousCtx := ctx
	authenticatedCtx := authcontext.WithSubject(ctx, userID)

	t.Run("anonymous", func(t *testing.T) {
		err := svc.ToggleReaction(anonymousCtx, targetType, targetID, userID, emoji)
		require.Error(t, err)

		accessDeniedErr := &authorization.AccessDeniedError{}
		require.ErrorAs(t, err, &accessDeniedErr)

		_, err = svc.GetTargetReactions(anonymousCtx, targetType, targetID, nil)
		require.NoError(t, err)
	})

	t.Run("authenticated", func(t *testing.T) {
		err := svc.ToggleReaction(authenticatedCtx, targetType, targetID, userID, emoji)
		require.NoError(t, err)

		_, err = svc.GetTargetReactions(authenticatedCtx, targetType, targetID, &userID)
		require.NoError(t, err)
	})
}
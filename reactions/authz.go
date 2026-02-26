package reactions

import (
	"context"
	"fmt"

	"github.com/nasermirzaei89/scribble/authorization"
)

const (
	ActionToggleReaction     = "toggleReaction"
	ActionGetTargetReactions = "getTargetReactions"
)

type AuthorizationMiddleware struct {
	authzClient *authorization.Client
	next        Service
}

var _ Service = (*AuthorizationMiddleware)(nil)

func NewAuthorizationMiddleware(authzClient *authorization.Client, next Service) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		authzClient: authzClient,
		next:        next,
	}
}

func (mw *AuthorizationMiddleware) AllowedEmojis(
	ctx context.Context,
	targetType TargetType,
	targetID string,
) ([]string, error) {
	emojis, err := mw.next.AllowedEmojis(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return emojis, nil
}

func (mw *AuthorizationMiddleware) ToggleReaction(
	ctx context.Context,
	targetType TargetType,
	targetID string,
	userID string,
	emoji string,
) error {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, string(targetType)+":"+targetID, ActionToggleReaction)
	if err != nil {
		return fmt.Errorf("failed to check authorization: %w", err)
	}

	err = mw.next.ToggleReaction(ctx, targetType, targetID, userID, emoji)
	if err != nil {
		return fmt.Errorf("failed to call next method: %w", err)
	}

	return nil
}

func (mw *AuthorizationMiddleware) GetTargetReactions(
	ctx context.Context,
	targetType TargetType,
	targetID string,
	currentUserID *string,
) (*TargetReactions, error) {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, string(targetType)+":"+targetID, ActionGetTargetReactions)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	res, err := mw.next.GetTargetReactions(ctx, targetType, targetID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return res, nil
}

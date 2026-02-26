package reactions

import (
	"context"
	"fmt"

	"github.com/nasermirzaei89/scribble/authorization"
)

const (
	ActionToggleReaction = "toggleReaction"
	ActionGetMyReactions = "getMyReactions"
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

func (mw *AuthorizationMiddleware) ToggleMyReaction(
	ctx context.Context,
	targetType TargetType,
	targetID string,
	emoji string,
) error {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, "", ActionToggleReaction)
	if err != nil {
		return fmt.Errorf("failed to check authorization: %w", err)
	}

	err = mw.next.ToggleMyReaction(ctx, targetType, targetID, emoji)
	if err != nil {
		return fmt.Errorf("failed to call next method: %w", err)
	}

	return nil
}

func (mw *AuthorizationMiddleware) GetMyReactions(
	ctx context.Context,
	targetType TargetType,
	targetID string,
) (*TargetReactions, error) {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, "", ActionGetMyReactions)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	res, err := mw.next.GetMyReactions(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return res, nil
}

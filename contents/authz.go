package contents

import (
	"context"
	"fmt"

	"github.com/nasermirzaei89/scribble/authorization"
)

const (
	ActionCreatePost = "createPost"
	ActionListPosts  = "listPosts"
	ActionGetPost    = "getPost"
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

func (mw *AuthorizationMiddleware) CreatePost(ctx context.Context, req CreatePostRequest) (*Post, error) {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, "", ActionCreatePost)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	post, err := mw.next.CreatePost(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return post, nil
}

func (mw *AuthorizationMiddleware) ListPosts(ctx context.Context) ([]*Post, error) {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, "", ActionListPosts)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	posts, err := mw.next.ListPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return posts, nil
}

func (mw *AuthorizationMiddleware) GetPost(ctx context.Context, postID string) (*Post, error) {
	err := mw.authzClient.CheckAccess(ctx, ServiceName, postID, ActionGetPost)
	if err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	post, err := mw.next.GetPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to call next method: %w", err)
	}

	return post, nil
}

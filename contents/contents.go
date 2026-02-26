package contents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nasermirzaei89/scribble/authorization"
)

const ServiceName = "github.com/nasermirzaei89/scribble/contents"

type Service interface {
	CreatePost(ctx context.Context, req CreatePostRequest) (*Post, error)
	ListPosts(ctx context.Context) ([]*Post, error)
	GetPost(ctx context.Context, postID string) (*Post, error)
}

type BaseService struct {
	postRepo PostRepository
}

var _ Service = (*BaseService)(nil)

func NewService(postRepo PostRepository, authzClient *authorization.Client) Service { //nolint:ireturn
	return NewAuthorizationMiddleware(authzClient, NewBaseService(postRepo))
}

func NewBaseService(postRepo PostRepository) *BaseService {
	return &BaseService{
		postRepo: postRepo,
	}
}

type CreatePostRequest struct {
	AuthorID string
	Content  string
}

func (svc *BaseService) CreatePost(ctx context.Context, req CreatePostRequest) (*Post, error) {
	post := &Post{
		ID:        uuid.NewString(),
		AuthorID:  req.AuthorID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err := svc.postRepo.Insert(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (svc *BaseService) ListPosts(ctx context.Context) ([]*Post, error) {
	posts, err := svc.postRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, nil
}

func (svc *BaseService) GetPost(ctx context.Context, postID string) (*Post, error) {
	post, err := svc.postRepo.Find(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find post: %w", err)
	}

	return post, nil
}

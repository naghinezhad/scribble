package reactions

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	authcontext "github.com/nasermirzaei89/scribble/authentication/context"
	"github.com/nasermirzaei89/scribble/authorization"
)

const ServiceName = "github.com/nasermirzaei89/scribble/reactions"

type Service interface {
	AllowedEmojis(ctx context.Context, targetType TargetType, targetID string) ([]string, error)
	ToggleMyReaction(ctx context.Context, targetType TargetType, targetID string, emoji string) error
	GetMyReactions(
		ctx context.Context,
		targetType TargetType,
		targetID string,
	) (*TargetReactions, error)
}

type BaseService struct {
	userReactionRepo UserReactionRepository
}

var _ Service = (*BaseService)(nil)

func NewService(userReactionRepo UserReactionRepository, authzClient *authorization.Client) Service { //nolint:ireturn
	return NewAuthorizationMiddleware(authzClient, NewBaseService(userReactionRepo))
}

func NewBaseService(userReactionRepo UserReactionRepository) *BaseService {
	return &BaseService{userReactionRepo: userReactionRepo}
}

type ReactionOption struct {
	Emoji     string
	Count     int
	Selected  bool
	Available bool
}

type TargetReactions struct {
	TargetType TargetType
	TargetID   string
	Options    []ReactionOption
}

func (svc *BaseService) AllowedEmojis(
	_ context.Context,
	targetType TargetType,
	_ string,
) ([]string, error) {
	if !targetType.IsValid() {
		return nil, InvalidTargetTypeError{TargetType: targetType}
	}

	return []string{"👍", "👎", "😂"}, nil
}

func (svc *BaseService) ToggleMyReaction(
	ctx context.Context,
	targetType TargetType,
	targetID string,
	emoji string,
) error {
	userID := authcontext.GetSubject(ctx)

	if !targetType.IsValid() {
		return InvalidTargetTypeError{TargetType: targetType}
	}

	allowedEmojis, err := svc.AllowedEmojis(ctx, targetType, targetID)
	if err != nil {
		return fmt.Errorf("failed to get allowed emojis: %w", err)
	}

	if !slices.Contains(allowedEmojis, emoji) {
		return InvalidEmojiError{
			TargetType: targetType,
			TargetID:   targetID,
			Emoji:      emoji,
			Allowed:    allowedEmojis,
		}
	}

	existingReaction, err := svc.userReactionRepo.FindByUserTarget(ctx, targetType, targetID, userID)
	if err != nil {
		if _, ok := errors.AsType[*UserReactionNotFoundError](err); !ok {
			return fmt.Errorf("failed to get existing reaction: %w", err)
		}
	}

	if existingReaction != nil && existingReaction.Emoji == emoji {
		err = svc.userReactionRepo.DeleteByUserTarget(ctx, targetType, targetID, userID)
		if err != nil {
			return fmt.Errorf("failed to remove reaction: %w", err)
		}

		return nil
	}

	userReaction := &UserReaction{
		TargetType: targetType,
		TargetID:   targetID,
		UserID:     userID,
		Emoji:      emoji,
		CreatedAt:  time.Now(),
	}

	err = svc.userReactionRepo.Upsert(ctx, userReaction)
	if err != nil {
		return fmt.Errorf("failed to set reaction: %w", err)
	}

	return nil
}

func (svc *BaseService) GetMyReactions(
	ctx context.Context,
	targetType TargetType,
	targetID string,
) (*TargetReactions, error) {
	allowedEmojis, err := svc.AllowedEmojis(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed emojis: %w", err)
	}

	counts, err := svc.userReactionRepo.CountByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get counts by target: %w", err)
	}

	selectedEmoji := ""
	currentUserID := authcontext.GetSubject(ctx)

	if currentUserID != "" && currentUserID != authcontext.Anonymous {
		userReaction, err := svc.userReactionRepo.FindByUserTarget(ctx, targetType, targetID, currentUserID)
		if err != nil {
			if _, ok := errors.AsType[*UserReactionNotFoundError](err); !ok {
				return nil, fmt.Errorf("failed to get user reaction: %w", err)
			}
		}

		if userReaction != nil {
			selectedEmoji = userReaction.Emoji
		}
	}

	options := make([]ReactionOption, 0, len(allowedEmojis))
	availableEmojiSet := make(map[string]struct{}, len(allowedEmojis))

	for _, emoji := range allowedEmojis {
		availableEmojiSet[emoji] = struct{}{}

		options = append(options, ReactionOption{
			Emoji:     emoji,
			Count:     counts[emoji],
			Selected:  emoji == selectedEmoji,
			Available: true,
		})
	}

	additionalEmojis := make([]string, 0)

	for emoji, count := range counts {
		if count <= 0 {
			continue
		}

		if _, ok := availableEmojiSet[emoji]; ok {
			continue
		}

		additionalEmojis = append(additionalEmojis, emoji)
	}

	slices.Sort(additionalEmojis)

	for _, emoji := range additionalEmojis {
		options = append(options, ReactionOption{
			Emoji:     emoji,
			Count:     counts[emoji],
			Selected:  emoji == selectedEmoji,
			Available: false,
		})
	}

	return &TargetReactions{
		TargetType: targetType,
		TargetID:   targetID,
		Options:    options,
	}, nil
}

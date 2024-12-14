package match

import (
	"context"
	"fmt"
)

// ============================
// Service Interfaces
// ============================
type UserService interface {
	GetUserById(ctx context.Context, uid uint64) (*User, error)
	GetUserIdBySession(ctx context.Context, sid string) (uint64, error)
	AddUserToChannel(ctx context.Context, channelID, userID uint64) error
}

type MatchService interface {
	Match(ctx context.Context, userId uint64) (*MatchResult, error)
	BroadcastMatchResult(ctx context.Context, result *MatchResult) error
	RemoveUserFromWaitList(ctx context.Context, userId uint64) error
}

// ============================
// Service Implementations
// ============================
type UserServiceImpl struct {
	userRepo UserRepo
}

func NewUserServiceImpl(userRepo UserRepo) *UserServiceImpl {
	return &UserServiceImpl{userRepo}
}

type MatchServiceImpl struct {
	matchRepo MatchRepo
	chanRepo  ChannelRepo
}

func NewMatchServiceImpl(matchRepo MatchRepo, chanRepo ChannelRepo) *MatchServiceImpl {
	return &MatchServiceImpl{matchRepo, chanRepo}
}

// ============================
// Service Functions
// ============================
func (s *UserServiceImpl) GetUserById(ctx context.Context, uid uint64) (*User, error) {
	user, err := s.userRepo.GetUserById(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error get user %d: %w", uid, err)
	}
	return user, nil
}

func (s *UserServiceImpl) GetUserIdBySession(ctx context.Context, sid string) (uint64, error) {
	userID, err := s.userRepo.GetUserIdBySession(ctx, sid)
	if err != nil {
		return 0, fmt.Errorf("error get user id by sid %s: %w", sid, err)
	}
	return userID, nil
}

func (s *UserServiceImpl) AddUserToChannel(ctx context.Context, channelID, userID uint64) error {
	if err := s.userRepo.AddUserToChannel(ctx, channelID, userID); err != nil {
		return fmt.Errorf("error add user %d to channel %d: %w", userID, channelID, err)
	}
	return nil
}

func (s *MatchServiceImpl) Match(ctx context.Context, userId uint64) (*MatchResult, error) {
	matched, peerId, err := s.matchRepo.PopOrPushWaitList(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error match user %d: %w", userId, err)
	}

	if matched {
		newChannelId, accessToken, err := s.chanRepo.CreateChannel(ctx)
		if err != nil {
			return nil, fmt.Errorf("error create channel for user %d: %w", userId, err)
		}

		return &MatchResult{
			Matched:     true,
			UserId:      userId,
			PeerId:      peerId,
			ChannelId:   newChannelId,
			AccessToken: accessToken,
		}, nil
	}

	return &MatchResult{
		Matched: false,
	}, nil
}

func (s *MatchServiceImpl) BroadcastMatchResult(ctx context.Context, result *MatchResult) error {
	if err := s.matchRepo.PublishMatchResult(ctx, result); err != nil {
		return fmt.Errorf("error publish match result: %w", err)
	}
	return nil
}

func (s *MatchServiceImpl) RemoveUserFromWaitList(ctx context.Context, userId uint64) error {
	if err := s.matchRepo.RemoveFromWaitList(ctx, userId); err != nil {
		return fmt.Errorf("error remove user %d from wait list: %w", userId, err)
	}
	return nil
}

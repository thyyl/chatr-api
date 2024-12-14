package forwarder

import (
	"context"

	"github.com/thyyl/chatr/pkg/chat"
)

type ForwarderService interface {
	RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error
	RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error
	ForwardMessage(ctx context.Context, chatMessage *chat.Message) error
}

type ForwarderServiceImpl struct {
	forwarderRepo ForwarderRepo
}

func NewForwarderServiceImpl(forwardRepo ForwarderRepo) *ForwarderServiceImpl {
	return &ForwarderServiceImpl{
		forwardRepo,
	}
}

func (s *ForwarderServiceImpl) RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error {
	return s.forwarderRepo.RegisterChannelSession(ctx, channelId, userId, subscriber)
}

func (s *ForwarderServiceImpl) RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error {
	return s.forwarderRepo.RemoveChannelSession(ctx, channelId, userId)
}

func (s *ForwarderServiceImpl) ForwardMessage(ctx context.Context, chatMessage *chat.Message) error {
	subscribers, err := s.forwarderRepo.GetSubscribers(ctx, chatMessage.ChannelId)
	if err != nil {
		return err
	}
	return s.forwarderRepo.ForwardMessage(ctx, chatMessage, subscribers)
}

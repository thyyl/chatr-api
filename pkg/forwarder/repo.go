package forwarder

import (
	"context"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/thyyl/chatr/pkg/chat"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/infra"
)

type Subscribers map[string]struct{}

type ForwarderRepo interface {
	RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error
	RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error
	GetSubscribers(ctx context.Context, channelId uint64) (Subscribers, error)
	ForwardMessage(ctx context.Context, chatMessage *chat.Message, subscribers Subscribers) error
}

type ForwarderRepoImpl struct {
	redis     infra.RedisCache
	publisher message.Publisher
}

func NewForwarderRepoImpl(redis infra.RedisCache, publisher message.Publisher) *ForwarderRepoImpl {
	return &ForwarderRepoImpl{
		redis:     redis,
		publisher: publisher,
	}
}

func (repo *ForwarderRepoImpl) RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error {
	key := constructKey(channelId)
	return repo.redis.HSet(ctx, key, strconv.FormatUint(userId, 10), subscriber)
}

func (repo *ForwarderRepoImpl) RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error {
	key := constructKey(channelId)
	return repo.redis.HDel(ctx, key, strconv.FormatUint(userId, 10))
}

func (repo *ForwarderRepoImpl) GetSubscribers(ctx context.Context, channelId uint64) (Subscribers, error) {
	key := constructKey(channelId)
	result, err := repo.redis.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	subscribers := make(Subscribers)
	for _, subscriber := range result {
		subscribers[subscriber] = struct{}{}
	}
	return subscribers, nil
}

func (repo *ForwarderRepoImpl) ForwardMessage(ctx context.Context, chatMessage *chat.Message, subscribers Subscribers) error {
	var err error

	for subscriber := range subscribers {
		err = repo.publisher.Publish(subscriber, message.NewMessage(
			watermill.NewUUID(),
			chatMessage.Encode(),
		))

		if err != nil {
			return err
		}
	}

	return nil
}

func constructKey(id uint64) string {
	return common.Join(common.ForwardRcKey, ":", strconv.FormatUint(id, 10))
}

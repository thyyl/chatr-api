package chat

import (
	"context"
	"strconv"

	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/infra"
)

// ============================
// Repository Interfaces
// ============================
type UserRepoCache interface {
	AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error
	GetUserById(ctx context.Context, userId uint64) (*User, error)
	IsChannelUserExists(ctx context.Context, channelId uint64, userId uint64) (bool, error)
	GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error)
	AddOnlineUser(ctx context.Context, channelId uint64, userId uint64) error
	DeleteOnlineUser(ctx context.Context, channelId uint64, userId uint64) error
	GetOnlineUserIds(ctx context.Context, channelId uint64) ([]uint64, error)
}

type ChatRepoCache interface {
	InsertMessage(ctx context.Context, chatMessage *Message) error
	MarkMessageSeen(ctx context.Context, channelId uint64, messageId uint64) error
	PublishMessage(ctx context.Context, chatMessage *Message) error
	ListMessages(ctx context.Context, channelId uint64, pageState string) ([]*Message, string, error)
}

type ChannelRepoCache interface {
	CreateChannel(ctx context.Context, channelId uint64) (*Channel, error)
	DeleteChannel(ctx context.Context, channelId uint64) error
}

// ============================
// Repository Implementations
// ============================
type UserRepoCacheImpl struct {
	redis    infra.RedisCache
	userRepo UserRepo
}

func NewUserRepoCacheImpl(redis infra.RedisCache, userRepo UserRepo) *UserRepoCacheImpl {
	return &UserRepoCacheImpl{
		redis:    redis,
		userRepo: userRepo,
	}
}

type ChatRepoCacheImpl struct {
	chatRepo ChatRepo
}

func NewChatRepoCacheImpl(chatRepo ChatRepo) *ChatRepoCacheImpl {
	return &ChatRepoCacheImpl{
		chatRepo: chatRepo,
	}
}

type ChannelRepoCacheImpl struct {
	redis       infra.RedisCache
	channelRepo ChannelRepo
}

func NewChannelRepoCacheImpl(redis infra.RedisCache, channelRepo ChannelRepo) *ChannelRepoCacheImpl {
	return &ChannelRepoCacheImpl{
		redis:       redis,
		channelRepo: channelRepo,
	}
}

// ============================
// Repository Functions
// ============================
func (cache *UserRepoCacheImpl) AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error {
	if err := cache.userRepo.AddUserToChannel(ctx, channelId, userId); err != nil {
		return err
	}

	key := constructKey(common.ChannelUsersRcKey, channelId)
	return cache.redis.HSet(ctx, key, strconv.FormatUint(userId, 10), 1)
}

func (cache *UserRepoCacheImpl) GetUserById(ctx context.Context, userId uint64) (*User, error) {
	return cache.userRepo.GetUserById(ctx, userId)
}

func (cache *UserRepoCacheImpl) IsChannelUserExists(ctx context.Context, channelId uint64, userId uint64) (bool, error) {
	key := constructKey(common.ChannelUsersRcKey, channelId)
	var dummy int
	var err error

	channelExists, userExists, err := cache.redis.HGetIfKeyExists(ctx, key, strconv.FormatUint(userId, 10), &dummy)
	if err != nil {
		return false, err
	}

	if channelExists {
		if !userExists {
			return false, nil
		}

		return true, nil
	}

	channelUserIds, err := cache.userRepo.GetChannelUserIds(ctx, channelId)
	if err != nil {
		return false, err
	}

	channelUserExist := false
	var args []interface{}

	for _, channelUserId := range channelUserIds {
		if userId == channelUserId {
			channelUserExist = true
		}

		args = append(args, channelUserId, 1)
	}

	if err := cache.redis.HSet(ctx, key, args...); err != nil {
		return false, err
	}

	return channelUserExist, nil
}

func (cache *UserRepoCacheImpl) GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error) {
	key := constructKey(common.ChannelUsersRcKey, channelId)
	userMap, err := cache.redis.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	var userIds []uint64

	if len(userMap) > 0 {
		for userIdStr := range userMap {
			userId, err := strconv.ParseUint(userIdStr, 10, 64)
			if err != nil {
				return nil, err
			}
			userIds = append(userIds, userId)
		}
		return userIds, nil
	}

	userIds, err = cache.userRepo.GetChannelUserIds(ctx, channelId)
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, userId := range userIds {
		args = append(args, userId, 1)
	}

	if err := cache.redis.HSet(ctx, key, args...); err != nil {
		return userIds, err
	}

	return userIds, nil
}

func (cache *UserRepoCacheImpl) AddOnlineUser(ctx context.Context, channelId uint64, userId uint64) error {
	key := constructKey(common.OnlineUsersRcKey, channelId)
	userKey := strconv.FormatUint(userId, 10)
	return cache.redis.HSet(ctx, key, userKey, 1)
}

func (cache *UserRepoCacheImpl) DeleteOnlineUser(ctx context.Context, channelId uint64, userId uint64) error {
	key := constructKey(common.OnlineUsersRcKey, channelId)
	userKey := strconv.FormatUint(userId, 10)
	return cache.redis.HDel(ctx, key, userKey)
}

func (cache *UserRepoCacheImpl) GetOnlineUserIds(ctx context.Context, channelId uint64) ([]uint64, error) {
	key := constructKey(common.OnlineUsersRcKey, channelId)

	userMap, err := cache.redis.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}

	var userIds []uint64
	for userIdStr := range userMap {
		userId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return userIds, nil
}

func (cache *ChatRepoCacheImpl) InsertMessage(ctx context.Context, chatMessage *Message) error {
	return cache.chatRepo.InsertMessage(ctx, chatMessage)
}

func (cache *ChatRepoCacheImpl) MarkMessageSeen(ctx context.Context, channelId uint64, messageId uint64) error {
	return cache.chatRepo.MarkMessageSeen(ctx, channelId, messageId)
}

func (cache *ChatRepoCacheImpl) PublishMessage(ctx context.Context, chatMessage *Message) error {
	return cache.chatRepo.PublishMessage(ctx, chatMessage)
}

func (cache *ChatRepoCacheImpl) ListMessages(ctx context.Context, channelId uint64, pageState string) ([]*Message, string, error) {
	return cache.chatRepo.ListMessages(ctx, channelId, pageState)
}

func (cache *ChannelRepoCacheImpl) CreateChannel(ctx context.Context, channelId uint64) (*Channel, error) {
	return cache.channelRepo.CreateChannel(ctx, channelId)
}

func (cache *ChannelRepoCacheImpl) DeleteChannel(ctx context.Context, channelId uint64) error {
	if err := cache.channelRepo.DeleteChannel(ctx, channelId); err != nil {
		return err
	}

	channelUsersKey := constructKey(common.ChannelUsersRcKey, channelId)
	onlineUsersKey := constructKey(common.OnlineUsersRcKey, channelId)

	cmds := []infra.RedisCmd{
		{
			OpType: infra.DELETE,
			Payload: infra.RedisDeletePayload{
				Key: onlineUsersKey,
			},
		},
		{
			OpType: infra.DELETE,
			Payload: infra.RedisDeletePayload{
				Key: channelUsersKey,
			},
		},
	}

	return cache.redis.ExecPipeLine(ctx, &cmds)
}

func constructKey(prefix string, id uint64) string {
	return common.Join(prefix, ":", strconv.FormatUint(id, 10))
}

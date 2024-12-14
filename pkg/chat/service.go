package chat

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/thyyl/chatr/pkg/common"
)

// ============================
// Service Interfaces
// ============================
type UserService interface {
	AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error
	GetUser(ctx context.Context, userId uint64) (*User, error)
	IsChannelUserExists(ctx context.Context, channelId uint64, userId uint64) (bool, error)
	GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error)
	AddOnlineUser(ctx context.Context, channelId uint64, userId uint64) error
	DeleteOnlineUser(ctx context.Context, channelId uint64, userId uint64) error
	GetOnlineUserIds(ctx context.Context, channelId uint64) ([]uint64, error)
}

type ChatService interface {
	BroadcastTextMessage(ctx context.Context, channelId uint64, userId uint64, payload string) error
	BroadcastConnectMessage(ctx context.Context, channelId uint64, userId uint64) error
	BroadcastActionMessage(ctx context.Context, channelId uint64, userId uint64, action Action) error
	BroadcastFileMessage(ctx context.Context, channelId uint64, userId uint64, payload string) error
	MarkMessageSeen(ctx context.Context, channelId uint64, userId uint64, messageId uint64) error
	InsertMessage(ctx context.Context, chatMessage *Message) error
	PublishMessage(ctx context.Context, chatMessage *Message) error
	ListMessages(ctx context.Context, channelId uint64, pageState string) ([]*Message, string, error)
}

type ChannelService interface {
	CreateChannel(ctx context.Context) (*Channel, error)
	DeleteChannel(ctx context.Context, channelId uint64) error
}

type ForwarderService interface {
	RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error
	RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error
}

// ============================
// Service Implementations
// ============================
type UserServiceImpl struct {
	userRepoCache UserRepoCache
}

func NewUserServiceImpl(userRepoCache UserRepoCache) *UserServiceImpl {
	return &UserServiceImpl{userRepoCache}
}

type ChatServiceImpl struct {
	chatRepoCache ChatRepoCache
	userRepoCache UserRepoCache
	sf            common.IDGenerator
}

func NewChatServiceImpl(chatRepoCache ChatRepoCache, userRepoCache UserRepoCache, sf common.IDGenerator) *ChatServiceImpl {
	return &ChatServiceImpl{chatRepoCache, userRepoCache, sf}
}

type ChannelServiceImpl struct {
	channelRepoCache ChannelRepoCache
	userRepoCache    UserRepoCache
	sf               common.IDGenerator
}

func NewChannelServiceImpl(channelRepoCache ChannelRepoCache, userRepoCache UserRepoCache, sf common.IDGenerator) *ChannelServiceImpl {
	return &ChannelServiceImpl{channelRepoCache, userRepoCache, sf}
}

type ForwarderServiceImpl struct {
	forwarderRepo ForwarderRepo
}

func NewForwarderServiceImpl(forwarderRepo ForwarderRepo) *ForwarderServiceImpl {
	return &ForwarderServiceImpl{forwarderRepo}
}

// ============================
// Service Functions
// ============================
func (s *UserServiceImpl) AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error {
	if err := s.userRepoCache.AddUserToChannel(ctx, channelId, userId); err != nil {
		return fmt.Errorf("error add user %d to channel %d: %w", userId, channelId, err)
	}
	return nil
}

func (s *UserServiceImpl) GetUser(ctx context.Context, userId uint64) (*User, error) {
	user, err := s.userRepoCache.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error get user %d: %w", userId, err)
	}
	return user, nil
}

func (s *UserServiceImpl) IsChannelUserExists(ctx context.Context, channelId uint64, userId uint64) (bool, error) {
	exist, err := s.userRepoCache.IsChannelUserExists(ctx, channelId, userId)
	if err != nil {
		return false, fmt.Errorf("error check user %d in channel %d: %w", userId, channelId, err)
	}

	return exist, nil
}

func (s *UserServiceImpl) GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error) {
	userIds, err := s.userRepoCache.GetChannelUserIds(ctx, channelId)
	if err != nil {
		return nil, fmt.Errorf("error get user ids in channel %d: %w", channelId, err)
	}

	return userIds, nil
}

func (s *UserServiceImpl) AddOnlineUser(ctx context.Context, channelId uint64, userId uint64) error {
	if err := s.userRepoCache.AddOnlineUser(ctx, channelId, userId); err != nil {
		return fmt.Errorf("error add online user %d in channel %d: %w", userId, channelId, err)
	}

	return nil
}

func (s *UserServiceImpl) DeleteOnlineUser(ctx context.Context, channelId uint64, userId uint64) error {
	if err := s.userRepoCache.DeleteOnlineUser(ctx, channelId, userId); err != nil {
		return fmt.Errorf("error delete online user %d in channel %d: %w", userId, channelId, err)
	}

	return nil
}

func (s *UserServiceImpl) GetOnlineUserIds(ctx context.Context, channelId uint64) ([]uint64, error) {
	userIds, err := s.userRepoCache.GetOnlineUserIds(ctx, channelId)
	if err != nil {
		return nil, fmt.Errorf("error get online user ids in channel %d: %w", channelId, err)
	}

	return userIds, nil
}

func (s *ChatServiceImpl) BroadcastTextMessage(ctx context.Context, channelId uint64, userId uint64, payload string) error {
	messageId, err := s.sf.NextID()
	if err != nil {
		return fmt.Errorf("error create snowflake ID for text message: %w", err)
	}

	chatMessage := &Message{
		MessageId: messageId,
		Event:     EventText,
		ChannelId: channelId,
		UserId:    userId,
		Payload:   payload,
		Time:      time.Now().UnixMilli(),
	}
	if err := s.chatRepoCache.InsertMessage(ctx, chatMessage); err != nil {
		return err
	}
	if err := s.PublishMessage(ctx, chatMessage); err != nil {
		return err
	}

	return nil
}

func (s *ChatServiceImpl) BroadcastConnectMessage(ctx context.Context, channelId uint64, userId uint64) error {
	onlineUserIds, err := s.userRepoCache.GetOnlineUserIds(context.Background(), channelId)
	if err != nil {
		return fmt.Errorf("error get online user ids from channel %d: %w", channelId, err)
	}

	if len(onlineUserIds) == 1 {
		return s.BroadcastActionMessage(ctx, channelId, userId, WaitingMessage)
	}

	return s.BroadcastActionMessage(ctx, channelId, userId, JoinedMessage)
}

func (s *ChatServiceImpl) BroadcastActionMessage(ctx context.Context, channelId uint64, userId uint64, action Action) error {
	eventMessageId, err := s.sf.NextID()
	if err != nil {
		return fmt.Errorf("error create snowflake ID for action message: %w", err)
	}

	chatMessage := &Message{
		MessageId: eventMessageId,
		Event:     EventAction,
		ChannelId: channelId,
		UserId:    userId,
		Payload:   string(action),
		Time:      time.Now().UnixMilli(),
	}

	if err := s.PublishMessage(ctx, chatMessage); err != nil {
		return err
	}

	return nil
}

func (s *ChatServiceImpl) BroadcastFileMessage(ctx context.Context, channelId uint64, userId uint64, payload string) error {
	messageId, err := s.sf.NextID()
	if err != nil {
		return fmt.Errorf("error create snowflake ID for file message: %w", err)
	}

	chatMessage := &Message{
		MessageId: messageId,
		Event:     EventFile,
		ChannelId: channelId,
		UserId:    userId,
		Payload:   payload,
		Time:      time.Now().UnixMilli(),
	}
	if err := s.chatRepoCache.InsertMessage(ctx, chatMessage); err != nil {
		return fmt.Errorf("error broadcast file message: %w", err)
	}
	if err := s.PublishMessage(ctx, chatMessage); err != nil {
		return fmt.Errorf("error broadcast file message: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) MarkMessageSeen(ctx context.Context, channelId uint64, userId uint64, messageId uint64) error {
	if err := s.chatRepoCache.MarkMessageSeen(ctx, channelId, messageId); err != nil {
		return fmt.Errorf("error mark message %d seen in channel %d: %w", messageId, channelId, err)
	}

	eventMessageId, err := s.sf.NextID()
	if err != nil {
		return fmt.Errorf("error create snowflake ID for seen message: %w", err)
	}

	chatMessage := &Message{
		MessageId: eventMessageId,
		Event:     EventSeen,
		ChannelId: channelId,
		UserId:    userId,
		Payload:   strconv.FormatUint(messageId, 10),
		Time:      time.Now().UnixMilli(),
		Seen:      true,
	}

	if err := s.PublishMessage(ctx, chatMessage); err != nil {
		return fmt.Errorf("error mark message %d seen in channel %d: %w", messageId, channelId, err)
	}

	return nil
}

func (s *ChatServiceImpl) InsertMessage(ctx context.Context, msg *Message) error {
	if err := s.chatRepoCache.InsertMessage(ctx, msg); err != nil {
		return fmt.Errorf("error insert message: %w", err)
	}
	return nil
}

func (s *ChatServiceImpl) PublishMessage(ctx context.Context, msg *Message) error {
	if err := s.chatRepoCache.PublishMessage(ctx, msg); err != nil {
		return fmt.Errorf("error publish message: %w", err)
	}
	return nil
}

func (s *ChatServiceImpl) ListMessages(ctx context.Context, channelId uint64, pageState string) ([]*Message, string, error) {
	messages, nextPageState, err := s.chatRepoCache.ListMessages(ctx, channelId, pageState)
	if err != nil {
		return nil, "", fmt.Errorf("error list messages in channel %d with page state %s: %w", channelId, pageState, err)
	}

	return messages, nextPageState, nil
}

func (s *ChannelServiceImpl) CreateChannel(ctx context.Context) (*Channel, error) {
	channelId, err := s.sf.NextID()
	if err != nil {
		return nil, fmt.Errorf("error create snowflake ID for channel: %w", err)
	}

	channel, err := s.channelRepoCache.CreateChannel(ctx, channelId)
	if err != nil {
		return nil, fmt.Errorf("error create channel: %w", err)
	}

	return channel, nil
}

func (s *ChannelServiceImpl) DeleteChannel(ctx context.Context, channelId uint64) error {
	if err := s.channelRepoCache.DeleteChannel(ctx, channelId); err != nil {
		return fmt.Errorf("error delete channel %d: %w", channelId, err)
	}

	return nil
}

func (s *ForwarderServiceImpl) RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error {
	return s.forwarderRepo.RegisterChannelSession(ctx, channelId, userId, subscriber)
}

func (s *ForwarderServiceImpl) RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error {
	return s.forwarderRepo.RemoveChannelSession(ctx, channelId, userId)
}

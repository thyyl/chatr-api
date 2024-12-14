package chat

import (
	"context"
	base64 "encoding/base64"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kit/kit/endpoint"
	"github.com/gocql/gocql"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"github.com/thyyl/chatr/pkg/transport"
	forwarderProto "github.com/thyyl/chatr/proto/forwarder"
	userProto "github.com/thyyl/chatr/proto/user"
)

// ============================
// Repository Interfaces
// ============================
type UserRepo interface {
	AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error
	GetUserById(ctx context.Context, userId uint64) (*User, error)
	GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error)
}

type ChannelRepo interface {
	CreateChannel(ctx context.Context, channelId uint64) (*Channel, error)
	DeleteChannel(ctx context.Context, channelId uint64) error
}

type ChatRepo interface {
	InsertMessage(ctx context.Context, chatMessage *Message) error
	MarkMessageSeen(ctx context.Context, channelId uint64, messageId uint64) error
	PublishMessage(ctx context.Context, chatMessage *Message) error
	ListMessages(ctx context.Context, channelId uint64, pageStateBase64 string) ([]*Message, string, error)
}

type ForwarderRepo interface {
	RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error
	RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error
}

// ============================
// Repository Implementations
// ============================
type UserRepoImpl struct {
	session *gocql.Session
	getUser endpoint.Endpoint
}

func NewUserRepoImpl(session *gocql.Session, userConn *UserClientConn) *UserRepoImpl {
	return &UserRepoImpl{
		session: session,
		getUser: transport.NewGrpcEndpoint(
			userConn.Conn,
			"user",
			"user.UserService",
			"GetUser",
			&userProto.GetUserResponse{},
		),
	}
}

type ChannelRepoImpl struct {
	session *gocql.Session
}

func NewChannelRepoImpl(session *gocql.Session) *ChannelRepoImpl {
	return &ChannelRepoImpl{
		session,
	}
}

type ChatRepoImpl struct {
	session     *gocql.Session
	publisher   message.Publisher
	maxMessages int64
	pagination  int
}

func NewChatRepoImpl(session *gocql.Session, publisher message.Publisher, config *config.Config) *ChatRepoImpl {
	return &ChatRepoImpl{
		session:     session,
		publisher:   publisher,
		maxMessages: config.Chat.Message.MaxNum,
		pagination:  config.Chat.Message.PaginationNum,
	}
}

type ForwarderRepoImpl struct {
	registerChannelSession endpoint.Endpoint
	removeChannelSession   endpoint.Endpoint
}

func NewForwarderRepoImpl(forwarderConn *ForwarderClientConn) *ForwarderRepoImpl {
	return &ForwarderRepoImpl{
		registerChannelSession: transport.NewGrpcEndpoint(
			forwarderConn.Conn,
			"forwarder",
			"forwarder.ForwarderService",
			"RegisterChannelSession",
			&forwarderProto.RegisterChannelSessionResponse{},
		),
		removeChannelSession: transport.NewGrpcEndpoint(
			forwarderConn.Conn,
			"forwarder",
			"forwarder.ForwarderService",
			"RemoveChannelSession",
			&forwarderProto.RemoveChannelSessionResponse{},
		),
	}
}

// ============================
// Repository Functions
// ============================
func (repo *UserRepoImpl) AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error {
	if err := repo.session.Query("INSERT INTO channels (id, user_id) VALUES (?, ?)",
		channelId, userId).WithContext(ctx).Exec(); err != nil {
		return err
	}

	return nil
}

func (repo *UserRepoImpl) GetUserById(ctx context.Context, userId uint64) (*User, error) {
	request := &userProto.GetUserRequest{Id: userId}
	response, err := repo.getUser(ctx, request)
	if err != nil {
		return nil, err
	}

	pbUser := response.(*userProto.GetUserResponse)
	if !pbUser.Exist {
		return nil, common.ErrorUserNotFound
	}

	return &User{
		Id:   pbUser.User.Id,
		Name: pbUser.User.Name,
	}, nil
}

func (repo *UserRepoImpl) GetChannelUserIds(ctx context.Context, channelId uint64) ([]uint64, error) {
	iteration := repo.session.Query("SELECT user_id FROM channels WHERE id = ?", channelId).WithContext(ctx).Idempotent(true).Iter()

	var userIds []uint64
	var userId uint64

	for iteration.Scan(&userId) {
		userIds = append(userIds, userId)
	}

	if err := iteration.Close(); err != nil {
		return nil, err
	}

	return userIds, nil
}

func (repo *ChannelRepoImpl) CreateChannel(ctx context.Context, channelId uint64) (*Channel, error) {
	if err := repo.session.Query("INSERT INTO channels (id, user_id) VALUES (?, ?)",
		channelId, 0).WithContext(ctx).Exec(); err != nil {
		return nil, err
	}

	accessToken, err := common.NewJWT(channelId)
	if err != nil {
		return nil, fmt.Errorf("error create JWT: %w", err)
	}

	return &Channel{
		Id:          channelId,
		AccessToken: accessToken,
	}, nil
}

func (repo *ChannelRepoImpl) DeleteChannel(ctx context.Context, channelId uint64) error {
	if err := repo.session.Query("DELETE FROM channels WHERE id = ?", channelId).WithContext(ctx).Exec(); err != nil {
		return err
	}

	return nil
}

func (repo *ChatRepoImpl) InsertMessage(ctx context.Context, chatMessage *Message) error {
	var messageNum int64

	err := repo.session.Query("SELECT message_num FROM chanmsg_counters WHERE channel_id = ? LIMIT 1", chatMessage.ChannelId).WithContext(ctx).Idempotent(true).Scan(&messageNum)
	if err != nil {
		if err == gocql.ErrNotFound {
			messageNum = 0
		} else {
			return err
		}
	}

	if messageNum >= repo.maxMessages {
		return common.ErrorExceedMessageNumLimits
	}

	if err := repo.session.Query("INSERT INTO messages (id, event, channel_id, user_id, payload, seen, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)",
		chatMessage.MessageId,
		chatMessage.Event,
		chatMessage.ChannelId,
		chatMessage.UserId,
		chatMessage.Payload,
		false,
		chatMessage.Time).WithContext(ctx).Exec(); err != nil {
		return err
	}

	return repo.session.Query("UPDATE chanmsg_counters SET message_num = message_num + 1 WHERE channel_id = ?", chatMessage.ChannelId).WithContext(ctx).Exec()
}

func (repo *ChatRepoImpl) MarkMessageSeen(ctx context.Context, channelId uint64, messageId uint64) error {
	if err := repo.session.Query("UPDATE messages SET seen = true WHERE channel_id = ? AND id = ?", channelId, messageId).
		WithContext(ctx).Idempotent(true).Exec(); err != nil {
		return err
	}

	return nil
}

func (repo *ChatRepoImpl) PublishMessage(ctx context.Context, chatMessage *Message) error {
	return repo.publisher.Publish(
		common.MessagePubTopic,
		message.NewMessage(
			watermill.NewUUID(),
			chatMessage.Encode(),
		),
	)
}

func (repo *ChatRepoImpl) ListMessages(ctx context.Context, channelId uint64, pageStateBase64 string) ([]*Message, string, error) {
	var messages []*Message

	pageState, err := base64.URLEncoding.DecodeString(pageStateBase64)
	if err != nil {
		return nil, "", err
	}

	iteration := repo.session.Query("SELECT id, event, channel_id, user_id, payload, seen, timestamp FROM messages WHERE channel_id = ?", channelId).
		WithContext(ctx).Idempotent(true).PageSize(repo.pagination).PageState(pageState).Iter()
	nextPageStateBase64 := base64.URLEncoding.EncodeToString(iteration.PageState())
	scanner := iteration.Scanner()

	for scanner.Next() {
		var message Message
		if err := scanner.Scan(
			&message.MessageId,
			&message.Event,
			&message.ChannelId,
			&message.UserId,
			&message.Payload,
			&message.Seen,
			&message.Time); err != nil {
			return nil, "", err
		}

		messages = append(messages, &message)
	}

	err = scanner.Err()
	if err != nil {
		return nil, "", err
	}

	return messages, nextPageStateBase64, nil
}

func (repo *ForwarderRepoImpl) RegisterChannelSession(ctx context.Context, channelId uint64, userId uint64, subscriber string) error {
	request := &forwarderProto.RegisterChannelSessionRequest{
		ChannelId:  channelId,
		UserId:     userId,
		Subscriber: subscriber,
	}
	_, err := repo.registerChannelSession(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ForwarderRepoImpl) RemoveChannelSession(ctx context.Context, channelId uint64, userId uint64) error {
	request := &forwarderProto.RemoveChannelSessionRequest{
		ChannelId: channelId,
		UserId:    userId,
	}
	_, err := repo.removeChannelSession(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

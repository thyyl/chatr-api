package match

import (
	"context"
	"strconv"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kit/kit/endpoint"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/infra"
	"github.com/thyyl/chatr/pkg/transport"
	chatProto "github.com/thyyl/chatr/proto/chat"
	userProto "github.com/thyyl/chatr/proto/user"
)

// ============================
// Repository Interfaces
// ============================
type ChannelRepo interface {
	CreateChannel(ctx context.Context) (uint64, string, error)
}

type UserRepo interface {
	GetUserById(ctx context.Context, userId uint64) (*User, error)
	GetUserIdBySession(ctx context.Context, session string) (uint64, error)
	AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error
}

type MatchRepo interface {
	PopOrPushWaitList(ctx context.Context, userId uint64) (bool, uint64, error)
	PublishMatchResult(ctx context.Context, result *MatchResult) error
	RemoveFromWaitList(ctx context.Context, userId uint64) error
}

// ============================
// Repository Implementations
// ============================
type ChannelRepoImpl struct {
	createChannel endpoint.Endpoint
}

func NewChannelRepoImpl(chatConn *ChatClientConn) *ChannelRepoImpl {
	return &ChannelRepoImpl{
		createChannel: transport.NewGrpcEndpoint(
			chatConn.Conn,
			"chat",
			"chat.ChannelService",
			"CreateChannel",
			&chatProto.CreateChannelResponse{},
		),
	}
}

type UserRepoImpl struct {
	getUserById        endpoint.Endpoint
	getUserIdBySession endpoint.Endpoint
	addUserToChannel   endpoint.Endpoint
}

func NewUserRepoImpl(userConn *UserClientConn, chatConn *ChatClientConn) *UserRepoImpl {
	return &UserRepoImpl{
		getUserById: transport.NewGrpcEndpoint(
			userConn.Conn,
			"user",
			"user.UserService",
			"GetUser",
			&userProto.GetUserResponse{},
		),
		getUserIdBySession: transport.NewGrpcEndpoint(
			userConn.Conn,
			"user",
			"user.UserService",
			"GetUserIdBySession",
			&userProto.GetUserIdBySessionResponse{},
		),
		addUserToChannel: transport.NewGrpcEndpoint(
			chatConn.Conn,
			"chat",
			"chat.UserService",
			"AddUserToChannel",
			&chatProto.AddUserResponse{},
		),
	}
}

type MatchRepoImpl struct {
	redis     infra.RedisCache
	publisher message.Publisher
}

func NewMatchRepoImpl(redis infra.RedisCache, publisher message.Publisher) *MatchRepoImpl {
	return &MatchRepoImpl{redis, publisher}
}

// ============================
// Repository Functions
// ============================
func (repo *ChannelRepoImpl) CreateChannel(ctx context.Context) (uint64, string, error) {
	response, err := repo.createChannel(ctx, &chatProto.CreateChannelRequest{})
	if err != nil {
		return 0, "", err
	}

	resp := response.(*chatProto.CreateChannelResponse)
	return resp.ChannelId, resp.AccessToken, nil
}

func (repo *UserRepoImpl) GetUserById(ctx context.Context, userID uint64) (*User, error) {
	response, err := repo.getUserById(ctx, &userProto.GetUserRequest{
		Id: userID,
	})
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

func (repo *UserRepoImpl) GetUserIdBySession(ctx context.Context, session string) (uint64, error) {
	response, err := repo.getUserIdBySession(ctx, &userProto.GetUserIdBySessionRequest{
		Session: session,
	})

	if err != nil {
		return 0, err
	}

	pbUserID := response.(*userProto.GetUserIdBySessionResponse)
	return pbUserID.Id, nil
}

func (repo *UserRepoImpl) AddUserToChannel(ctx context.Context, channelId uint64, userId uint64) error {
	_, err := repo.addUserToChannel(ctx, &chatProto.AddUserRequest{
		ChannelId: channelId,
		UserId:    userId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (repo *MatchRepoImpl) PopOrPushWaitList(ctx context.Context, userId uint64) (bool, uint64, error) {
	currentTime := time.Now().Unix()
	match, peerIdString, err := repo.redis.ZPopMinOrAddOne(ctx, common.UserWaitListRcKey, float64(currentTime), userId)
	if err != nil {
		return false, 0, err
	}
	if !match {
		return false, 0, nil
	}

	peerId, err := strconv.ParseUint(peerIdString, 10, 64)
	if err != nil {
		return false, 0, err
	}

	return true, peerId, nil
}

func (repo *MatchRepoImpl) RemoveFromWaitList(ctx context.Context, userId uint64) error {
	return repo.redis.ZRemOne(ctx, common.UserWaitListRcKey, userId)
}

func (repo *MatchRepoImpl) PublishMatchResult(ctx context.Context, result *MatchResult) error {
	return repo.publisher.Publish(common.MatchPubSubTopicRcKey, message.NewMessage(
		watermill.NewUUID(),
		common.Encode(result),
	))
}

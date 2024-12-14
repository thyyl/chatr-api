// wire.go
//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/thyyl/chatr/pkg/chat"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"github.com/thyyl/chatr/pkg/forwarder"
	"github.com/thyyl/chatr/pkg/infra"
	"github.com/thyyl/chatr/pkg/match"
	"github.com/thyyl/chatr/pkg/uploader"
	"github.com/thyyl/chatr/pkg/user"
)

func InitializeChatServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,
		common.NewHttpLog,
		common.NewGrpcLog,
		common.NewSonyFlake,

		infra.NewRedisClient,
		infra.NewRedisCacheImpl,
		wire.Bind(new(infra.RedisCache), new(*infra.RedisCacheImpl)),

		infra.NewKafkaPublisher,
		infra.NewKafkaSubscriber,
		infra.NewBrokerRouter,

		infra.NewCassandraSession,

		chat.NewUserClientConn,
		chat.NewForwarderClientConn,

		chat.NewUserRepoImpl,
		wire.Bind(new(chat.UserRepo), new(*chat.UserRepoImpl)),
		chat.NewChannelRepoImpl,
		wire.Bind(new(chat.ChannelRepo), new(*chat.ChannelRepoImpl)),
		chat.NewChatRepoImpl,
		wire.Bind(new(chat.ChatRepo), new(*chat.ChatRepoImpl)),
		chat.NewForwarderRepoImpl,
		wire.Bind(new(chat.ForwarderRepo), new(*chat.ForwarderRepoImpl)),

		chat.NewUserRepoCacheImpl,
		wire.Bind(new(chat.UserRepoCache), new(*chat.UserRepoCacheImpl)),
		chat.NewChannelRepoCacheImpl,
		wire.Bind(new(chat.ChannelRepoCache), new(*chat.ChannelRepoCacheImpl)),
		chat.NewChatRepoCacheImpl,
		wire.Bind(new(chat.ChatRepoCache), new(*chat.ChatRepoCacheImpl)),

		chat.NewMelodyChat,
		chat.NewMessageSubscriber,

		chat.NewUserServiceImpl,
		wire.Bind(new(chat.UserService), new(*chat.UserServiceImpl)),
		chat.NewChannelServiceImpl,
		wire.Bind(new(chat.ChannelService), new(*chat.ChannelServiceImpl)),
		chat.NewChatServiceImpl,
		wire.Bind(new(chat.ChatService), new(*chat.ChatServiceImpl)),
		chat.NewForwarderServiceImpl,
		wire.Bind(new(chat.ForwarderService), new(*chat.ForwarderServiceImpl)),

		chat.NewGinServer,

		chat.NewHttpServer,
		wire.Bind(new(common.HttpServer), new(*chat.HttpServer)),
		chat.NewGrpcServer,
		wire.Bind(new(common.GrpcServer), new(*chat.GrpcServer)),
		chat.NewRouter,
		wire.Bind(new(common.Router), new(*chat.Router)),
		chat.NewInfraCloser,
		wire.Bind(new(common.InfraCloser), new(*chat.InfraCloser)),
		common.NewServer,
	)

	return &common.Server{}, nil
}

func InitializeForwarderServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,
		common.NewGrpcLog,

		infra.NewRedisClient,
		infra.NewRedisCacheImpl,
		wire.Bind(new(infra.RedisCache), new(*infra.RedisCacheImpl)),

		infra.NewKafkaPublisher,
		infra.NewKafkaSubscriber,
		infra.NewBrokerRouter,

		forwarder.NewForwarderRepoImpl,
		wire.Bind(new(forwarder.ForwarderRepo), new(*forwarder.ForwarderRepoImpl)),

		forwarder.NewForwarderServiceImpl,
		wire.Bind(new(forwarder.ForwarderService), new(*forwarder.ForwarderServiceImpl)),

		forwarder.NewMessageSubscriber,

		forwarder.NewGrpcServer,
		wire.Bind(new(common.GrpcServer), new(*forwarder.GrpcServer)),
		forwarder.NewRouter,
		wire.Bind(new(common.Router), new(*forwarder.Router)),
		forwarder.NewInfraCloser,
		wire.Bind(new(common.InfraCloser), new(*forwarder.InfraCloser)),
		common.NewServer,
	)
	return &common.Server{}, nil
}

func InitializeMatchServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,
		common.NewHttpLog,

		infra.NewRedisClient,
		infra.NewRedisCacheImpl,
		wire.Bind(new(infra.RedisCache), new(*infra.RedisCacheImpl)),

		infra.NewKafkaPublisher,
		infra.NewKafkaSubscriber,
		infra.NewBrokerRouter,

		match.NewMelodyMatchConn,
		match.NewGinServer,

		match.NewChatClientConn,
		match.NewUserClientConn,

		match.NewChannelRepoImpl,
		wire.Bind(new(match.ChannelRepo), new(*match.ChannelRepoImpl)),
		match.NewUserRepoImpl,
		wire.Bind(new(match.UserRepo), new(*match.UserRepoImpl)),
		match.NewMatchRepoImpl,
		wire.Bind(new(match.MatchRepo), new(*match.MatchRepoImpl)),

		match.NewMatchSubscriber,

		match.NewUserServiceImpl,
		wire.Bind(new(match.UserService), new(*match.UserServiceImpl)),
		match.NewMatchServiceImpl,
		wire.Bind(new(match.MatchService), new(*match.MatchServiceImpl)),

		match.NewHttpServer,
		wire.Bind(new(common.HttpServer), new(*match.HttpServer)),
		match.NewRouter,
		wire.Bind(new(common.Router), new(*match.Router)),
		match.NewInfraCloser,
		wire.Bind(new(common.InfraCloser), new(*match.InfraCloser)),
		common.NewServer,
	)
	return &common.Server{}, nil
}

func InitializeUploaderServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,
		common.NewHttpLog,

		infra.NewRedisClient,

		uploader.NewGinServer,

		uploader.NewChannelUploadRateLimiter,

		uploader.NewHttpServer,
		wire.Bind(new(common.HttpServer), new(*uploader.HttpServer)),
		uploader.NewRouter,
		wire.Bind(new(common.Router), new(*uploader.Router)),
		uploader.NewInfraCloser,
		wire.Bind(new(common.InfraCloser), new(*uploader.InfraCloser)),
		common.NewServer,
	)
	return &common.Server{}, nil
}

func InitializeUserServer(name string) (*common.Server, error) {
	wire.Build(
		config.NewConfig,
		common.NewHttpLog,
		common.NewGrpcLog,

		infra.NewRedisClient,
		infra.NewRedisCacheImpl,
		wire.Bind(new(infra.RedisCache), new(*infra.RedisCacheImpl)),

		user.NewUserRepoImpl,
		wire.Bind(new(user.UserRepo), new(*user.UserRepoImpl)),

		common.NewSonyFlake,

		user.NewUserServiceImpl,
		wire.Bind(new(user.UserService), new(*user.UserServiceImpl)),

		user.NewGinServer,

		user.NewHttpServer,
		wire.Bind(new(common.HttpServer), new(*user.HttpServer)),
		user.NewGrpcServer,
		wire.Bind(new(common.GrpcServer), new(*user.GrpcServer)),
		user.NewRouter,
		wire.Bind(new(common.Router), new(*user.Router)),
		user.NewInfraCloser,
		wire.Bind(new(common.InfraCloser), new(*user.InfraCloser)),
		common.NewServer,
	)
	return &common.Server{}, nil
}

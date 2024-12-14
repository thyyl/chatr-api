package chat

import (
	"log/slog"
	"net"
	"os"

	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"github.com/thyyl/chatr/pkg/transport"
	chatProto "github.com/thyyl/chatr/proto/chat"
	"google.golang.org/grpc"
)

var UserConn *UserClientConn

type UserClientConn struct {
	Conn *grpc.ClientConn
}

func NewUserClientConn(config *config.Config) (*UserClientConn, error) {
	conn, err := transport.InitializeGrpcClient(config.Chat.Grpc.Client.User.Endpoint)
	if err != nil {
		return nil, err
	}

	UserConn = &UserClientConn{Conn: conn}
	return UserConn, nil
}

var ForwarderConn *ForwarderClientConn

type ForwarderClientConn struct {
	Conn *grpc.ClientConn
}

func NewForwarderClientConn(config *config.Config) (*ForwarderClientConn, error) {
	conn, err := transport.InitializeGrpcClient(config.Chat.Grpc.Client.Forwarder.Endpoint)
	if err != nil {
		return nil, err
	}

	ForwarderConn = &ForwarderClientConn{Conn: conn}
	return ForwarderConn, nil
}

type GrpcServer struct {
	grpcPort       string
	logger         common.GrpcLog
	server         *grpc.Server
	userService    UserService
	channelService ChannelService
	*chatProto.UnimplementedChannelServiceServer
	*chatProto.UnimplementedUserServiceServer
}

func NewGrpcServer(name string, logger common.GrpcLog, config *config.Config, userService UserService, channelService ChannelService) *GrpcServer {
	grpcServer := &GrpcServer{
		grpcPort:       config.Chat.Grpc.Server.Port,
		logger:         logger,
		userService:    userService,
		channelService: channelService,
	}

	grpcServer.server = transport.InitializeGrpcServer(name, grpcServer.logger)
	return grpcServer
}

func (s *GrpcServer) RegisterServices() {
	chatProto.RegisterChannelServiceServer(s.server, s)
	chatProto.RegisterUserServiceServer(s.server, s)
}

func (s *GrpcServer) Run() {
	go func() {
		addr := "0.0.0.0:" + s.grpcPort
		s.logger.Info("grpc server listening", slog.String("addr", addr))
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
		if err := s.server.Serve(lis); err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *GrpcServer) GracefulStop() error {
	s.server.GracefulStop()
	return nil
}

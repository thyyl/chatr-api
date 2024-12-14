package user

import (
	"log/slog"
	"net"
	"os"

	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"github.com/thyyl/chatr/pkg/transport"
	userProto "github.com/thyyl/chatr/proto/user"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	grpcPort    string
	logger      common.GrpcLog
	server      *grpc.Server
	userService UserService
	userProto.UnimplementedUserServiceServer
}

func NewGrpcServer(name string, config *config.Config, logger common.GrpcLog, userService UserService) *GrpcServer {
	grpcServer := &GrpcServer{
		grpcPort:    config.Users.Grpc.Server.Port,
		logger:      logger,
		userService: userService,
	}

	grpcServer.server = transport.InitializeGrpcServer(name, logger)
	return grpcServer
}

func (s *GrpcServer) RegisterServices() {
	userProto.RegisterUserServiceServer(s.server, s)
}

func (s *GrpcServer) Run() {
	go func() {
		address := "0.0.0.0:" + s.grpcPort
		s.logger.Info("GRPC server listening", slog.String("address", address))
		listener, err := net.Listen("tcp", address)
		if err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}

		if err := s.server.Serve(listener); err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *GrpcServer) GracefulStop() error {
	s.server.GracefulStop()
	return nil
}

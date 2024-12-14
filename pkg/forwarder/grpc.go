package forwarder

import (
	"log/slog"
	"net"
	"os"

	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"github.com/thyyl/chatr/pkg/transport"
	forwarderProto "github.com/thyyl/chatr/proto/forwarder"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	grpcPort          string
	logger            common.GrpcLog
	server            *grpc.Server
	forwarderService  ForwarderService
	messageSubscriber *MessageSubscriber
	forwarderProto.UnimplementedForwarderServiceServer
}

func NewGrpcServer(name string, logger common.GrpcLog, config *config.Config, forwarderService ForwarderService, messageSubscriber *MessageSubscriber) *GrpcServer {
	grpcServer := &GrpcServer{
		grpcPort:          config.Forwarder.Grpc.Server.Port,
		logger:            logger,
		forwarderService:  forwarderService,
		messageSubscriber: messageSubscriber,
	}
	grpcServer.server = transport.InitializeGrpcServer(name, grpcServer.logger)
	return grpcServer
}

func (s *GrpcServer) RegisterServices() {
	s.messageSubscriber.RegisterHandler()
	forwarderProto.RegisterForwarderServiceServer(s.server, s)
}

func (s *GrpcServer) Run() {
	go func() {
		address := "0.0.0.0:" + s.grpcPort
		s.logger.Info("grpc server listening", slog.String("addr", address))

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

	go func() {
		err := s.messageSubscriber.Run()
		if err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *GrpcServer) GracefulStop() error {
	s.server.GracefulStop()
	return s.messageSubscriber.GracefulStop()
}

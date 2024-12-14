package forwarder

import (
	"context"

	forwarderProto "github.com/thyyl/chatr/proto/forwarder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GrpcServer) RegisterChannelSession(ctx context.Context, req *forwarderProto.RegisterChannelSessionRequest) (*forwarderProto.RegisterChannelSessionResponse, error) {
	if err := s.forwarderService.RegisterChannelSession(ctx, req.ChannelId, req.UserId, req.Subscriber); err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &forwarderProto.RegisterChannelSessionResponse{}, nil
}

func (s *GrpcServer) RemoveChannelSession(ctx context.Context, req *forwarderProto.RemoveChannelSessionRequest) (*forwarderProto.RemoveChannelSessionResponse, error) {
	if err := s.forwarderService.RemoveChannelSession(ctx, req.ChannelId, req.UserId); err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &forwarderProto.RemoveChannelSessionResponse{}, nil
}

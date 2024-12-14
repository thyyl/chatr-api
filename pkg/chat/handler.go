package chat

import (
	"context"

	chatProto "github.com/thyyl/chatr/proto/chat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GrpcServer) CreateChannel(ctx context.Context, request *chatProto.CreateChannelRequest) (*chatProto.CreateChannelResponse, error) {
	channel, err := s.channelService.CreateChannel(ctx)
	if err != nil {
		return nil, err
	}

	return &chatProto.CreateChannelResponse{
		ChannelId:   channel.Id,
		AccessToken: channel.AccessToken,
	}, nil
}

func (s *GrpcServer) AddUserToChannel(ctx context.Context, request *chatProto.AddUserRequest) (*chatProto.AddUserResponse, error) {
	err := s.userService.AddUserToChannel(ctx, request.ChannelId, request.UserId)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &chatProto.AddUserResponse{}, nil
}

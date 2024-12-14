package user

import (
	"context"
	"errors"

	"github.com/thyyl/chatr/pkg/common"
	userProto "github.com/thyyl/chatr/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GrpcServer) GetUser(ctx context.Context, request *userProto.GetUserRequest) (*userProto.GetUserResponse, error) {
	userId := request.Id

	user, err := s.userService.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, common.ErrorUserNotFound) {
			return &userProto.GetUserResponse{
				Exist: false,
			}, nil
		}

		s.logger.Error(err.Error())
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	return &userProto.GetUserResponse{
		Exist: true,
		User: &userProto.User{
			Id:   user.Id,
			Name: user.Name,
		},
	}, nil
}

func (s *GrpcServer) GetUserIdBySession(ctx context.Context, request *userProto.GetUserIdBySessionRequest) (*userProto.GetUserIdBySessionResponse, error) {
	session := request.Session

	userId, err := s.userService.GetUserIdBySession(ctx, session)
	if err != nil {
		if errors.Is(err, common.ErrorSessionNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		s.logger.Error(err.Error())
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	return &userProto.GetUserIdBySessionResponse{
		Id: userId,
	}, nil
}

package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/thyyl/chatr/pkg/common"
)

type UserService interface {
	GetGoogleUser(ctx context.Context, code string) (*GoogleUserDto, error)
	GetOrCreateUserByOAuth(ctx context.Context, user *User) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	SetUserSession(ctx context.Context, uid uint64) (string, error)
	GetUserById(ctx context.Context, uid uint64) (*User, error)
	GetUserIdBySession(ctx context.Context, sid string) (uint64, error)
}

type UserServiceImpl struct {
	userRepo UserRepo
	sf       common.IDGenerator
}

func NewUserServiceImpl(userRepo UserRepo, sf common.IDGenerator) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		sf:       sf,
	}
}

func (s *UserServiceImpl) GetGoogleUser(ctx context.Context, accessToken string) (*GoogleUserDto, error) {
	req, err := http.NewRequest("GET", common.Join(common.OAuthGoogleUrlAPI, accessToken), nil)
	if err != nil {
		return nil, fmt.Errorf("create http request error: %w", err)
	}
	req = req.WithContext(ctx)

	client := http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer func() {
		err = response.Body.Close()
	}()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read google user response: %w", err)
	}
	var googleUser GoogleUserDto
	if err := json.Unmarshal(contents, &googleUser); err != nil {
		return nil, fmt.Errorf("failed marshal google user response: %w", err)
	}
	return &googleUser, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user *User) (*User, error) {
	userId, err := s.sf.NextID()
	if err != nil {
		return nil, fmt.Errorf("error create snowflake ID: %w", err)
	}
	newUser := &User{
		Id:       userId,
		Email:    user.Email,
		Name:     user.Name,
		Photo:    user.Photo,
		AuthType: user.AuthType,
	}
	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("error create user %d: %w", userId, err)
	}
	return newUser, nil
}

func (s *UserServiceImpl) SetUserSession(ctx context.Context, uid uint64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error create sid: %w", err)
	}
	sid := base64.URLEncoding.EncodeToString(b)
	if err := s.userRepo.SetUserSession(ctx, uid, sid); err != nil {
		return "", fmt.Errorf("error set sid for user %d: %w", uid, err)
	}
	return sid, nil
}

func (s *UserServiceImpl) GetUserById(ctx context.Context, uid uint64) (*User, error) {
	user, err := s.userRepo.GetUserById(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error get user %d: %w", uid, err)
	}
	return user, nil
}

func (s *UserServiceImpl) GetUserIdBySession(ctx context.Context, sid string) (uint64, error) {
	userId, err := s.userRepo.GetUserIdBySession(ctx, sid)
	if err != nil {
		return 0, fmt.Errorf("error get user id by sid %s: %w", sid, err)
	}
	return userId, nil
}

func (s *UserServiceImpl) GetOrCreateUserByOAuth(ctx context.Context, user *User) (*User, error) {
	existedUser, err := s.userRepo.GetUserByOAuthEmail(ctx, user.AuthType, user.Email)
	if err != nil {
		if !errors.Is(err, common.ErrorUserNotFound) {
			return nil, fmt.Errorf("error get user by google email %s: %w", user.Email, err)
		}
		userId, err := s.sf.NextID()
		if err != nil {
			return nil, fmt.Errorf("error create snowflake ID: %w", err)
		}
		newUser := &User{
			Id:       userId,
			Email:    user.Email,
			Name:     user.Name,
			Photo:    user.Photo,
			AuthType: user.AuthType,
		}
		if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
			return nil, fmt.Errorf("error create user by google email %s: %w", newUser.Email, err)
		}
		return newUser, nil
	}
	return existedUser, nil
}

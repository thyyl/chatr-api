package user

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/infra"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserById(ctx context.Context, userId uint64) (*User, error)
	GetUserByOAuthEmail(ctx context.Context, authType AuthType, email string) (*User, error)
	SetUserSession(ctx context.Context, userId uint64, session string) error
	GetUserIdBySession(ctx context.Context, session string) (uint64, error)
}

type UserRepoImpl struct {
	redis infra.RedisCache
}

func NewUserRepoImpl(redis infra.RedisCache) *UserRepoImpl {
	return &UserRepoImpl{redis}
}

func (repo *UserRepoImpl) CreateUser(ctx context.Context, user *User) error {
	data, err := json.Marshal(user)

	if err != nil {
		return err
	}

	userKey := constructKey(common.UserRcKey, user.Id)
	if err := repo.redis.Set(ctx, userKey, data); err != nil {
		return err
	}

	userAuthKey := constructOAuthKey(user.AuthType, user.Email)
	if err := repo.redis.Set(ctx, userAuthKey, data); err != nil {
		return err
	}

	return nil
}

func (repo *UserRepoImpl) GetUserById(ctx context.Context, userId uint64) (*User, error) {
	var user User
	key := constructKey(common.UserRcKey, userId)

	exist, err := repo.redis.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, common.ErrorUserNotFound
	}

	return &user, nil
}

func (repo *UserRepoImpl) GetUserByOAuthEmail(ctx context.Context, authType AuthType, email string) (*User, error) {
	var user User
	key := constructOAuthKey(authType, email)

	data, err := repo.redis.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	if !data {
		return nil, common.ErrorUserNotFound
	}

	return &user, nil
}

func (repo *UserRepoImpl) SetUserSession(ctx context.Context, userId uint64, session string) error {
	key := common.Join(common.SessionRcKey, ":", session)
	return repo.redis.Set(ctx, key, userId)
}

func (repo *UserRepoImpl) GetUserIdBySession(ctx context.Context, session string) (uint64, error) {
	key := common.Join(common.SessionRcKey, ":", session)
	var userId uint64

	exist, err := repo.redis.Get(ctx, key, &userId)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, common.ErrorSessionNotFound
	}

	return userId, nil
}

func constructKey(prefix string, id uint64) string {
	return common.Join(prefix, ":", strconv.FormatUint(id, 10))
}

func constructOAuthKey(authType AuthType, email string) string {
	return common.Join(common.UserRcKey, ":", string(authType), ":", email)
}

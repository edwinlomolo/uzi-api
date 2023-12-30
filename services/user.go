package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/cache"
	"github.com/3dw1nM0535/uzi-api/store"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ErrNoUser = errors.New("userservice: no user found")
var ErrNoCacheValue = errors.New("userservicecache: no value for key")
var userService User

type User interface {
	FindOrCreate(user model.SigninInput) (*model.User, error)
}

type userClient struct {
	store  *store.Queries
	logger *logrus.Logger
	ctx    context.Context
	cache  cache.Cache
}

func GetUserService() User {
	return userService
}

func NewUserService(store *store.Queries, redis *redis.Client, logger *logrus.Logger) User {
	c := newusercache(redis, logger)
	userService = &userClient{store, logger, context.TODO(), c}
	return userService
}

func (u *userClient) FindOrCreate(user model.SigninInput) (*model.User, error) {
	foundUser, foundUserErr := u.getUser(user.Phone)
	if foundUserErr == ErrNoUser {
		return u.createUser(user)
	} else if foundUserErr != nil {
		return nil, foundUserErr
	}

	return foundUser, nil
}

func (u *userClient) createUser(user model.SigninInput) (*model.User, error) {
	newUser, newUserErr := u.store.CreateUser(context.Background(), store.CreateUserParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	})
	if newUserErr != nil {
		u.logger.Errorf("%s-%v", "createUserErr", newUserErr.Error())
		return nil, newUserErr
	}

	return &model.User{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Phone:     newUser.Phone,
	}, nil
}

func (u *userClient) getUser(phone string) (*model.User, error) {
	var user model.User
	cacheUser, cacheErr := u.cache.Get(phone)
	if cacheErr == ErrNoCacheValue {
		foundUser, getUserErr := u.store.FindByPhone(context.Background(), phone)
		if getUserErr == sql.ErrNoRows {
			return nil, ErrNoUser
		} else if getUserErr != nil {
			u.logger.Errorf("%s-%v", "getUserErr", getUserErr.Error())
			return nil, getUserErr
		}

		user.ID = foundUser.ID
		user.FirstName = foundUser.FirstName
		user.LastName = foundUser.LastName
		user.Phone = foundUser.Phone

		if err := u.cache.Set(phone, &user); err != nil {
			return nil, err
		}

		return &user, nil
	}

	return (cacheUser).(*model.User), nil
}

type usercacheclient struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func newusercache(redis *redis.Client, logger *logrus.Logger) cache.Cache {
	return &usercacheclient{redis, logger}
}

func (usc *usercacheclient) Get(key string) (interface{}, error) {
	var res model.User

	keyValue, err := usc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, ErrNoCacheValue
	} else if err != nil {
		usc.logger.Errorf("%s-%v", "UserServiceCacheGetErr", err.Error())
		return nil, err
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		usc.logger.Errorf("%s-%v", "UserServiceCacheGetValueParsingErr", err.Error())
		return nil, err
	}

	return &res, nil
}

func (usc *usercacheclient) Set(key string, value interface{}) error {
	userinfo := value.(*model.User)
	data, err := json.Marshal(userinfo)
	if err != nil {
		usc.logger.Errorf("%s-%v", "UserServiceCacheSetErr", err.Error())
		return err
	}

	if err := usc.redis.Set(context.Background(), key, data, time.Minute*1).Err(); err != nil {
		usc.logger.Errorf("%s-%v", "UserServiceCacheSetOperationErr", err.Error())
		return err
	}

	return nil
}

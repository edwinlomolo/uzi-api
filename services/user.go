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

var (
	userService User
)

type User interface {
	FindOrCreate(user model.SigninInput) (*model.User, error)
	OnboardUser(user model.SigninInput) (*model.User, error)
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
	if foundUser == nil && foundUserErr == nil {
		return u.createUser(user)
	} else if foundUserErr != nil {
		return nil, foundUserErr
	}

	return foundUser, nil
}

func (u *userClient) createUser(user model.SigninInput) (*model.User, error) {
	var res model.User

	foundUser, getUserErr := u.store.FindByPhone(context.Background(), user.Phone)
	if getUserErr == sql.ErrNoRows {
		newUser, newUserErr := u.store.CreateUser(context.Background(), store.CreateUserParams{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
		})
		if newUserErr != nil {
			err := model.UziErr{Err: newUserErr.Error(), Message: "create user error", Code: 500}
			u.logger.Errorf("%s: %s", err.Message, err.Err)
			return nil, err
		}

		res.ID = newUser.ID
		res.FirstName = newUser.FirstName
		res.LastName = newUser.LastName
		res.Phone = newUser.Phone

		return &res, nil
	} else if getUserErr != nil {
		err := model.UziErr{Err: getUserErr.Error(), Message: "get user", Code: 500}
		u.logger.Errorf("%s: %s", err.Message, err.Err)
		return nil, err
	}

	res.ID = foundUser.ID
	res.FirstName = foundUser.FirstName
	res.LastName = foundUser.LastName
	res.Phone = foundUser.Phone

	return &res, nil
}

func (u *userClient) getUser(phone string) (*model.User, error) {
	cacheUser, cacheErr := u.cache.Get(phone)
	if cacheErr == nil && cacheUser == nil {
		return nil, nil
	} else if cacheErr != nil {
		err := model.UziErr{Err: cacheErr.Error(), Message: "user cache get", Code: 500}
		u.logger.Errorf("%s: %s", err.Message, err.Err)
		return nil, err
	}

	return (cacheUser).(*model.User), nil
}

func (u *userClient) OnboardUser(user model.SigninInput) (*model.User, error) {
	var updatedUser model.User

	if len(user.FirstName) == 0 || len(user.LastName) == 0 {
		inputErr := model.UziErr{Err: errors.New("name can't be empty").Error(), Message: "name can't be empty", Code: 400}
		u.logger.Errorf("%s: %s", inputErr.Message, inputErr.Err)
		return nil, inputErr
	}

	newUser, onboardErr := u.store.UpdateUserName(context.Background(), store.UpdateUserNameParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	})
	if onboardErr != nil {
		err := model.UziErr{Err: onboardErr.Error(), Message: "updateusername", Code: 500}
		u.logger.Errorf("%s: %v", err.Message, err.Err)
		return nil, err
	}

	if _, err := u.store.SetOnboardingStatus(context.Background(), store.SetOnboardingStatusParams{
		Phone:      user.Phone,
		Onboarding: false,
	}); err != nil {
		onboardingErr := model.UziErr{Err: err.Error(), Message: "setuseronboarding", Code: 500}
		u.logger.Errorf("%s: %v", onboardingErr.Message, onboardingErr.Err)
		return nil, onboardingErr
	}

	updatedUser.ID = newUser.ID
	updatedUser.FirstName = newUser.FirstName
	updatedUser.LastName = newUser.LastName
	updatedUser.Phone = newUser.Phone

	if err := u.cache.Set(user.Phone, &updatedUser); err != nil {
		return nil, model.UziErr{Err: err.Error(), Message: "cache set err", Code: 500}
	}

	return &updatedUser, nil
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
		return nil, nil
	} else if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "getusercache", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := model.UziErr{Err: err.Error(), Message: "getusercachemarshal", Code: 400}
		usc.logger.Errorf("%s: %s", jsonErr.Message, jsonErr.Err)
		return nil, jsonErr
	}

	return &res, nil
}

func (usc *usercacheclient) Set(key string, value interface{}) error {
	userinfo := value.(*model.User)
	data, err := json.Marshal(userinfo)
	if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setusercachemarshal", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	if err := usc.redis.Set(context.Background(), key, data, time.Minute*1).Err(); err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setusercache", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	return nil
}

package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/util"
	"github.com/edwinlomolo/uzi-api/services/courier"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	FindOrCreate(user SigninInput) (*model.User, error)
	OnboardUser(user SigninInput) (*model.User, error)
	GetUserByPhone(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
}

var (
	User         UserService
	userNotFound = errors.New("user not found")
	noEmptyName  = errors.New("name can't be empty")
)

type SigninInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Courier   bool   `json:"courier"`
}

type userClient struct {
	store  *sqlStore.Queries
	logger *logrus.Logger
	ctx    context.Context
	cache  cache.Cache
}

func NewUserService() {
	User = &userClient{store.DB, logger.Logger, context.TODO(), newCache()}
	logger.Logger.Infoln("User service...OK")
}

func (u *userClient) FindOrCreate(user SigninInput) (*model.User, error) {
	foundUser, foundUserErr := u.getUser(user.Phone)
	if foundUser == nil && foundUserErr == nil {
		return u.createUser(user)
	} else if foundUserErr != nil {
		return nil, foundUserErr
	}

	return foundUser, nil
}

func (u *userClient) createUser(user SigninInput) (*model.User, error) {
	var res model.User

	newUser, newUserErr := u.store.CreateUser(context.Background(), sqlStore.CreateUserParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	})
	if newUserErr != nil {
		err := fmt.Errorf("%s:%v", "create user", newUserErr)
		u.logger.Errorf(err.Error())
		return nil, err
	}

	if user.Courier {
		if _, courierErr := courier.Courier.FindOrCreate(newUser.ID); courierErr != nil {
			return nil, courierErr
		}
	}

	res.ID = newUser.ID
	res.FirstName = newUser.FirstName
	res.LastName = newUser.LastName
	res.Phone = newUser.Phone

	return &res, nil
}

func (u *userClient) getUser(phone string) (*model.User, error) {
	var user model.User
	cacheKey := util.Base64Key(phone)

	cacheUser, cacheErr := u.cache.Get(cacheKey)
	if cacheErr == nil && cacheUser == nil {
		foundUser, getErr := u.store.FindByPhone(context.Background(), phone)
		if getErr == sql.ErrNoRows {
			return nil, nil
		} else if getErr != nil {
			err := fmt.Errorf("%s:%v", "get user by phone", getErr)
			u.logger.Errorf(err.Error())
			return nil, err
		}

		user.ID = foundUser.ID
		user.FirstName = foundUser.FirstName
		user.LastName = foundUser.LastName
		user.Phone = foundUser.Phone

		go u.cache.Set(cacheKey, &user)

		return &user, nil
	} else if cacheErr != nil {
		return nil, cacheErr
	}

	return (cacheUser).(*model.User), nil
}

func (u *userClient) GetUserByPhone(phone string) (*model.User, error) {
	return u.getUser(phone)
}

func (u *userClient) findUserByID(id uuid.UUID) (*model.User, error) {
	cacheKey := util.Base64Key(id.String())
	cacheUser, cacheErr := u.cache.Get(cacheKey)
	if cacheUser == nil && cacheErr == nil {
		var user *model.User
		foundUser, getErr := u.store.FindUserByID(context.Background(), id)
		if getErr == sql.ErrNoRows {
			err := fmt.Errorf("%s:%v", "not found", userNotFound)
			u.logger.Errorf(err.Error())
			return nil, err
		} else if getErr != nil {
			err := fmt.Errorf("%s:%v", "find user by id", getErr)
			u.logger.Errorf(err.Error())
			return nil, err
		}

		user.ID = foundUser.ID
		user.FirstName = foundUser.FirstName
		user.LastName = foundUser.LastName
		user.Phone = foundUser.Phone

		go u.cache.Set(cacheKey, user)

		return user, nil
	} else if cacheErr != nil {
		return nil, cacheErr
	}

	return (cacheUser).(*model.User), nil
}

func (u *userClient) FindUserByID(id uuid.UUID) (*model.User, error) {
	return u.findUserByID(id)
}

func (u *userClient) OnboardUser(user SigninInput) (*model.User, error) {
	var updatedUser model.User
	cacheKey := util.Base64Key(user.Phone)

	if len(user.FirstName) == 0 || len(user.LastName) == 0 {
		inputErr := fmt.Errorf("%s:%v", "invalid inputs", noEmptyName)
		u.logger.Errorf(inputErr.Error())
		return nil, inputErr
	}

	newUser, onboardErr := u.store.UpdateUserName(context.Background(), sqlStore.UpdateUserNameParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	})
	if onboardErr != nil {
		err := fmt.Errorf("%s:%v", "update user name", onboardErr)
		u.logger.Errorf(err.Error())
		return nil, err
	}

	if _, err := u.store.SetOnboardingStatus(context.Background(), sqlStore.SetOnboardingStatusParams{
		Phone:      user.Phone,
		Onboarding: false,
	}); err != nil {
		onboardingErr := fmt.Errorf("%s:%v", "set user onboarding", err)
		u.logger.Errorf(onboardingErr.Error())
		return nil, onboardingErr
	}

	updatedUser.ID = newUser.ID
	updatedUser.FirstName = newUser.FirstName
	updatedUser.LastName = newUser.LastName
	updatedUser.Phone = newUser.Phone

	go u.cache.Set(cacheKey, &updatedUser)

	return &updatedUser, nil
}

package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	FindOrCreate(user SigninInput) (*model.User, error)
	OnboardUser(user SigninInput) (*model.User, error)
	GetUser(phone string) (*model.User, error)
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

	foundUser, getUserErr := u.store.FindByPhone(context.Background(), user.Phone)
	if getUserErr == sql.ErrNoRows {
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

		res.ID = newUser.ID
		res.FirstName = newUser.FirstName
		res.LastName = newUser.LastName
		res.Phone = newUser.Phone

		return &res, nil
	} else if getUserErr != nil {
		err := fmt.Errorf("%s:%v", "get user", getUserErr)
		u.logger.Errorf(err.Error())
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
		var user model.User
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

		if err := u.cache.Set(user.Phone, &user); err != nil {
			return nil, err
		}

		return &user, nil
	} else if cacheErr != nil {
		return nil, cacheErr
	}

	return (cacheUser).(*model.User), nil
}

func (u *userClient) GetUser(phone string) (*model.User, error) {
	return u.getUser(phone)
}

func (u *userClient) findUserByID(id uuid.UUID) (*model.User, error) {
	cacheUser, cacheErr := u.cache.Get(id.String())
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

		if err := u.cache.Set(user.ID.String(), user); err != nil {
			return nil, err
		}

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

	if err := u.cache.Set(user.Phone, &updatedUser); err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

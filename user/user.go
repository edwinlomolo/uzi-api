package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/courier"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/logger"
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
	store *sqlStore.Queries
	log   *logrus.Logger
	ctx   context.Context
}

func NewUserService() {
	User = &userClient{
		store.DB,
		logger.Logger,
		context.TODO(),
	}
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
	newUser, newUserErr := u.store.CreateUser(
		context.Background(),
		sqlStore.CreateUserParams{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
		})
	if newUserErr != nil {
		u.log.WithFields(logrus.Fields{
			"error": newUserErr,
			"user":  user,
		}).Errorf("create new user")
		return nil, newUserErr
	}

	if user.Courier {
		if _, courierErr := courier.Courier.FindOrCreate(newUser.ID); courierErr != nil {
			return nil, courierErr
		}
	}

	return &model.User{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Phone:     newUser.Phone,
	}, nil
}

func (u *userClient) getUser(phone string) (*model.User, error) {
	foundUser, getErr := u.store.FindByPhone(context.Background(), phone)
	if getErr == sql.ErrNoRows {
		return nil, nil
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"error": getErr,
			"phone": phone,
		}).Errorf("get user by phone")
		return nil, getErr
	}

	return &model.User{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Phone:     foundUser.Phone,
	}, nil
}

func (u *userClient) GetUserByPhone(phone string) (*model.User, error) {
	return u.getUser(phone)
}

func (u *userClient) findUserByID(id uuid.UUID) (*model.User, error) {
	foundUser, getErr := u.store.FindUserByID(context.Background(), id)
	if getErr == sql.ErrNoRows {
		u.log.WithFields(logrus.Fields{
			"user_id": id,
			"error":   userNotFound.Error(),
		}).Errorf(userNotFound.Error())
		return nil, userNotFound
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":   getErr,
			"user_id": id,
		}).Errorf("get user by id")
		return nil, getErr
	}

	return &model.User{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Phone:     foundUser.Phone,
	}, nil
}

func (u *userClient) FindUserByID(id uuid.UUID) (*model.User, error) {
	return u.findUserByID(id)
}

func (u *userClient) OnboardUser(user SigninInput) (*model.User, error) {
	if len(user.FirstName) == 0 || len(user.LastName) == 0 {
		inputErr := fmt.Errorf("%s:%v", "invalid inputs", noEmptyName)
		u.log.Errorf(inputErr.Error())
		return nil, inputErr
	}

	newUser, onboardErr := u.store.UpdateUserName(
		context.Background(),
		sqlStore.UpdateUserNameParams{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
		})
	if onboardErr != nil {
		err := fmt.Errorf("%s:%v", "update user", onboardErr)
		u.log.Errorf(err.Error())
		return nil, err
	}

	if _, err := u.store.SetOnboardingStatus(
		context.Background(),
		sqlStore.SetOnboardingStatusParams{
			Phone:      user.Phone,
			Onboarding: false,
		}); err != nil {
		onboardingErr := fmt.Errorf("%s:%v", "user onboarding", err)
		u.log.WithFields(logrus.Fields{
			"error": err,
			"phone": user.Phone,
		}).Errorf("set user onboarding status")
		return nil, onboardingErr
	}

	return &model.User{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Phone:     newUser.Phone,
	}, nil
}

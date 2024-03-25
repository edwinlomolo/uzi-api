package controllers

import (
	"github.com/edwinlomolo/uzi-api/gql/model"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
)

var (
	uService UserController
)

type UserController interface {
	FindOrCreate(user r.SigninInput) (*model.User, error)
	OnboardUser(user r.SigninInput) (*model.User, error)
	GetUserByPhone(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
	SignIn(signin r.SigninInput, ip, userAgent string) (*model.Session, error)
	SoftDelete(phone string) error
}

type userClient struct {
	r r.UserRepository
}

func NewUserController(q *sqlc.Queries) {
	ur := r.UserRepository{}
	ur.Init(q)
	uService = &userClient{ur}
}

func GetUserController() UserController {
	return uService
}

func (u *userClient) SignIn(
	signin r.SigninInput,
	ip,
	userAgent string,
) (*model.Session, error) {
	return u.r.SignIn(signin, ip, userAgent)
}

func (u *userClient) FindOrCreate(input r.SigninInput) (*model.User, error) {
	return u.r.FindOrCreate(input)
}

func (u *userClient) GetUserByPhone(phone string) (*model.User, error) {
	return u.r.GetUserByPhone(phone)
}

func (u *userClient) FindUserByID(id uuid.UUID) (*model.User, error) {
	return u.r.FindUserByID(id)
}

func (u *userClient) OnboardUser(user r.SigninInput) (*model.User, error) {
	return u.r.OnboardUser(user)
}

func (u *userClient) SoftDelete(phone string) error {
	return nil
}

package user

import (
	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/gql/model"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
)

type UserService interface {
	FindOrCreate(user r.SigninInput) (*model.User, error)
	OnboardUser(user r.SigninInput) (*model.User, error)
	GetUserByPhone(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
	SignIn(signin r.SigninInput, ip, userAgent string) (*model.Session, error)
}

type userClient struct {
	r r.UserRepository
}

func New(store *sqlc.Queries, redis cache.Cache) UserService {
	ur := r.UserRepository{}
	ur.Init(store, redis)
	return &userClient{ur}
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

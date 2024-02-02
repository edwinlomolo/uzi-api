package user

import (
	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/google/uuid"
)

type User interface {
	FindOrCreate(user SigninInput) (*model.User, error)
	OnboardUser(user SigninInput) (*model.User, error)
	GetUser(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
}

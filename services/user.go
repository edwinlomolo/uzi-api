package services

import (
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/google/uuid"
)

type User interface {
	FindOrCreate(user model.SigninInput) (*model.User, error)
	OnboardUser(user model.SigninInput) (*model.User, error)
	GetUser(phone string) (*model.User, error)
	FindUserByID(id uuid.UUID) (*model.User, error)
}

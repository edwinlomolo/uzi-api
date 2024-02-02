package session

import "github.com/3dw1nM0535/uzi-api/gql/model"

type Session interface {
	SignIn(user model.User, ip, userAgent string) (*model.Session, error)
}

package jwt

import "github.com/golang-jwt/jwt/v5"

type Jwt interface {
	Sign(secret []byte, claims jwt.Claims) (string, error)
	Validate(token string) (*jwt.Token, error)
}

type jwtclient struct{}

func NewJwtClient() Jwt {
	return &jwtclient{}
}

func (jwtc *jwtclient) Sign(secret []byte, claims jwt.Claims) (string, error) {
	return "", nil
}

func (jwtc *jwtclient) Validate(jwt string) (*jwt.Token, error) {
	return nil, nil
}

package jwt

import jsonwebtoken "github.com/golang-jwt/jwt/v5"

type Jwt interface {
	Sign(secret []byte, claims jsonwebtoken.Claims) (string, error)
	Validate(token string) (*jsonwebtoken.Token, error)
}

package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	Phone string `json:"phone"`
	jwt.RegisteredClaims
}

func NewPayload(id string, phone string, duration time.Duration) (*Payload, error) {
	p := &Payload{
		phone,
		jwt.RegisteredClaims{
			Issuer:    "Uzi",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        id,
		},
	}

	return p, nil
}

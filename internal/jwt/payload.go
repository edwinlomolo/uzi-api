package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	Phone string `json:"phone"`
	IP    string `json:"ip"`
	jwt.RegisteredClaims
}

func NewPayload(
	id, ip, phone string,
	duration time.Duration,
) (*Payload, error) {
	p := &Payload{
		phone,
		ip,
		jwt.RegisteredClaims{
			Issuer:    "Uzi",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        id,
		},
	}

	return p, nil
}

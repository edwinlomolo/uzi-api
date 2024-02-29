package jwt

import (
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var (
	Jwt                 JwtService
	ErrInvalidAlgorithm = errors.New("invalid signing algorithm")
)

type JwtService interface {
	Sign(secret []byte, claims jsonwebtoken.Claims) (string, error)
	Validate(token string) (*jsonwebtoken.Token, error)
}

type jwtClient struct {
	logger *logrus.Logger
	secret string
}

func NewJwtService() {
	Jwt = &jwtClient{logger.Logger, config.Config.Jwt.Secret}
}

func (jwtc *jwtClient) Sign(
	secret []byte,
	claims jsonwebtoken.Claims,
) (string, error) {
	token, signJwtErr := jsonwebtoken.NewWithClaims(
		jsonwebtoken.SigningMethodHS256,
		claims,
	).SignedString(secret)
	if signJwtErr != nil {
		uziErr := fmt.Errorf("%s:%v", "sign jwt", signJwtErr.Error())
		jwtc.logger.Errorf(uziErr.Error())
		return "", uziErr
	}

	return token, nil
}

func (jwtc *jwtClient) Validate(
	jwt string,
) (*jsonwebtoken.Token, error) {
	keyFunc := func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			jwtc.logger.Errorf(ErrInvalidAlgorithm.Error())
			return nil, ErrInvalidAlgorithm
		}

		return []byte(jwtc.secret), nil
	}

	token, tokenErr := jsonwebtoken.ParseWithClaims(jwt, &Payload{}, keyFunc)
	if tokenErr != nil {
		uziErr := fmt.Errorf("%s:%v", "parse claims", tokenErr.Error())
		jwtc.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return token, nil
}

package jwt

import (
	"errors"
	"fmt"

	"github.com/3dw1nM0535/uzi-api/config"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var (
	jwtService           Jwt
	invalidSignAlgorithm = errors.New("invalid signing algorithm")
)

type jwtclient struct {
	logger *logrus.Logger
	config config.Jwt
}

func GetJwtService() Jwt { return jwtService }

func NewJwtClient(logger *logrus.Logger, config config.Jwt) Jwt {
	jwtService = &jwtclient{logger, config}
	return jwtService
}

func (jwtc *jwtclient) Sign(secret []byte, claims jsonwebtoken.Claims) (string, error) {
	token, signJwtErr := jsonwebtoken.NewWithClaims(jsonwebtoken.SigningMethodHS256, claims).SignedString(secret)
	if signJwtErr != nil {
		uziErr := fmt.Errorf("%s:%v", "signjwt", signJwtErr.Error())
		jwtc.logger.Errorf(uziErr.Error())
		return "", uziErr
	}

	return token, nil
}

func (jwtc *jwtclient) Validate(jwt string) (*jsonwebtoken.Token, error) {
	keyFunc := func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			jwtc.logger.Errorf(invalidSignAlgorithm.Error())
			return nil, invalidSignAlgorithm
		}

		return []byte(jwtc.config.Secret), nil
	}

	token, tokenErr := jsonwebtoken.ParseWithClaims(jwt, &Payload{}, keyFunc)
	if tokenErr != nil {
		uziErr := fmt.Errorf("%s:%v", "invalidtoken", tokenErr.Error())
		jwtc.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return token, nil
}

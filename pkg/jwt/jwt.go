package jwt

import (
	"fmt"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/golang-jwt/jwt/v5"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Jwt interface {
	Sign(secret []byte, claims jwt.Claims) (string, error)
	Validate(token string) (*jwt.Token, error)
}

type jwtclient struct {
	logger *logrus.Logger
	config config.Jwt
}

func NewJwtClient(logger *logrus.Logger, config config.Jwt) Jwt {
	return &jwtclient{logger, config}
}

func (jwtc *jwtclient) Sign(secret []byte, claims jwt.Claims) (string, error) {
	token, signJwtErr := jsonwebtoken.NewWithClaims(jsonwebtoken.SigningMethodHS256, claims).SignedString(secret)
	if signJwtErr != nil {
		jwtc.logger.Errorf("%s-%v", "SignJwtErr", signJwtErr.Error())
		return "", &model.UziErr{Err: signJwtErr.Error(), Message: "signjwt", Code: 401}
	}

	return token, nil
}

func (jwtc *jwtclient) Validate(jwt string) (*jwt.Token, error) {
	keyFunc := func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			jwtc.logger.Errorf("%s-%v", "TokenParseErr", "invalid signing algorithm")
			return nil, fmt.Errorf("%s-%v", "invalid signing algorithm", tkn.Header["alg"])
		}

		return []byte(jwtc.config.Secret), nil
	}

	token, tokenErr := jsonwebtoken.Parse(jwt, keyFunc)
	if tokenErr != nil {
		jwtc.logger.Errorf("%s-%v", "TokenParseErr", tokenErr.Error())
		return nil, &model.UziErr{Err: tokenErr.Error(), Message: "invalid token", Code: 401}
	}

	return token, nil
}

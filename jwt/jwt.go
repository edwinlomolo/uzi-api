package jwt

import (
	"errors"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidAlgorithm = errors.New("invalid signing algorithm")
)

type JwtService interface {
	Sign(secret []byte, claims jsonwebtoken.Claims) (string, error)
	Validate(token string) (*jsonwebtoken.Token, error)
}

type jwtClient struct {
	log    *logrus.Logger
	secret string
}

func New() JwtService {
	return &jwtClient{logger.GetLogger(), config.Config.Jwt.Secret}
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
		jwtc.log.WithFields(logrus.Fields{
			"claims": claims,
			"error":  signJwtErr,
		}).Errorf("sign jwt")
		return "", signJwtErr
	}

	return token, nil
}

func (jwtc *jwtClient) Validate(
	jwt string,
) (*jsonwebtoken.Token, error) {
	keyFunc := func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			jwtc.log.WithFields(logrus.Fields{
				"jwt": jwt,
			}).Errorf(ErrInvalidAlgorithm.Error())
			return nil, ErrInvalidAlgorithm
		}

		return []byte(jwtc.secret), nil
	}

	token, tokenErr := jsonwebtoken.ParseWithClaims(jwt, &Payload{}, keyFunc)
	if tokenErr != nil {
		jwtc.log.WithFields(logrus.Fields{
			"error": tokenErr,
		}).Errorf("parse token claims")
		return nil, tokenErr
	}

	return token, nil
}

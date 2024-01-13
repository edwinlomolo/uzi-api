package jwt

import (
	"errors"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var jwtService Jwt

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
		uziErr := model.UziErr{Err: signJwtErr.Error(), Message: "signjwt", Code: http.StatusUnauthorized}
		jwtc.logger.Errorf(uziErr.Error())
		return "", uziErr
	}

	return token, nil
}

func (jwtc *jwtclient) Validate(jwt string) (*jsonwebtoken.Token, error) {
	keyFunc := func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			uziErr := model.UziErr{Err: errors.New("TokenParser").Error(), Message: "invalidsigningalgorithm", Code: http.StatusUnauthorized}
			jwtc.logger.Errorf(uziErr.Error())
			return nil, uziErr
		}

		return []byte(jwtc.config.Secret), nil
	}

	token, tokenErr := jsonwebtoken.Parse(jwt, keyFunc)
	if tokenErr != nil {
		uziErr := model.UziErr{Err: tokenErr.Error(), Message: "invalidtoken", Code: http.StatusUnauthorized}
		jwtc.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return token, nil
}

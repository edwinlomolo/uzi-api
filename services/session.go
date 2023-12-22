package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/netip"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/jwt"
	"github.com/3dw1nM0535/uzi-api/store"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Session interface {
	jwt.Jwt
	FindOrCreate(userID uuid.UUID, ipAddress netip.Addr) (*model.Session, error)
}

type sessionClient struct {
	jwtClient jwt.Jwt
	store     *store.Queries
	logger    *logrus.Logger
	config    config.Jwt
}

func NewSessionService(store *store.Queries, logger *logrus.Logger, jwtConfig config.Jwt) Session {
	return &sessionClient{jwt.NewJwtClient(logger, jwtConfig), store, logger, jwtConfig}
}

func (sc *sessionClient) FindOrCreate(userID uuid.UUID, ipAddress netip.Addr) (*model.Session, error) {
	foundSession, foundSessionErr := sc.store.GetSession(context.Background(), userID)
	if foundSessionErr == sql.ErrNoRows {
		claims := jsonwebtoken.MapClaims{
			"user_id": base64.StdEncoding.EncodeToString([]byte(userID.String())),
			"ip":      ipAddress,
			"exp":     sc.config.Expires,
			"iss":     "Uzi",
		}

		sessionJwt, sessionJwtErr := sc.Sign([]byte(sc.config.Secret), claims)
		newSession, err := sc.store.CreateSession(context.Background(), store.CreateSessionParams{
			UserID:  userID,
			Ip:      ipAddress,
			Token:   sessionJwt,
			Expires: sc.config.Expires.String(),
		})
		if sessionJwtErr != nil {
			sc.logger.Errorf("%s-%v", "CreateNewSessionErr", err.Error())
			return nil, err
		}

		return &model.Session{
			ID:        newSession.ID,
			IP:        newSession.Ip.String(),
			Token:     newSession.Token,
			UserID:    newSession.UserID,
			CreatedAt: &newSession.CreatedAt,
			UpdatedAt: &newSession.UpdatedAt,
		}, nil
	} else if foundSessionErr != nil {
		sc.logger.Errorf("%s-%v", "GetActiveSessionErr", foundSessionErr.Error())
		return nil, foundSessionErr
	}

	return &model.Session{
		ID:        foundSession.ID,
		IP:        foundSession.Ip.String(),
		Token:     foundSession.Token,
		UserID:    foundSession.UserID,
		CreatedAt: &foundSession.CreatedAt,
		UpdatedAt: &foundSession.UpdatedAt,
	}, nil
}

func (sc *sessionClient) Sign(secret []byte, claims jsonwebtoken.Claims) (string, error) {
	token, signJwtErr := sc.Sign(secret, claims)
	if signJwtErr != nil {
		sc.logger.Errorf("%s-%v", "SignJwtErr", signJwtErr.Error())
		return "", signJwtErr
	}

	return token, nil
}

func (sc *sessionClient) Validate(jwt string) (*jsonwebtoken.Token, error) {
	token, tokenErr := jsonwebtoken.Parse(jwt, func(tkn *jsonwebtoken.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jsonwebtoken.SigningMethodHMAC); !ok {
			sc.logger.Errorf("%s-%v", "TokenParseErr", "invalid signing algorithm")
			return nil, fmt.Errorf("%s-%v", "invalid signing algorithm", tkn.Header["alg"])
		}

		return []byte(sc.config.Secret), nil
	})
	if tokenErr != nil {
		sc.logger.Errorf("%s-%v", "TokenParseErr", tokenErr.Error())
		return nil, tokenErr
	}

	return token, nil
}

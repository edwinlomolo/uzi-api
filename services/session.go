package services

import (
	"context"
	"encoding/base64"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/jwt"
	"github.com/3dw1nM0535/uzi-api/store"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var sessionService Session

type Session interface {
	SignIn(userID uuid.UUID, ip, phone string) (*model.Session, error)
}

type sessionClient struct {
	jwtClient jwt.Jwt
	store     *store.Queries
	logger    *logrus.Logger
	config    config.Jwt
}

func GetSessionService() Session {
	return sessionService
}

func NewSessionService(store *store.Queries, logger *logrus.Logger, jwtConfig config.Jwt) Session {
	sessionService = &sessionClient{jwt.NewJwtClient(logger, jwtConfig), store, logger, jwtConfig}
	return sessionService
}

func (sc *sessionClient) SignIn(userID uuid.UUID, ip, phone string) (*model.Session, error) {
	return sc.createSession(userID, ip, phone)
}

func (sc *sessionClient) createSession(userID uuid.UUID, ip, phone string) (*model.Session, error) {
	var session model.Session

	claims := jsonwebtoken.MapClaims{
		"user_id": base64.StdEncoding.EncodeToString([]byte(userID.String())),
		"ip":      ip,
		"exp":     sc.config.Expires,
		"iss":     "Uzi",
	}

	sessionJwt, signJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, isUserOnboardingErr := sc.store.IsUserOnboarding(context.Background(), userID)
	if isUserOnboardingErr != nil {
		return nil, model.UziErr{Err: isUserOnboardingErr.Error(), Message: "isuseronboarding", Code: 500}
	}

	isCourier, isCourierErr := GetCourierService().IsCourier(userID)
	if isUserOnboardingErr != nil {
		return nil, isCourierErr
	}

	courierStatus, courierStatusErr := GetCourierService().GetCourierStatus(userID)
	if isUserOnboardingErr != nil {
		return nil, courierStatusErr
	}

	session.Token = sessionJwt
	session.IsCourier = isCourier
	session.Phone = phone
	session.Onboarding = isUserOnboarding
	session.CourierStatus = &courierStatus

	return &session, nil
}

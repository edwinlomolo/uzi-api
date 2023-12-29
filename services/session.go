package services

import (
	"context"
	"database/sql"
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
	FindOrCreate(uuid.UUID, string) (*model.Session, error)
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

func (sc *sessionClient) FindOrCreate(userID uuid.UUID, ipAddress string) (*model.Session, error) {
	isUserOnboarding, isUserOnboardingErr := sc.store.IsUserOnboarding(context.Background(), userID)
	if isUserOnboardingErr != nil {
		return nil, isUserOnboardingErr
	}

	isCourier, isCourierErr := GetCourierService().IsCourier(userID)
	if isUserOnboardingErr != nil {
		return nil, isCourierErr
	}

	courierStatus, courierStatusErr := GetCourierService().GetCourierStatus(userID)
	if isUserOnboardingErr != nil {
		return nil, courierStatusErr
	}

	foundSession, foundSessionErr := sc.store.GetSession(context.Background(), userID)
	if foundSessionErr == sql.ErrNoRows {
		claims := jsonwebtoken.MapClaims{
			"user_id": base64.StdEncoding.EncodeToString([]byte(userID.String())),
			"ip":      ipAddress,
			"exp":     sc.config.Expires,
			"iss":     "Uzi",
		}

		sessionJwt, sessionJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
		if sessionJwtErr != nil {
			sc.logger.Errorf("%s-%v", "SignSessionJwtErr", sessionJwtErr.Error())
			return nil, sessionJwtErr
		}

		newSession, newSessionErr := sc.store.CreateSession(context.Background(), store.CreateSessionParams{
			UserID:  userID,
			Ip:      ipAddress,
			Token:   sessionJwt,
			Expires: sc.config.Expires,
		})
		if newSessionErr != nil {
			sc.logger.Errorf("%s-%v", "CreateNewSessionErr", newSessionErr.Error())
			return nil, newSessionErr
		}

		return &model.Session{
			ID:         newSession.ID,
			IP:         newSession.Ip,
			Token:      newSession.Token,
			UserID:     newSession.UserID,
			Onboarding: isUserOnboarding,
			IsCourier:  isCourier,
			Status:     &courierStatus,
			Expires:    newSession.Expires,
			CreatedAt:  &newSession.CreatedAt,
			UpdatedAt:  &newSession.UpdatedAt,
		}, nil
	} else if foundSessionErr != nil {
		sc.logger.Errorf("%s-%v", "GetActiveSessionErr", foundSessionErr.Error())
		return nil, foundSessionErr
	}

	return &model.Session{
		ID:         foundSession.ID,
		IP:         foundSession.Ip,
		Token:      foundSession.Token,
		UserID:     foundSession.UserID,
		Onboarding: isUserOnboarding,
		IsCourier:  isCourier,
		Status:     &courierStatus,
		Expires:    foundSession.Expires,
		CreatedAt:  &foundSession.CreatedAt,
		UpdatedAt:  &foundSession.UpdatedAt,
	}, nil
}

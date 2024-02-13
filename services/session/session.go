package session

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/jwt"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/services/courier"
	userService "github.com/edwinlomolo/uzi-api/services/user"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var Session SessionService

type SessionService interface {
	SignIn(user model.User, ip, userAgent string) (*model.Session, error)
}

type sessionClient struct {
	jwtClient jwt.JwtService
	store     *sqlStore.Queries
	logger    *logrus.Logger
	config    config.Jwt
}

func NewSessionService() {
	Session = &sessionClient{jwt.Jwt, store.DB, logger.Logger, config.Config.Jwt}
	logger.Logger.Infoln("Session service...OK")
}

func (sc *sessionClient) SignIn(user model.User, ip, userAgent string) (*model.Session, error) {
	return sc.findOrCreate(user, ip, userAgent)
}

func (sc *sessionClient) findOrCreate(user model.User, ip, userAgent string) (*model.Session, error) {
	sess, sessErr := sc.getSession(user.ID)
	if sess == nil && sessErr == nil {
		newSess, newSessErr := sc.createNewSession(user.ID, ip, user.Phone, userAgent)
		if newSessErr != nil {
			return nil, newSessErr
		}

		return newSess, nil
	} else if sessErr != nil {
		return nil, sessErr
	}

	return sess, nil
}

func (sc *sessionClient) createNewSession(userID uuid.UUID, ip, phone, userAgent string) (*model.Session, error) {
	sessParams := sqlStore.CreateSessionParams{
		ID:        userID,
		Ip:        ip,
		UserAgent: userAgent,
		Phone:     phone,
	}
	newSession, newSessErr := sc.store.CreateSession(context.Background(), sessParams)
	if newSessErr != nil {
		err := fmt.Errorf("%s:%v", "create session", newSessErr)
		sc.logger.Errorf(err.Error())
		return nil, err
	}

	claims, err := jwt.NewPayload(userID.String(), ip, phone, sc.config.Expires)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, _ := sc.store.IsUserOnboarding(context.Background(), userID)

	isCourier, courierStatus, courierErr := sc.getRelevantCourierData(userID)
	if courierErr != nil {
		return nil, courierErr
	}

	user, userErr := userService.User.GetUser(phone)
	if userErr != nil {
		return nil, userErr
	}

	return &model.Session{
		ID:            newSession.ID,
		IP:            newSession.Ip,
		FirstName:     &user.FirstName,
		LastName:      &user.LastName,
		Phone:         newSession.Phone,
		UserAgent:     newSession.UserAgent,
		Token:         sessionJwt,
		CourierStatus: &courierStatus,
		Onboarding:    isUserOnboarding,
		IsCourier:     isCourier,
	}, nil
}

func (sc *sessionClient) getSession(sessionID uuid.UUID) (*model.Session, error) {
	foundSess, sessErr := sc.store.GetSession(context.Background(), sessionID)
	if sessErr == sql.ErrNoRows {
		return nil, nil
	} else if sessErr != nil {
		err := fmt.Errorf("%s:%v", "get session", sessErr)
		sc.logger.Errorf(err.Error())
		return nil, err
	}

	claims, err := jwt.NewPayload(foundSess.ID.String(), foundSess.Ip, foundSess.Phone, sc.config.Expires)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, _ := sc.store.IsUserOnboarding(context.Background(), foundSess.ID)

	isCourier, courierStatus, courierErr := sc.getRelevantCourierData(foundSess.ID)
	if courierErr != nil {
		return nil, courierErr
	}

	user, userErr := userService.User.GetUser(foundSess.Phone)
	if userErr != nil {
		return nil, userErr
	}

	return &model.Session{
		ID:            foundSess.ID,
		IP:            foundSess.Ip,
		FirstName:     &user.FirstName,
		LastName:      &user.LastName,
		Phone:         foundSess.Phone,
		UserAgent:     foundSess.UserAgent,
		Token:         sessionJwt,
		CourierStatus: &courierStatus,
		Onboarding:    isUserOnboarding,
		IsCourier:     isCourier,
	}, nil

}

func (sc *sessionClient) getRelevantCourierData(userID uuid.UUID) (bool, model.CourierStatus, error) {

	courierStatus, courierStatusErr := courier.Courier.GetCourierStatus(userID)
	if courierStatusErr != nil {
		return false, courierStatus, courierStatusErr
	}

	isCourier, isCourierErr := courier.Courier.IsCourier(userID)
	if isCourierErr != nil {
		return false, courierStatus, isCourierErr
	}

	return isCourier, courierStatus, nil
}

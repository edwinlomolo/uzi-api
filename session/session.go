package session

import (
	"context"
	"database/sql"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/courier"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/jwt"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/edwinlomolo/uzi-api/user"
	userService "github.com/edwinlomolo/uzi-api/user"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var Session SessionService

type SessionService interface {
	SignIn(signin user.SigninInput, ip, userAgent string) (*model.Session, error)
}

type sessionClient struct {
	jwtClient jwt.JwtService
	store     *sqlStore.Queries
	log       *logrus.Logger
	config    config.Jwt
}

func NewSessionService() {
	Session = &sessionClient{
		jwt.Jwt,
		store.DB,
		logger.Logger,
		config.Config.Jwt,
	}
	logger.Logger.Infoln("Session service...OK")
}

func (sc *sessionClient) SignIn(
	signin user.SigninInput,
	ip,
	userAgent string,
) (*model.Session, error) {
	return sc.findOrCreate(signin, ip, userAgent)
}

func (sc *sessionClient) findOrCreate(
	signin user.SigninInput,
	ip,
	userAgent string,
) (*model.Session, error) {
	user, userErr := user.User.FindOrCreate(signin)
	if userErr != nil {
		return nil, userErr
	}

	sess, sessErr := sc.getSession(user.ID)
	if sess == nil && sessErr == nil {
		newSess, newSessErr := sc.createNewSession(
			user.ID,
			ip,
			user.Phone,
			userAgent,
		)
		if newSessErr != nil {
			return nil, newSessErr
		}

		return newSess, nil
	} else if sessErr != nil {
		return nil, sessErr
	}

	return sess, nil
}

func (sc *sessionClient) createNewSession(
	userID uuid.UUID,
	ip,
	phone,
	userAgent string,
) (*model.Session, error) {
	sessParams := sqlStore.CreateSessionParams{
		ID:        userID,
		Ip:        ip,
		UserAgent: userAgent,
		Phone:     phone,
	}
	newSession, newSessErr := sc.store.CreateSession(
		context.Background(),
		sessParams,
	)
	if newSessErr != nil {
		sc.log.WithFields(logrus.Fields{
			"user_id":    userID,
			"user_agent": userAgent,
			"error":      newSessErr,
		}).Errorf("create new session")
		return nil, newSessErr
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

	user, userErr := userService.User.GetUserByPhone(phone)
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

func (sc *sessionClient) getSession(userID uuid.UUID) (*model.Session, error) {
	foundSess, sessErr := sc.store.GetSession(context.Background(), userID)
	if sessErr == sql.ErrNoRows {
		return nil, nil
	} else if sessErr != nil {
		sc.log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   sessErr,
		}).Errorf("get user session")
		return nil, sessErr
	}

	claims, err := jwt.NewPayload(
		foundSess.ID.String(),
		foundSess.Ip,
		foundSess.Phone,
		sc.config.Expires,
	)
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

	user, userErr := userService.User.GetUserByPhone(foundSess.Phone)
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

func (sc *sessionClient) getRelevantCourierData(
	userID uuid.UUID,
) (bool, model.CourierStatus, error) {

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

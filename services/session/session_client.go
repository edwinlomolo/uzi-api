package session

import (
	"context"
	"database/sql"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/jwt"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var sessionService Session

type sessionClient struct {
	jwtClient jwt.Jwt
	store     *sqlStore.Queries
	logger    *logrus.Logger
	config    config.Jwt
}

func GetSessionService() Session {
	return sessionService
}

func NewSessionService(store *sqlStore.Queries, logger *logrus.Logger, jwtConfig config.Jwt) Session {
	sessionService = &sessionClient{jwt.NewJwtClient(logger, jwtConfig), store, logger, jwtConfig}
	logger.Infoln("Session service...OK")
	return sessionService
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
		err := model.UziErr{Err: newSessErr.Error(), Message: "createnewsession", Code: 500}
		sc.logger.Errorf(err.Error())
		return nil, err
	}

	claims, err := jwt.NewPayload(userID.String(), phone, sc.config.Expires)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, isUserOnboardingErr := sc.store.IsUserOnboarding(context.Background(), userID)
	if isUserOnboardingErr != nil {
		onboardErr := model.UziErr{Err: isUserOnboardingErr.Error(), Message: "isuseronboarding", Code: 500}
		sc.logger.Errorf(onboardErr.Error())
		return nil, onboardErr
	}

	isCourier, courierStatus, courierErr := sc.getRelevantCourierData(userID)
	if courierErr != nil {
		return nil, courierErr
	}

	return &model.Session{
		ID:            newSession.ID,
		IP:            newSession.Ip,
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
		err := model.UziErr{Err: sessErr.Error(), Message: "getsession", Code: 500}
		sc.logger.Errorf(err.Error())
		return nil, err
	}

	claims, err := jwt.NewPayload(foundSess.ID.String(), foundSess.Phone, sc.config.Expires)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := sc.jwtClient.Sign([]byte(sc.config.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, isUserOnboardingErr := sc.store.IsUserOnboarding(context.Background(), foundSess.ID)
	if isUserOnboardingErr != nil {
		onboardErr := model.UziErr{Err: isUserOnboardingErr.Error(), Message: "isuseronboarding", Code: 500}
		sc.logger.Errorf(onboardErr.Error())
		return nil, onboardErr
	}

	isCourier, courierStatus, courierErr := sc.getRelevantCourierData(foundSess.ID)
	if courierErr != nil {
		return nil, courierErr
	}

	return &model.Session{
		ID:            foundSess.ID,
		IP:            foundSess.Ip,
		Phone:         foundSess.Phone,
		UserAgent:     foundSess.UserAgent,
		Token:         sessionJwt,
		CourierStatus: &courierStatus,
		Onboarding:    isUserOnboarding,
		IsCourier:     isCourier,
	}, nil

}

func (sc *sessionClient) getRelevantCourierData(userID uuid.UUID) (bool, model.CourierStatus, error) {

	courierStatus, courierStatusErr := courier.GetCourierService().GetCourierStatus(userID)
	if courierStatusErr != nil {
		return false, courierStatus, courierStatusErr
	}

	isCourier, isCourierErr := courier.GetCourierService().IsCourier(userID)
	if isCourierErr != nil {
		return false, courierStatus, isCourierErr
	}

	return isCourier, courierStatus, nil
}

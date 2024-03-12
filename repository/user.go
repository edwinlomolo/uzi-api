package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	sqlStore "github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	userNotFound = errors.New("user not found")
	noEmptyName  = errors.New("name can't be empty")
)

type SigninInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Courier   bool   `json:"courier"`
}
type UserRepository struct {
	jwt   internal.JwtService
	store *sqlc.Queries
	log   *logrus.Logger
}

func (u *UserRepository) Init() {
	u.jwt = internal.NewJwtClient()
	u.store = sqlStore.GetDb()
	u.log = internal.GetLogger()
}

func (u *UserRepository) FindOrCreate(user SigninInput) (*model.User, error) {
	foundUser, foundUserErr := u.getUser(user.Phone)
	if foundUser == nil && foundUserErr == nil {
		return u.createUser(user)
	} else if foundUserErr != nil {
		return nil, foundUserErr
	}

	return foundUser, nil
}

func (u *UserRepository) FindOrCreateCourier(userID uuid.UUID) (*model.Courier, error) {
	courier, err := u.getCourierByUserID(userID)
	if err == nil && courier == nil {
		newCourier, newErr := u.store.CreateCourier(
			context.Background(),
			uuid.NullUUID{
				UUID:  userID,
				Valid: true,
			},
		)
		if newErr != nil {
			u.log.WithFields(logrus.Fields{
				"courier_user_id": userID,
				"error":           newErr,
			}).Errorf("find/create courier")
			return nil, newErr
		}

		return &model.Courier{
			ID: newCourier.ID,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &model.Courier{
		ID: courier.ID,
	}, nil
}

func (u *UserRepository) getCourierByUserID(userID uuid.UUID) (*model.Courier, error) {
	foundCourier, err := u.store.GetCourierByUserID(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		u.log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Errorf("get courier by id")
		return nil, err
	}

	return &model.Courier{
		ID: foundCourier.ID,
	}, nil
}

func (u *UserRepository) createUser(user SigninInput) (*model.User, error) {
	createArgs := sqlc.CreateUserParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	}
	newUser, newUserErr := u.store.CreateUser(context.Background(), createArgs)
	if newUserErr != nil {
		u.log.WithFields(logrus.Fields{
			"error": newUserErr,
			"user":  user,
		}).Errorf("create new user")
		return nil, newUserErr
	}

	if user.Courier {
		if _, courierErr := u.FindOrCreateCourier(newUser.ID); courierErr != nil {
			return nil, courierErr
		}
	}

	return &model.User{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Phone:     newUser.Phone,
	}, nil
}

func (u *UserRepository) getUser(phone string) (*model.User, error) {
	foundUser, getErr := u.store.FindByPhone(context.Background(), phone)
	if getErr == sql.ErrNoRows {
		return nil, nil
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"error": getErr,
			"phone": phone,
		}).Errorf("get user by phone")
		return nil, getErr
	}

	return &model.User{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Phone:     foundUser.Phone,
	}, nil
}

func (u *UserRepository) GetUserByPhone(phone string) (*model.User, error) {
	return u.getUser(phone)
}

func (u *UserRepository) findUserByID(id uuid.UUID) (*model.User, error) {
	foundUser, getErr := u.store.FindUserByID(context.Background(), id)
	if getErr == sql.ErrNoRows {
		u.log.WithFields(logrus.Fields{
			"user_id": id,
			"error":   userNotFound.Error(),
		}).Errorf(userNotFound.Error())
		return nil, userNotFound
	} else if getErr != nil {
		u.log.WithFields(logrus.Fields{
			"error":   getErr,
			"user_id": id,
		}).Errorf("get user by id")
		return nil, getErr
	}

	return &model.User{
		ID:        foundUser.ID,
		FirstName: foundUser.FirstName,
		LastName:  foundUser.LastName,
		Phone:     foundUser.Phone,
	}, nil
}

func (u *UserRepository) FindUserByID(id uuid.UUID) (*model.User, error) {
	return u.findUserByID(id)
}

func (u *UserRepository) OnboardUser(user SigninInput) (*model.User, error) {
	if len(user.FirstName) == 0 || len(user.LastName) == 0 {
		u.log.WithError(noEmptyName).Errorf("invalid names")
		return nil, noEmptyName
	}

	updateArgs := sqlc.UpdateUserNameParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
	}
	newUser, onboardErr := u.store.UpdateUserName(context.Background(), updateArgs)
	if onboardErr != nil {
		u.log.WithFields(logrus.Fields{
			"args":  updateArgs,
			"error": onboardErr,
		}).Errorf("update user name")
		return nil, onboardErr
	}

	statusArgs := sqlc.SetOnboardingStatusParams{
		Phone:      user.Phone,
		Onboarding: false,
	}
	if _, err := u.store.SetOnboardingStatus(context.Background(), statusArgs); err != nil {
		u.log.WithFields(logrus.Fields{
			"error": err,
			"phone": user.Phone,
		}).Errorf("set user onboarding status")
		return nil, err
	}

	return &model.User{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Phone:     newUser.Phone,
	}, nil
}

func (u *UserRepository) SignIn(
	signin SigninInput, ip, userAgent string,
) (*model.Session, error) {
	return u.findOrCreateSession(signin, ip, userAgent)
}

func (u *UserRepository) findOrCreateSession(
	signin SigninInput,
	ip, userAgent string,
) (*model.Session, error) {
	user, userErr := u.FindOrCreate(signin)
	if userErr != nil {
		return nil, userErr
	}

	sess, sessErr := u.getSession(user.ID)
	if sess == nil && sessErr == nil {
		newSess, newSessErr := u.createNewSession(
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

func (u *UserRepository) getSession(userID uuid.UUID) (*model.Session, error) {
	foundSess, sessErr := u.store.GetSession(context.Background(), userID)
	if sessErr == sql.ErrNoRows {
		return nil, nil
	} else if sessErr != nil {
		u.log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   sessErr,
		}).Errorf("get user session")
		return nil, sessErr
	}

	claims, err := internal.NewPayload(
		foundSess.ID.String(),
		foundSess.Ip,
		foundSess.Phone,
		config.Config.Jwt.Expires,
	)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := u.jwt.Sign([]byte(config.Config.Jwt.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, _ := u.store.IsUserOnboarding(context.Background(), foundSess.ID)

	isCourier, courierStatus, courierErr := u.getRelevantCourierData(foundSess.ID)
	if courierErr != nil {
		return nil, courierErr
	}

	user, userErr := u.GetUserByPhone(foundSess.Phone)
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

func (u *UserRepository) createNewSession(
	userID uuid.UUID,
	ip, phone, userAgent string,
) (*model.Session, error) {
	sessParams := sqlc.CreateSessionParams{
		ID:        userID,
		Ip:        ip,
		UserAgent: userAgent,
		Phone:     phone,
	}
	newSession, newSessErr := u.store.CreateSession(
		context.Background(),
		sessParams,
	)
	if newSessErr != nil {
		u.log.WithFields(logrus.Fields{
			"user_id":    userID,
			"user_agent": userAgent,
			"error":      newSessErr,
		}).Errorf("create new session")
		return nil, newSessErr
	}

	claims, err := internal.NewPayload(userID.String(), ip, phone, config.Config.Jwt.Expires)
	if err != nil {
		return nil, err
	}

	sessionJwt, signJwtErr := u.jwt.Sign([]byte(config.Config.Jwt.Secret), claims)
	if signJwtErr != nil {
		return nil, signJwtErr
	}

	isUserOnboarding, _ := u.store.IsUserOnboarding(context.Background(), userID)

	isCourier, courierStatus, courierErr := u.getRelevantCourierData(userID)
	if courierErr != nil {
		return nil, courierErr
	}

	user, userErr := u.GetUserByPhone(phone)
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

func (u *UserRepository) getCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := u.store.GetCourierStatus(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return model.CourierStatusOnboarding, nil
	} else if err != nil {
		u.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           err,
		}).Errorf("get courier status")
		return model.CourierStatusOffline, err
	}

	return model.CourierStatus(status), nil
}

func (u *UserRepository) isCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := u.store.IsCourier(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		u.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           err,
		}).Errorf("is courier check")
		return false, err
	}

	return isCourier.Bool, nil
}

func (u *UserRepository) getRelevantCourierData(userID uuid.UUID) (bool, model.CourierStatus, error) {

	courierStatus, courierStatusErr := u.getCourierStatus(userID)
	if courierStatusErr != nil {
		return false, courierStatus, courierStatusErr
	}

	isCourier, isCourierErr := u.isCourier(userID)
	if isCourierErr != nil {
		return false, courierStatus, isCourierErr
	}

	return isCourier, courierStatus, nil
}

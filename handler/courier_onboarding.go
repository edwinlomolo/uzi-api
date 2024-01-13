package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/services"
)

func CourierOnboarding() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyReq model.SigninInput
		logger := logger.GetLogger()
		userService := services.GetUserService()
		sessionService := services.GetSessionService()
		ip := GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			uziErr := model.UziErr{Err: errors.New("CourierOnboardingBodyErr").Error(), Message: "ioread", Code: http.StatusBadRequest}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		if err := json.Unmarshal(body, &bodyReq); err != nil {
			uziErr := model.UziErr{Err: errors.New("UmarshalCourierBody").Error(), Message: "unmarshal", Code: http.StatusBadRequest}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		updatedUser, onboardErr := userService.OnboardUser(bodyReq)
		if onboardErr != nil {
			http.Error(w, onboardErr.Error(), http.StatusInternalServerError)
			return
		}

		session, sessionErr := sessionService.SignIn(*updatedUser, ip)
		if sessionErr != nil {
			http.Error(w, sessionErr.Error(), http.StatusInternalServerError)
			return
		}

		res, resErr := json.Marshal(session)
		if resErr != nil {
			uziErr := model.UziErr{Err: errors.New("MarshalOnboardRes").Error(), Message: "unmarshal", Code: http.StatusInternalServerError}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}

package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/model"
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
			logger.Errorf("%s-%v", "courierOnboardingBodyErr", bodyErr.Error())
			http.Error(w, bodyErr.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &bodyReq); err != nil {
			logger.Errorf("%s-%v", "unmarshalCourierBodyErr", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedUser, onboardErr := userService.OnboardUser(bodyReq)
		if onboardErr != nil {
			http.Error(w, onboardErr.ErrorString(), onboardErr.Code)
			return
		}

		session, sessionErr := sessionService.SignIn(updatedUser.ID, ip, updatedUser.Phone)
		if sessionErr != nil {
			http.Error(w, sessionErr.ErrorString(), sessionErr.Code)
			return
		}

		res, resErr := json.Marshal(session)
		if resErr != nil {
			logger.Errorf("%s-%v", "marshalOnboardResErr", resErr.Error())
			http.Error(w, resErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}

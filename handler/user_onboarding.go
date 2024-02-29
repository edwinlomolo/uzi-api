package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/session"
	"github.com/edwinlomolo/uzi-api/user"
)

func UserOnboarding() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyReq user.SigninInput
		logger := logger.Logger
		userService := user.User
		sessionService := session.Session
		ip := GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			logger.WithError(bodyErr).Errorf("reading req body")
			http.Error(w, bodyErr.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &bodyReq); err != nil {
			logger.WithError(err).Errorf("unmarshal body req")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, onboardErr := userService.OnboardUser(bodyReq); onboardErr != nil {
			http.Error(w, onboardErr.Error(), http.StatusInternalServerError)
			return
		}

		session, sessionErr := sessionService.SignIn(bodyReq, ip, r.UserAgent())
		if sessionErr != nil {
			http.Error(w, sessionErr.Error(), http.StatusInternalServerError)
			return
		}

		res, resErr := json.Marshal(session)
		if resErr != nil {
			logger.WithError(resErr).Errorf("marshal session res")
			http.Error(w, resErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}

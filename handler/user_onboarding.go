package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/util"
	"github.com/edwinlomolo/uzi-api/services/session"
	"github.com/edwinlomolo/uzi-api/services/user"
)

func UserOnboarding() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyReq user.SigninInput
		logger := logger.Logger
		userService := user.User
		sessionService := session.Session
		ip := util.GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			uziErr := fmt.Errorf("%s:%v", "reading body", bodyErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &bodyReq); err != nil {
			uziErr := fmt.Errorf("%s:%v", "unmarshalbody", err)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
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
			uziErr := fmt.Errorf("%s:%v", "session marshal", resErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}

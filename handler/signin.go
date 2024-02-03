package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/util"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/3dw1nM0535/uzi-api/services/session"
	"github.com/3dw1nM0535/uzi-api/services/user"
)

func Signin() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput user.SigninInput
		logger := logger.Logger
		userService := user.User
		sessionService := session.Session
		courierService := courier.Courier
		userIp := util.GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			uziErr := fmt.Errorf("%s:%v", "reading body", bodyErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			uziErr := fmt.Errorf("%s:%v", "unmarshal body", marshalErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		findUser, findUserErr := userService.FindOrCreate(loginInput)
		if findUserErr != nil {
			http.Error(w, findUserErr.Error(), http.StatusInternalServerError)
			return
		}

		if loginInput.Courier {
			_, courierErr := courierService.FindOrCreate(findUser.ID)
			if courierErr != nil {
				http.Error(w, courierErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		findSession, findSessionErr := sessionService.SignIn(*findUser, userIp, r.UserAgent())
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			uziErr := fmt.Errorf("%s:%v", "marshal session", jsonErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

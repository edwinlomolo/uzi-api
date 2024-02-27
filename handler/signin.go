package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/services/session"
	"github.com/edwinlomolo/uzi-api/services/user"
)

func Signin() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput user.SigninInput
		logger := logger.Logger
		sessionService := session.Session
		userIp := GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			uziErr := fmt.Errorf("%s:%v", "reading body", bodyErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			uziErr := fmt.Errorf("%s:%v", "unmarshal login req body", marshalErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		findSession, findSessionErr := sessionService.SignIn(loginInput, userIp, r.UserAgent())
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			uziErr := fmt.Errorf("%s:%v", "marshal session res", jsonErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

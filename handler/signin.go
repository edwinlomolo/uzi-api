package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services"
)

func Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput model.SigninInput
		logger := logger.GetLogger()
		userService := services.GetUserService()
		sessionService := services.GetSessionService()
		userIp := GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			logger.Errorf("%s-%v", "ReadingSigninRequestBodyErr", bodyErr.Error())
			http.Error(w, bodyErr.Error(), http.StatusBadRequest)
			return
		}
		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			logger.Errorf("%s-%v", "SigninRequestBodyMarshalErr", marshalErr.Error())
			http.Error(w, marshalErr.Error(), http.StatusBadRequest)
			return
		}

		findUser, findUserErr := userService.FindOrCreate(loginInput)
		if findUserErr != nil {
			http.Error(w, findUserErr.Error(), http.StatusInternalServerError)
			return
		}

		findSession, findSessionErr := sessionService.SignIn(findUser.ID, userIp, findUser.Phone)
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			logger.Errorf("%s-%v", "SigninResponseMarshalErr", jsonErr.Error())
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

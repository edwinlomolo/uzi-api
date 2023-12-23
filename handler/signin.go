package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services"
	"github.com/sirupsen/logrus"
)

func Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput model.SigninInput
		ctx := r.Context()
		logger := ctx.Value("logger").(*logrus.Logger)
		userService := ctx.Value("userService").(services.User)
		sessionService := ctx.Value("sessionService").(services.Session)
		userIp := GetIp(r)

		if r.Method != http.MethodPost {
			http.Error(w, errors.New("Only POST method supported").Error(), http.StatusMethodNotAllowed)
			return
		}

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			logger.Errorf("%s-%v", "ReadRequestBodyErr", bodyErr.Error())
			http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
			return
		}
		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			logger.Errorf("%s-%v", "RequestBodyMarshalErr", marshalErr.Error())
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		findUser, findUserErr := userService.FindOrCreate(loginInput)
		if findUserErr != nil {
			http.Error(w, findUserErr.Error(), http.StatusInternalServerError)
			return
		}

		if config.IsDev() {
			userIp = "127.0.0.1"
		}

		findSession, findSessionErr := sessionService.FindOrCreate(findUser.ID, userIp)
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			logger.Errorf("%s-%v", "JsonResponseMarshalErr", jsonErr.Error())
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
		return
	})
}

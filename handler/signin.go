package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/services/courier"
	"github.com/3dw1nM0535/uzi-api/services/session"
	"github.com/3dw1nM0535/uzi-api/services/user"
)

func Signin() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput model.SigninInput
		logger := logger.GetLogger()
		userService := user.GetUserService()
		sessionService := session.GetSessionService()
		courierService := courier.GetCourierService()
		userIp := GetIp(r)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			uziErr := model.UziErr{Err: errors.New("ReadingSigninRequestBody").Error(), Message: "ioread", Code: http.StatusInternalServerError}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			uziErr := model.UziErr{Err: errors.New("SigninRequestBodyMarshal").Error(), Message: "marshal", Code: http.StatusInternalServerError}
			logger.Errorf(uziErr.Error())
			http.Error(w, marshalErr.Error(), uziErr.Code)
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

		findSession, findSessionErr := sessionService.SignIn(*findUser, userIp)
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			uziErr := model.UziErr{Err: errors.New("SigninResponseMarshal").Error(), Message: "marshal", Code: http.StatusInternalServerError}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

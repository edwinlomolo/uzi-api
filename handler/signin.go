package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/edwinlomolo/uzi-api/internal"
	repo "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/services"
)

func Signin() http.HandlerFunc {
	userService := services.GetUserService()
	log := internal.GetLogger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loginInput repo.SigninInput
		userIp := r.Context().Value("ip").(string)

		body, bodyErr := io.ReadAll(r.Body)
		if bodyErr != nil {
			log.WithError(bodyErr).Errorf("handler: reading signin body")
			http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
			return
		}

		if marshalErr := json.Unmarshal(body, &loginInput); marshalErr != nil {
			log.WithError(marshalErr).Errorf("handler: unmarshal signin body")
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		findSession, findSessionErr := userService.SignIn(loginInput, userIp, r.UserAgent())
		if findSessionErr != nil {
			http.Error(w, findSessionErr.Error(), http.StatusInternalServerError)
			return
		}

		jsonRes, jsonErr := json.Marshal(findSession)
		if jsonErr != nil {
			log.WithError(jsonErr).Errorf("handler: marshal signin res")
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

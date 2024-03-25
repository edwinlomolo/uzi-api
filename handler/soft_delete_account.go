package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/edwinlomolo/uzi-api/controllers"
	"github.com/edwinlomolo/uzi-api/internal"
)

func SoftDeleteAccount() http.HandlerFunc {
	userS := controllers.GetUserController()
	log := internal.GetLogger()
	deleteRequest := struct {
		Phone string `json:"phone"`
	}{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&deleteRequest); err != nil {
			log.WithError(err).Errorf("handler: reading delete request body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := userS.GetUserByPhone(deleteRequest.Phone)
		if err != nil && u == nil {
			http.Error(w, errors.New("not found").Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := userS.SoftDelete(deleteRequest.Phone); err != nil {
			log.WithError(err).Error("handler: soft delete user")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}

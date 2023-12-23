package handler

import (
	"encoding/json"
	"errors"
	"net/http"
)

func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errors.New("Only POST method supported").Error(), http.StatusMethodNotAllowed)
			return
		}

		res, _ := json.Marshal("message: ok")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
		return
	})
}

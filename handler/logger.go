package handler

import (
	"net/http"

	"github.com/edwinlomolo/uzi-api/logger"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetIp(r)
		log := logger.Logger
		// Some info on what is happening with request(s)
		log.Infof("%s-%s-%s-%s", ip, r.Method, r.URL, r.Proto)
		h.ServeHTTP(w, r)
	})
}

package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := GetIp(r)
		log := ctx.Value("logger").(*logrus.Logger)
		// Some info on what is happening with request(s)
		log.Infof("%s-%s-%s-%s", ip, r.Method, r.URL, r.Proto)
		h.ServeHTTP(w, r)
	})
}

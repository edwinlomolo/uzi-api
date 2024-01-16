package handler

import (
	"net/http"

	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/util"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := util.GetIp(r)
		log := logger.GetLogger()
		// Some info on what is happening with request(s)
		log.Infof("%s-%s-%s-%s", ip, r.Method, r.URL, r.Proto)
		h.ServeHTTP(w, r)
	})
}

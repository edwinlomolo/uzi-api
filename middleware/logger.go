package middleware

import (
	"net/http"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ip := r.Context().Value("ip").(string)
			// Some info on what is happening with request(s)
			log.Infof("%s-%s-%s-%s", ip, r.Method, r.URL, r.Proto)
			h.ServeHTTP(w, r)
		})
}

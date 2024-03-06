package middleware

import (
	"context"
	"net/http"
)

func GetIp(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			clientIp := r.Header.Get("X-FORWARDED-FOR")
			if clientIp == "" {
				clientIp = r.RemoteAddr
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "ip", clientIp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
}

package middleware

import (
	"context"
	"net/http"
)

func EnrichRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = context.WithValue(ctx, "ip", r.RemoteAddr)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
}

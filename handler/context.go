package handler

import (
	"context"
	"net/http"
)

func Context(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := GetIp(r)

		ctx = context.WithValue(ctx, "ip", ip)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

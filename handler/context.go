package handler

import (
	"context"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/services"
)

func Context(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := GetIp(r)
		logger := logger.GetLogger()
		ipinfoService := services.GetIpinfoServices()

		ctx = context.WithValue(ctx, "ipinfoService", ipinfoService)
		ctx = context.WithValue(ctx, "logger", logger)
		ctx = context.WithValue(ctx, "ip", ip)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

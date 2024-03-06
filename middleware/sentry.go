package middleware

import (
	"net/http"

	"github.com/edwinlomolo/uzi-api/config"
	sentryHttp "github.com/getsentry/sentry-go/http"
)

func Sentry(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if isStaging() || isProd() {
				sentryHttp.New(sentryHttp.Options{}).Handle(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
}

func isStaging() bool {
	return config.Config.Server.Env == "staging"
}

func isProd() bool {
	return config.Config.Server.Env == "production"
}

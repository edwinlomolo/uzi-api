package middleware

import (
	"net/http"

	sentryHttp "github.com/getsentry/sentry-go/http"
)

func Sentry(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			sentryHttp.New(sentryHttp.Options{}).Handle(next).ServeHTTP(w, r)
		})
}

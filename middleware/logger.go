package middleware

import (
	"encoding/json"
	"net/http"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqJson := struct {
				OperationName string `json:"operationName"`
			}{}
			path := r.URL.Path
			err := json.NewDecoder(r.Body).Decode(&reqJson)
			if err != nil && path == "/api/graphql" {
				log.Warnln("middleware: unmarshal http req body for GraphQL API logging")
			}
			ip := r.Context().Value("ip").(string)
			// Some info on what is happening with request(s)
			if path == "/api/graphql" {
				// GraphQL API
				log.Infof("%s-%s-%s-%s-%s-%s", r.UserAgent(), ip, r.Method, reqJson.OperationName, r.URL, r.Proto)
			} else {
				// Rest API
				log.Infof("%s-%s-%s-%s-%s", r.UserAgent(), ip, r.Method, r.URL, r.Proto)
			}
			h.ServeHTTP(w, r)
		})
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/edwinlomolo/uzi-api/internal"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidHeader = errors.New("invalid token header")
	log              = internal.GetLogger()
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var (
				userID,
				ip string
			)
			ctx := r.Context()

			jwtToken, jwtErr := validateAuthorizationHeader(r)
			if jwtErr != nil {
				http.Error(w, jwtErr.Error(), http.StatusUnauthorized)
				return
			}

			if claims, ok := jwtToken.Claims.(*internal.Payload); ok && jwtToken.Valid {
				userID = claims.ID
				ip = claims.IP
			}

			ctx = context.WithValue(ctx, "userID", userID)
			ctx = context.WithValue(ctx, "ip", ip)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
}

func validateAuthorizationHeader(
	r *http.Request,
) (*jsonwebtoken.Token, error) {
	var tokenString string
	jwtService := internal.NewJwtClient()

	authorizationHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(authorizationHeader) != 2 || authorizationHeader[0] != "Bearer" {
		log.Warnln(ErrInvalidHeader.Error())
		return nil, ErrInvalidHeader
	}

	tokenString = authorizationHeader[1]
	jwtToken, jwtErr := jwtService.Validate(tokenString)
	if jwtErr != nil {
		return nil, jwtErr
	}
	return jwtToken, nil
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/jwt"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	jsonwebtoken "github.com/golang-jwt/jwt/v5"
)

func Auth(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			userID string
		)
		ctx := r.Context()

		jwtToken, jwtErr := validateAuthorizationHeader(r)
		if jwtErr != nil {
			http.Error(w, jwtErr.Error(), http.StatusUnauthorized)
			return
		}

		if claims, ok := jwtToken.Claims.(*jwt.Payload); ok && jwtToken.Valid {
			userID = claims.ID
		}

		ctx = context.WithValue(ctx, "userID", userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateAuthorizationHeader(r *http.Request) (*jsonwebtoken.Token, error) {
	var tokenString string
	logger := logger.GetLogger()
	jwtService := jwt.GetJwtService()

	authorizationHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(authorizationHeader) != 2 || authorizationHeader[0] != "Bearer" {
		uziErr := model.UziErr{Err: errors.New("invalid token header").Error(), Message: "invalidtoken", Code: http.StatusUnauthorized}
		logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	tokenString = authorizationHeader[1]
	jwtToken, jwtErr := jwtService.Validate(tokenString)
	if jwtErr != nil {
		return nil, jwtErr
	}
	return jwtToken, nil
}

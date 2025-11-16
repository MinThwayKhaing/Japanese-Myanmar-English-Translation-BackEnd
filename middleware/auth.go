package middleware

import (
	"context"
	"net/http"
	"strings"

	"USDT_BackEnd/config"

	"github.com/dgrijalva/jwt-go"
)

type key int

const UserKey key = 0

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

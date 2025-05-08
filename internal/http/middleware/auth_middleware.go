package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/superlogarifm/goCalc-v3/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	AuthService *auth.AuthService
}

func NewAuthMiddleware(authService *auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{AuthService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format (expected Bearer token)", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		userID, _, err := m.AuthService.ValidateToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			errMsg := "Invalid token"
			if errors.Is(err, auth.ErrTokenExpired) {
				errMsg = "Token expired"
			}
			http.Error(w, errMsg, status)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint)
	return userID, ok
}

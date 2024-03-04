package middleware

import (
	"context"
	"gophermart/internal/app/usecase"
	"net/http"
)

type UserService interface {
	GetUserIDFromToken(token string) (int, error)
}

type SecurityMiddleware struct {
	UserService
}

func NewSecurityMiddleware(s UserService) *SecurityMiddleware {
	return &SecurityMiddleware{s}
}

func (m *SecurityMiddleware) SecurityMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(usecase.UserID))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := m.GetUserIDFromToken(cookie.Value)
		if err != nil || userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), usecase.UserID, userID)))
	})
}

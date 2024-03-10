package middleware

import (
	"context"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase"
)

type UserService interface {
	GetUserIDFromToken(token string) (int, error)
	ExistsUser(ctx context.Context, userID int) bool
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
		if !m.ExistsUser(r.Context(), userID) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), usecase.UserID, userID)))
	})
}

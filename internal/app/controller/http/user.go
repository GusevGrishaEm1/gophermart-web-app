package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"
)

type UserService interface {
	RegisterUser(context.Context, *RegisterRequest) (string, error)
	LoginUser(context.Context, *LoginRequest) (string, error)
	GetUserIDFromContext(ctx context.Context) (int, error)
}

type UserHandler struct {
	config *config.Config
	UserService
}

func NewUserHandler(config *config.Config, service UserService) *UserHandler {
	return &UserHandler{config, service}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (userHandler *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var dto RegisterRequest
	err = json.Unmarshal(buf, &dto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := userHandler.RegisterUser(r.Context(), &dto)
	shouldReturn := userHandler.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	cookie := &http.Cookie{
		Name:  string("USER_ID"),
		Value: token,
	}
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (userHandler *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var dto LoginRequest
	err = json.Unmarshal(buf, &dto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := userHandler.LoginUser(r.Context(), &dto)
	shouldReturn := userHandler.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	cookie := &http.Cookie{
		Name:  string("USER_ID"),
		Value: token,
	}
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (*UserHandler) validateErrorAfter(err error, w http.ResponseWriter) bool {
	if err != nil {
		customErr := &customerr.CustomError{}
		if errors.As(err, &customErr) {
			w.WriteHeader(customErr.HTTPStatus)
			return true
		}
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

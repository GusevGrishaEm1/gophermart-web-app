package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/go-playground/validator/v10"
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
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (userHandler *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		sendClientErr(err, w)
		return
	}
	var dto RegisterRequest
	err = json.Unmarshal(buf, &dto)
	if err != nil {
		sendClientErr(err, w)
		return
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(dto)
	if err != nil {
		sendClientErr(err, w)
		return
	}
	token, err := userHandler.RegisterUser(r.Context(), &dto)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	sendOKWithCookie(token, w)
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (userHandler *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		sendClientErr(err, w)
		return
	}
	var dto LoginRequest
	err = json.Unmarshal(buf, &dto)
	if err != nil {
		sendClientErr(err, w)
		return
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(dto)
	if err != nil {
		sendClientErr(err, w)
		return
	}
	token, err := userHandler.LoginUser(r.Context(), &dto)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	sendOKWithCookie(token, w)
}

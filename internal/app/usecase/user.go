package usecase

import (
	"context"
	"errors"
	nethttp "net/http"
	"time"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserInfo string

const UserID UserInfo = "USER_ID"

type UserService struct {
	c *config.Config
	repository.UserRepository
}

func NewUserService(c *config.Config, r repository.UserRepository) *UserService {
	return &UserService{c, r}
}

func (s *UserService) RegisterUser(ctx context.Context, dto *http.RegisterRequest) (string, error) {
	hash, err := hashPassword(dto.Password)
	if err != nil {
		return "", err
	}
	user := &entity.User{
		Login:    dto.Login,
		Password: hash,
	}
	id, err := s.Save(ctx, user)
	if err != nil {
		return "", err
	}
	jwt, err := buildJWTString(id)
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *UserService) LoginUser(ctx context.Context, dto *http.LoginRequest) (string, error) {
	user, err := s.FindByLogin(ctx, dto.Login)
	if err != nil {
		return "", err
	}
	err = checkPasswordHash(dto.Password, user.Password)
	if err != nil {
		return "", customerr.NewError(err, nethttp.StatusUnauthorized)
	}
	jwt, err := buildJWTString(user.ID)
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func checkPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func buildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			UserID: userID,
		},
	)
	tokenString, err := token.SignedString([]byte("secretkey"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *UserService) GetUserIDFromToken(token string) (int, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("secretkey"), nil
	})
	if err != nil {
		return 0, err
	}
	return int(claims.UserID), nil
}

func (s *UserService) GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserID).(int)
	if !ok {
		return 0, errors.New("user_id is nil")
	}
	return userID, nil
}

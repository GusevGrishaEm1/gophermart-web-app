package repository

import (
	"context"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository/postgres"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) (int, error)
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

func NewUserRepository(ctx context.Context, config *config.Config) (UserRepository, error) {
	return postgres.NewUserRepository(ctx, config, config.Pool)
}

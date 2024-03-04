package repository

import (
	"context"
	"gophermart/cmd/gophermart/internal/app/config"
	"gophermart/cmd/gophermart/internal/app/entity"
	"gophermart/cmd/gophermart/internal/app/infrastructure/repository/postgres"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

func NewUserRepository(ctx context.Context, config *config.Config) (UserRepository, error) {
	return postgres.NewUserRepository(ctx, config)
}

package repository

import (
	"context"
	"gophermart/internal/app/config"
	"gophermart/internal/app/entity"
	"gophermart/internal/app/infrastructure/repository/postgres"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

func NewUserRepository(ctx context.Context, config *config.Config) (UserRepository, error) {
	return postgres.NewUserRepository(ctx, config)
}

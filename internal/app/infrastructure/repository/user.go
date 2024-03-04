package repository

import (
	"context"
	"gophermart/internal/app/config"
	"gophermart/internal/app/entity"
	"gophermart/internal/app/infrastructure/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

func NewUserRepository(ctx context.Context, config *config.Config, pool *pgxpool.Pool) (UserRepository, error) {
	return postgres.NewUserRepository(ctx, config, pool)
}

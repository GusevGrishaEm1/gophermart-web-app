package repository

import (
	"context"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceOperationRepository interface {
	SaveOrder(ctx context.Context, balanceOperation *entity.BalanceOperation) error
	FindOrdersByUser(ctx context.Context, userID int) ([]*entity.BalanceOperation, error)
	GetBalanceByUser(ctx context.Context, userID int) (int, int, error)
	FindWithdrawsByUser(ctx context.Context, userID int) ([]*entity.BalanceOperation, error)
	SaveWithdraw(ctx context.Context, balanceOperation *entity.BalanceOperation) error
	FindOrdersToProcess(ctx context.Context) ([]*entity.BalanceOperation, error)
	UpdateOrders(ctx context.Context, balanceOperation []*entity.BalanceOperation) error
}

func NewBalanceOperationRepository(ctx context.Context, config *config.Config, pool *pgxpool.Pool) (BalanceOperationRepository, error) {
	return postgres.NewBalanceOperationRepository(ctx, config, pool)
}

package repository

import (
	"context"
	"gophermart/internal/app/config"
	"gophermart/internal/app/entity"
	"gophermart/internal/app/infrastructure/repository/postgres"
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

func NewBalanceOperationRepository(ctx context.Context, config *config.Config) (BalanceOperationRepository, error) {
	return postgres.NewBalanceOperationRepository(ctx, config)
}

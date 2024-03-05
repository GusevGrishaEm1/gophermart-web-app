package postgres

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceOperationRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceOperationRepository(ctx context.Context, config *config.Config, pool *pgxpool.Pool) (*BalanceOperationRepository, error) {
	return &BalanceOperationRepository{pool: pool}, nil
}

func (r *BalanceOperationRepository) SaveOrder(ctx context.Context, balanceOperation *entity.BalanceOperation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	shouldReturn, err := r.saveWithTx(ctx, tx, balanceOperation)
	if shouldReturn {
		return err
	}
	tx.Commit(ctx)
	return nil
}

func (r *BalanceOperationRepository) FindOrdersByUser(ctx context.Context, userID int) ([]*entity.BalanceOperation, error) {
	query := `
		select "id", "order", "status", "sum", "created_at" from "balance_operation" where "user_id" = $1 and "deleted_at" is null 
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, customerr.NewError(err, http.StatusInternalServerError)
	}
	result := make([]*entity.BalanceOperation, 0)
	for rows.Next() {
		balance := &entity.BalanceOperation{}
		var status string
		err = rows.Scan(&balance.ID, &balance.Order, &status, &balance.Sum, &balance.CreatedAt)
		balance.Status = entity.ProcessStatus(status)
		if err != nil {
			return nil, customerr.NewError(err, http.StatusInternalServerError)
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *BalanceOperationRepository) GetBalanceByUser(ctx context.Context, userID int) (int, int, error) {
	query := `
		select 
			coalesce((select sum("sum") from "balance_operation" where "user_id" = $1 and "deleted_at" is null and status = 'PROCESSED'), 0) as "current",
			coalesce((select sum("sum") from "balance_operation" where "user_id" = $1 and "deleted_at" is null and type = 'WITHDRAW' and status = 'PROCESSED'), 0) as "withdrawn"
	`
	row := r.pool.QueryRow(ctx, query, userID)
	var current int
	var withdrawn int
	err := row.Scan(&current, &withdrawn)
	if err != nil {
		return 0, 0, err
	}
	return current, withdrawn, nil
}

func (r *BalanceOperationRepository) FindWithdrawsByUser(ctx context.Context, userID int) ([]*entity.BalanceOperation, error) {
	query := `
		select "order", "sum", "created_at" from "balance_operation" where "user_id" = $1 and "deleted_at" is null and type = 'WITHDRAW' and status = 'PROCESSED'
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, customerr.NewError(err, http.StatusInternalServerError)
	}
	result := make([]*entity.BalanceOperation, 0)
	for rows.Next() {
		balance := &entity.BalanceOperation{}
		var status string
		err = rows.Scan(&balance.ID, &balance.Order, &balance.Sum, &balance.CreatedAt)
		balance.Status = entity.ProcessStatus(status)
		if err != nil {
			return nil, customerr.NewError(err, http.StatusInternalServerError)
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *BalanceOperationRepository) SaveWithdraw(ctx context.Context, balanceOperation *entity.BalanceOperation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	current, _, err := r.GetBalanceByUser(ctx, balanceOperation.ID)
	if err != nil {
		return customerr.NewError(err, http.StatusInternalServerError)
	}
	if balanceOperation.Sum*(-1) > current {
		log.Print("sum")
		log.Print(balanceOperation.Sum * (-1))
		log.Print(current)
		return customerr.NewError(errors.New("current balance < withdraw"), http.StatusPaymentRequired)
	}
	shouldReturn, err := r.saveWithTx(ctx, tx, balanceOperation)
	if shouldReturn {
		return err
	}
	tx.Commit(ctx)
	return nil
}

func (*BalanceOperationRepository) saveWithTx(ctx context.Context, tx pgx.Tx, balanceOperation *entity.BalanceOperation) (bool, error) {
	query := `
		with ins as (
			insert into "balance_operation" ("order", "status", "type", "user_id") values($1, $2, $3, $4) on conflict("order") where "deleted_at" is null do nothing returning id
		) select 
			case when (select ins.id from ins) is null
			then (select "user_id" from "balance_operation" where "order" = $1 and "deleted_at" is null)
			else 0 end as userID
	`
	row := tx.QueryRow(ctx, query, balanceOperation.Order, string(balanceOperation.Status), string(balanceOperation.Type), balanceOperation.UserID)
	var userID int
	err := row.Scan(&userID)
	if err != nil {
		return true, customerr.NewError(err, http.StatusInternalServerError)
	}
	if userID != 0 {
		if userID == balanceOperation.UserID {
			return true, customerr.NewError(errors.New("order is already saved"), http.StatusOK)
		}
		return true, customerr.NewError(errors.New("order is already saved for another user"), http.StatusConflict)
	}
	return false, nil
}

func (r *BalanceOperationRepository) FindOrdersToProcess(ctx context.Context) ([]*entity.BalanceOperation, error) {
	query := `
		with upd as (
			update "balance_operation"
			set status = 'PROCESSING'
			where "deleted_at" is null
			and type = 'ACCRUAL'
			and status = 'NEW' returning "id", "order"
		) select "id", "order" from upd
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, customerr.NewError(err, http.StatusInternalServerError)
	}
	result := make([]*entity.BalanceOperation, 0)
	for rows.Next() {
		balance := &entity.BalanceOperation{}
		err = rows.Scan(&balance.ID, &balance.Order)
		if err != nil {
			return nil, customerr.NewError(err, http.StatusInternalServerError)
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *BalanceOperationRepository) UpdateOrders(ctx context.Context, balanceOperations []*entity.BalanceOperation) error {
	query := `
		update "balance_operation"
		set 
			status = $2,
			sum = sum - $3
		where id = $1
	`
	batch := &pgx.Batch{}
	for _, el := range balanceOperations {
		batch.Queue(query, el.ID, el.Status, el.Sum)
	}
	results := r.pool.SendBatch(ctx, batch)
	_, err := results.Exec()
	if err != nil {
		return err
	}
	return nil
}

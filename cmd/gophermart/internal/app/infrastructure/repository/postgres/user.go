package postgres

import (
	"context"
	"errors"
	"gophermart/cmd/gophermart/internal/app/config"
	"gophermart/cmd/gophermart/internal/app/entity"
	customerr "gophermart/cmd/gophermart/internal/app/error"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, config *config.Config) (*UserRepository, error) {
	pool, err := pgxpool.New(ctx, config.DatabaseURI)
	if err != nil {
		return nil, err
	}
	return &UserRepository{pool: pool}, nil
}

func (r *UserRepository) Save(ctx context.Context, user *entity.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	query := `
		with new_id as (
			insert into "user" ("login", "password") values($1, $2) on conflict("login") where "deleted_at" is null do nothing returning id
		) select case when exists(select * from new_id) then (select id from new_id) else 0 end as id;
	`
	var id int
	err = tx.QueryRow(ctx, query, user.Login, user.Password).Scan(&id)
	if err != nil {
		tx.Rollback(ctx)
		return customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	if id == 0 {
		tx.Rollback(ctx)
		return customerr.NewError(
			errors.New("login conflict"),
			http.StatusConflict,
		)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	return nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	query := `
		select "id", "password" from "user" where "login" = $1 and deleted_at is null
	`
	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Password)
	if err != nil {
		customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	return user, nil
}

package postgres

import (
	"context"
	"errors"
	"net/http"

	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, config *config.Config, pool *pgxpool.Pool) (*UserRepository, error) {
	return &UserRepository{pool: pool}, nil
}

func (r *UserRepository) Save(ctx context.Context, user *entity.User) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, customerr.NewError(
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
		return 0, customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	if id == 0 {
		tx.Rollback(ctx)
		return 0, customerr.NewError(
			errors.New("login conflict"),
			http.StatusConflict,
		)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, customerr.NewError(
			err,
			http.StatusInternalServerError,
		)
	}
	return id, nil
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

func (r *UserRepository) ExistsByID(ctx context.Context, ID int) bool {
	query := `
		select exists(select * from "user" where "id" = $1 and deleted_at is null) as res
	`
	var res bool
	err := r.pool.QueryRow(ctx, query, ID).Scan(&res)
	if err != nil {
		return false
	}
	return res
}

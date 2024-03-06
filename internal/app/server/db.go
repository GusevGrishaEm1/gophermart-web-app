package server

import (
	"context"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		create table if not exists "user" (
			"id" serial not null,
			"login" varchar(255) not null,
			"password" varchar(255) not null,
			"created_at" timestamp default now(),
			"deleted_at" timestamp,
			constraint "user_pk" primary key ("id")
		);
		
		create table if not exists "balance_operation" (
			"id" serial not null,
			"sum" integer default 0,
			"order" varchar(255) not null,
			"status" varchar(255),
			"type" varchar(255) not null,
			"user_id" integer not null,
			"created_at" timestamp default now(),
			"deleted_at" timestamp,
			constraint "balance_operation_pk" primary key ("id")
		);
		
		ALTER TABLE "balance_operation" ADD CONSTRAINT "balance_operation_fk" FOREIGN KEY ("user_id") REFERENCES "user"("id");
		CREATE UNIQUE INDEX "order_idx" ON "balance_operation"("order") where "deleted_at" is null and "type" = 'ACCRUAL';
		CREATE UNIQUE INDEX "login_idx" ON "user"("login") where "deleted_at" is null;
	`
	_, err := pool.Exec(ctx, query)
	return err
}

func InitPool(ctx context.Context, config *config.Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, config.DatabaseURI)
}

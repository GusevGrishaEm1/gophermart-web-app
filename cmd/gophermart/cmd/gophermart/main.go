package main

import (
	"context"
	"gophermart/cmd/gophermart/internal/app/config"
	"gophermart/cmd/gophermart/internal/app/server"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	config := config.New()
	db, err := sql.Open("postgres", config.DatabaseURI)
	if err != nil {
		panic(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		panic(err)
	}
	err = m.Up()
	if err != nil {
		panic(err)
	}
	err = server.Start(ctx, config)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"gophermart/internal/app/config"
	"gophermart/internal/app/server"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	config := config.New()
	// db, err := sql.Open("postgres", config.DatabaseURI+"?sslmode=disable")
	// if err != nil {
	// 	panic(err)
	// }
	// driver, err := postgres.WithInstance(db, &postgres.Config{})
	// if err != nil {
	// 	panic(err)
	// }
	// m, err := migrate.NewWithDatabaseInstance(
	// 	"file://../../migrations",
	// 	"postgres", driver)
	// if err != nil {
	// 	panic(err)
	// }
	// err = m.Up()
	// if err != nil {
	// 	panic(err)
	// }
	err := server.Start(ctx, config)
	if err != nil {
		panic(err)
	}
}

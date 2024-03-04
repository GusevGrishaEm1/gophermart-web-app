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
	err := server.Start(ctx, config)
	if err != nil {
		panic(err)
	}
}

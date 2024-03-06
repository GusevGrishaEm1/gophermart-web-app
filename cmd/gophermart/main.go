package main

import (
	"context"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/server"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	config, err := config.New(ctx)
	if err != nil {
		panic(err)
	}
	err = server.Start(ctx, config)
	if err != nil {
		panic(err)
	}
}

package server

import (
	"context"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	handlers "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http/middleware"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase/job"

	"github.com/go-chi/chi"
)

// `POST /api/user/register` — регистрация пользователя;
// `POST /api/user/login` — аутентификация пользователя;
// `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
// `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
// `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.

type BalanceOperationHandler interface {
	CreateOrderHandler(w http.ResponseWriter, r *http.Request)
	GetOrdersHandler(w http.ResponseWriter, r *http.Request)
	GetBalanceHandler(w http.ResponseWriter, r *http.Request)
	WithdrawHandler(w http.ResponseWriter, r *http.Request)
	GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
}

type UserHandler interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
}

type SecurityMiddleware interface {
	SecurityMiddleware(h http.Handler) http.Handler
}

type LoggingMiddleware interface {
	LoggingMiddleware(h http.Handler) http.Handler
}

type CompressionMiddleware interface {
	CompressionMiddleware(h http.Handler) http.Handler
}

func Start(cxt context.Context, config *config.Config) error {
	pool, err := InitPool(cxt, config)
	if err != nil {
		return err
	}
	err = InitTables(cxt, pool)
	if err != nil {
		return err
	}
	// err := runMigrate(config)
	// if err != nil {
	// 	return err
	// }
	userRepo, err := repository.NewUserRepository(cxt, config, pool)
	if err != nil {
		return err
	}
	userService := usecase.NewUserService(config, userRepo)
	userHandler := handlers.NewUserHandler(config, userService)

	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, config, pool)
	if err != nil {
		return err
	}
	balanceOperationService := usecase.NewBalanceOperationService(config, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(config, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)

	balanceOperationJob := job.NewBalanceOperationJob(config, balanceOperationRepo)

	go balanceOperationJob.ConsumeOrder(cxt)
	go balanceOperationJob.ProduceOrder(cxt)

	rMain := chi.NewRouter()
	rMain.Post("/api/user/register", userHandler.RegisterHandler)
	rMain.Post("/api/user/login", userHandler.LoginHandler)
	rBalanceOperation := chi.NewRouter()
	rBalanceOperation.Use(securityMiddleware.SecurityMiddleware)
	rBalanceOperation.Get("/api/user/orders", balanceOperationhandler.GetOrdersHandler)
	rBalanceOperation.Get("/api/user/balance", balanceOperationhandler.GetBalanceHandler)
	rBalanceOperation.Get("/api/user/withdrawals", balanceOperationhandler.GetWithdrawalsHandler)
	rBalanceOperation.Post("/api/user/orders", balanceOperationhandler.CreateOrderHandler)
	rBalanceOperation.Post("/api/user/balance/withdraw", balanceOperationhandler.WithdrawHandler)
	rMain.Mount("/", rBalanceOperation)

	err = http.ListenAndServe(config.RunAddress, rMain)
	return err
}

// func runMigrate(config *config.Config) error {
// 	db, err := sql.Open("postgres", config.DatabaseURI)
// 	if err != nil {
// 		panic(err)
// 	}
// 	driver, err := postgres.WithInstance(db, &postgres.Config{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	m, err := migrate.NewWithDatabaseInstance(
// 		"file://../../migrations",
// 		"postgres", driver)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = m.Up()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return err
// }

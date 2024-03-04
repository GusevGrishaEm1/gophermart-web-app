package server

import (
	"context"
	"gophermart/cmd/gophermart/internal/app/config"
	handlers "gophermart/cmd/gophermart/internal/app/controller/http"
	"gophermart/cmd/gophermart/internal/app/controller/http/middleware"
	"gophermart/cmd/gophermart/internal/app/infrastructure/repository"
	"gophermart/cmd/gophermart/internal/app/usecase"
	"gophermart/cmd/gophermart/internal/app/usecase/job"
	"net/http"

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
	userRepo, err := repository.NewUserRepository(cxt, config)
	if err != nil {
		return err
	}
	userService := usecase.NewUserService(config, userRepo)
	userHandler := handlers.NewUserHandler(config, userService)

	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, config)
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
	rBalanceOperation := chi.NewRouter()
	rBalanceOperation.Use(securityMiddleware.SecurityMiddleware)
	rMain.Post("/api/user/register", userHandler.RegisterHandler)
	rMain.Post("/api/user/login", userHandler.LoginHandler)
	rBalanceOperation.Post("/api/user/orders", balanceOperationhandler.CreateOrderHandler)
	rBalanceOperation.Get("/api/user/orders", balanceOperationhandler.GetOrdersHandler)
	rBalanceOperation.Get("/api/user/balance", balanceOperationhandler.GetBalanceHandler)
	rBalanceOperation.Post("/api/user/balance/withdraw", balanceOperationhandler.WithdrawHandler)
	rBalanceOperation.Get("/api/user/withdrawals", balanceOperationhandler.GetWithdrawalsHandler)
	rMain.Mount("/", rBalanceOperation)

	err = http.ListenAndServe(config.RunAddress, rMain)
	return err
}

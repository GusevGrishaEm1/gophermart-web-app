package server

import (
	"context"
	"net/http"
	"os"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	handlers "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http/middleware"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/webapi"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase/job"
	log "github.com/go-kit/log"

	"github.com/go-chi/chi"
)

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

func Start(ctx context.Context, config *config.Config) error {
	err := initTables(ctx, config.Pool)
	if err != nil {
		return err
	}

	userRepo, err := repository.NewUserRepository(ctx, config)
	if err != nil {
		return err
	}
	userService := usecase.NewUserService(config, userRepo)
	userHandler := handlers.NewUserHandler(config, userService)

	balanceOperationRepo, err := repository.NewBalanceOperationRepository(ctx, config)
	if err != nil {
		return err
	}
	balanceOperationService := usecase.NewBalanceOperationService(config, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(config, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "loc", log.DefaultCaller)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	compressionMiddleware := middleware.NewCompressionMiddleware()

	runJobs(ctx, config, balanceOperationRepo)

	r := getRouter(userHandler, securityMiddleware, balanceOperationhandler)
	r.Use(loggingMiddleware.LoggingMiddleware)
	r.Use(compressionMiddleware.CompressionMiddleware)

	err = http.ListenAndServe(config.RunAddress, r)
	return err
}

func runJobs(ctx context.Context, config *config.Config, balanceOperationRepo repository.BalanceOperationRepository) {
	balanceOperationJob := job.NewBalanceOperationJob(config, balanceOperationRepo, webapi.NewAccrualWebAPI(config))
	go balanceOperationJob.ConsumeOrder(ctx)
	go balanceOperationJob.ProduceOrder(ctx)
}

func getRouter(userH *handlers.UserHandler, securityM *middleware.SecurityMiddleware, balanceH *handlers.BalanceOperationHandler) *chi.Mux {
	rMain := chi.NewRouter()
	rMain.Post("/api/user/register", userH.RegisterHandler)
	rMain.Post("/api/user/login", userH.LoginHandler)
	rBalanceOperation := chi.NewRouter()
	rBalanceOperation.Use(securityM.SecurityMiddleware)
	rBalanceOperation.Get("/api/user/orders", balanceH.GetOrdersHandler)
	rBalanceOperation.Get("/api/user/balance", balanceH.GetBalanceHandler)
	rBalanceOperation.Get("/api/user/withdrawals", balanceH.GetWithdrawalsHandler)
	rBalanceOperation.Post("/api/user/orders", balanceH.CreateOrderHandler)
	rBalanceOperation.Post("/api/user/balance/withdraw", balanceH.WithdrawHandler)
	rMain.Mount("/", rBalanceOperation)
	return rMain
}

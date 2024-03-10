package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	handlers "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/controller/http/middleware"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/usecase"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var c *config.Config

type AccrualWebAPIForTest struct{}

func (webAPI *AccrualWebAPIForTest) GetAccrualRequest(order string) (*entity.AccrualResponse, error) {
	return &entity.AccrualResponse{
		Order:   order,
		Status:  entity.PROCESSED,
		Accrual: 200.00,
	}, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	testDB := SetupTestDatabase()
	conf := &config.Config{}
	conf.Pool = testDB.DbInstance
	err := initTables(ctx, conf.Pool)
	if err != nil {
		return
	}
	c = conf
	defer testDB.TearDown()
	os.Exit(m.Run())
}

func TestRegisterHandler(t *testing.T) {
	ctx := context.Background()
	userRepo, err := repository.NewUserRepository(ctx, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name: "test#1",
			body: `
			{
				"login": "test",
				"password": "test"
			}
			`,
			expectedStatus: 200,
		},
		{
			name: "test#2",
			body: `
			{
				"login": "test",
				"password": "test2"
			}
			`,
			expectedStatus: 409,
		},
		{
			name: "test#3",
			body: `
			{
				"lain": "test",
				"password": "test3"
			}
			`,
			expectedStatus: 400,
		},
		{
			name: "test#4",
			body: `
			{
				"login": "test2",
				"password": "test2"
			}
			`,
			expectedStatus: 200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(test.body)))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			userHandler.RegisterHandler(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctx := context.Background()
	userRepo, err := repository.NewUserRepository(ctx, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name: "test#1",
			body: `
			{
				"login": "test",
				"password": "test"
			}
			`,
			expectedStatus: 200,
		},
		{
			name: "test#2",
			body: `
			{
				"login": "test",
				"password": "test2"
			}
			`,
			expectedStatus: 401,
		},
		{
			name: "test#3",
			body: `
			{
				"login": "test",
				"addd": "test"
			}
			`,
			expectedStatus: 400,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte(test.body)))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			userHandler.LoginHandler(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func login(login string, password string, userHandler UserHandler) string {
	body := fmt.Sprintf(`
	{
		"login": "%s",
		"password": "%s"
	}
	`, login, password)
	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte(body)))
	request.Header.Set("content-type", "application/json")
	w := httptest.NewRecorder()
	userHandler.LoginHandler(w, request)
	res := w.Result()
	if len(res.Cookies()) == 0 {
		return ""
	}
	return res.Cookies()[0].Value
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func generateTokenToCheat(t *testing.T) string {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			UserID: 999,
		},
	)
	tokenString, err := token.SignedString([]byte("secretkey"))
	require.NoError(t, err)
	return tokenString
}

func TestCreateOrderHandler(t *testing.T) {
	cxt := context.Background()
	userRepo, err := repository.NewUserRepository(cxt, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, c)
	require.NoError(t, err)
	balanceOperationService := usecase.NewBalanceOperationService(c, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(c, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)
	handler := securityMiddleware.SecurityMiddleware(http.HandlerFunc(balanceOperationhandler.CreateOrderHandler))
	tests := []struct {
		name           string
		token          string
		order          string
		expectedStatus int
	}{
		{
			name:           "test#1",
			token:          login("test", "test", userHandler),
			order:          "12345678903",
			expectedStatus: 202,
		},
		{
			name:           "test#2",
			token:          login("test", "test", userHandler),
			order:          "12345678923",
			expectedStatus: 422,
		},
		{
			name:           "test#3",
			token:          login("test", "test", userHandler),
			order:          "12345678903",
			expectedStatus: 200,
		},
		{
			name:           "test#4",
			token:          login("test2", "test2", userHandler),
			order:          "12345678903",
			expectedStatus: 409,
		},
		{
			name:           "test#5",
			token:          login("test", "aa", userHandler),
			order:          "12345678903",
			expectedStatus: 401,
		},
		{
			name:           "test#6",
			token:          generateTokenToCheat(t),
			order:          "12345578903",
			expectedStatus: 401,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(test.order)))
			request.Header.Set("Content-Type", "text/plain")
			request.AddCookie(&http.Cookie{
				Name:  "USER_ID",
				Value: test.token,
			})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestGetOrdersHandler(t *testing.T) {
	cxt := context.Background()
	userRepo, err := repository.NewUserRepository(cxt, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, c)
	require.NoError(t, err)
	balanceOperationService := usecase.NewBalanceOperationService(c, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(c, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)
	handler := securityMiddleware.SecurityMiddleware(http.HandlerFunc(balanceOperationhandler.GetOrdersHandler))
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "test#1",
			token:          login("test", "test", userHandler),
			expectedStatus: 200,
		},
		{
			name:           "test#2",
			token:          login("test2", "test2", userHandler),
			expectedStatus: 204,
		},
		{
			name:           "test#3",
			token:          login("test", "aa", userHandler),
			expectedStatus: 401,
		},
		{
			name:           "test#4",
			token:          generateTokenToCheat(t),
			expectedStatus: 401,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			request.AddCookie(&http.Cookie{
				Name:  "USER_ID",
				Value: test.token,
			})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestGetBalanceHandler(t *testing.T) {
	cxt := context.Background()
	userRepo, err := repository.NewUserRepository(cxt, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, c)
	require.NoError(t, err)
	balanceOperationService := usecase.NewBalanceOperationService(c, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(c, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)
	handler := securityMiddleware.SecurityMiddleware(http.HandlerFunc(balanceOperationhandler.GetBalanceHandler))
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "test#1",
			token:          login("test", "test", userHandler),
			expectedStatus: 200,
		},
		{
			name:           "test#2",
			token:          login("test2", "test2", userHandler),
			expectedStatus: 200,
		},
		{
			name:           "test#3",
			token:          login("test", "aa", userHandler),
			expectedStatus: 401,
		},
		{
			name:           "test#4",
			token:          generateTokenToCheat(t),
			expectedStatus: 401,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			request.AddCookie(&http.Cookie{
				Name:  "USER_ID",
				Value: test.token,
			})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func prepareData(context context.Context, config *config.Config) {
	query := `
		update "balance_operation"
		set 
			status = 'PROCESSED',
			sum = 20000
		where "deleted_at" is null
		and type = 'ACCRUAL'
		and status = 'NEW';
	`
	config.Pool.Exec(context, query)
}

func TestWithdrawHandler(t *testing.T) {
	cxt := context.Background()
	prepareData(cxt, c)
	userRepo, err := repository.NewUserRepository(cxt, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, c)
	require.NoError(t, err)
	balanceOperationService := usecase.NewBalanceOperationService(c, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(c, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)
	handler := securityMiddleware.SecurityMiddleware(http.HandlerFunc(balanceOperationhandler.WithdrawHandler))
	tests := []struct {
		name           string
		token          string
		body           string
		expectedStatus int
	}{
		{
			name:  "test#1",
			token: login("test", "test", userHandler),
			body: `
			{
				"order": "2377225624",
				"sum": 100
			}
			`,
			expectedStatus: 200,
		},
		{
			name:  "test#2",
			token: login("test2", "test2", userHandler),
			body: `
			{
				"order": "2377225624",
				"sum": 500
			}
			`,
			expectedStatus: 402,
		},
		{
			name:  "test#3",
			token: login("test", "fff", userHandler),
			body: `
			{
				"order": "2375625624",
				"sum": 100
			}
			`,
			expectedStatus: 401,
		},
		{
			name:  "test#4",
			token: login("test2", "test2", userHandler),
			body: `
			{
				"order": "2277225624",
				"sum": 500
			}
			`,
			expectedStatus: 422,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte(test.body)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(&http.Cookie{
				Name:  "USER_ID",
				Value: test.token,
			})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestGetWithdrawalsHandler(t *testing.T) {
	cxt := context.Background()
	userRepo, err := repository.NewUserRepository(cxt, c)
	require.NoError(t, err)
	userService := usecase.NewUserService(c, userRepo)
	userHandler := handlers.NewUserHandler(c, userService)
	balanceOperationRepo, err := repository.NewBalanceOperationRepository(cxt, c)
	require.NoError(t, err)
	balanceOperationService := usecase.NewBalanceOperationService(c, balanceOperationRepo)
	balanceOperationhandler := handlers.NewBalanceOperationHandler(c, balanceOperationService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(userService)
	handler := securityMiddleware.SecurityMiddleware(http.HandlerFunc(balanceOperationhandler.GetWithdrawalsHandler))
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "test#1",
			token:          login("test", "test", userHandler),
			expectedStatus: 200,
		},
		{
			name:           "test#2",
			token:          login("test2", "test2", userHandler),
			expectedStatus: 204,
		},
		{
			name:           "test#3",
			token:          login("test", "aa", userHandler),
			expectedStatus: 401,
		},
		{
			name:           "test#4",
			token:          generateTokenToCheat(t),
			expectedStatus: 401,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			request.AddCookie(&http.Cookie{
				Name:  "USER_ID",
				Value: test.token,
			})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			log.Print(string(data))
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

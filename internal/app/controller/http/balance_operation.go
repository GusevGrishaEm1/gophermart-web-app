package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"
)

type BalanceOperationService interface {
	CreateNewOrder(context.Context, *CreateOrderRequest) error
	GetListOrders(ctx context.Context, userID int) ([]*OrderResponse, error)
	GetBalance(ctx context.Context, userID int) (*BalanceResponse, error)
	CreateWithdraw(ctx context.Context, userID int, withdraw *WithdrawRequest) error
	GetWithdrawals(ctx context.Context, userID int) ([]*WithdrawResponse, error)
}

type BalanceOperationHandler struct {
	c *config.Config
	BalanceOperationService
	UserService
}

func NewBalanceOperationHandler(c *config.Config, balanceS BalanceOperationService, userS UserService) *BalanceOperationHandler {
	return &BalanceOperationHandler{c, balanceS, userS}
}

type CreateOrderRequest struct {
	Order  string
	UserID int
}

func (h *BalanceOperationHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order := string(buf)
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dto := &CreateOrderRequest{
		Order:  order,
		UserID: userID,
	}
	err = h.CreateNewOrder(r.Context(), dto)
	shouldReturn := h.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type OrderResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

func (h *BalanceOperationHandler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseArr, err := h.GetListOrders(r.Context(), userID)
	shouldReturn := h.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	data, err := json.Marshal(responseArr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type BalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func (h *BalanceOperationHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response, err := h.GetBalance(r.Context(), userID)
	shouldReturn := h.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	data, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h *BalanceOperationHandler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var withdraw *WithdrawRequest
	err = json.Unmarshal(buf, &withdraw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.CreateWithdraw(r.Context(), userID, withdraw)
	shouldReturn := h.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type WithdrawResponse struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *BalanceOperationHandler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseArr, err := h.GetWithdrawals(r.Context(), userID)
	shouldReturn := h.validateErrorAfter(err, w)
	if shouldReturn {
		return
	}
	data, err := json.Marshal(responseArr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (*BalanceOperationHandler) validateErrorAfter(err error, w http.ResponseWriter) bool {
	if err != nil {
		customErr := &customerr.CustomError{}
		if errors.As(err, &customErr) {
			w.WriteHeader(customErr.HTTPStatus)
			return true
		}
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
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
		sendClientErr(err, w)
		return
	}
	order := string(buf)
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		sendServerErr(err, w)
		return
	}
	dto := &CreateOrderRequest{
		Order:  order,
		UserID: userID,
	}
	err = h.CreateNewOrder(r.Context(), dto)
	if err != nil {
		sendServerErr(err, w)
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
		sendServerErr(err, w)
		return
	}
	responseArr, err := h.GetListOrders(r.Context(), userID)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	sendOKWithBody(w, responseArr)
}

type BalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func (h *BalanceOperationHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		sendServerErr(err, w)
		return
	}
	balanceResponse, err := h.GetBalance(r.Context(), userID)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	sendOKWithBody(w, balanceResponse)
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h *BalanceOperationHandler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(io.Reader(r.Body))
	if err != nil {
		sendClientErr(err, w)
		return
	}
	var withdraw *WithdrawRequest
	err = json.Unmarshal(buf, &withdraw)
	if err != nil {
		sendClientErr(err, w)
		return
	}
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		sendServerErr(err, w)
		return
	}
	err = h.CreateWithdraw(r.Context(), userID, withdraw)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type WithdrawResponse struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *BalanceOperationHandler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.GetUserIDFromContext(r.Context())
	if err != nil {
		sendServerErr(err, w)
		return
	}
	responseArr, err := h.GetWithdrawals(r.Context(), userID)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	sendOKWithBody(w, responseArr)
}

package usecase

import (
	"context"
	"errors"
	"gophermart/internal/app/config"
	"gophermart/internal/app/controller/http"
	"gophermart/internal/app/entity"
	customerr "gophermart/internal/app/error"
	"gophermart/internal/app/infrastructure/repository"
	nethttp "net/http"
	"strconv"
	"time"
)

type BalanceOperationService struct {
	c *config.Config
	repository.BalanceOperationRepository
}

func NewBalanceOperationService(c *config.Config, r repository.BalanceOperationRepository) *BalanceOperationService {
	return &BalanceOperationService{c, r}
}

func (s *BalanceOperationService) CreateNewOrder(ctx context.Context, dto *http.CreateOrderRequest) error {
	if !checkLuhn(dto.Order) {
		return customerr.NewError(errors.New("luhn alg validation failed"), nethttp.StatusUnprocessableEntity)
	}
	balanceOperation := &entity.BalanceOperation{
		Order:  dto.Order,
		UserID: dto.UserID,
		Status: entity.NEW,
		Type:   entity.ACCRUAL,
	}
	return s.SaveOrder(ctx, balanceOperation)
}

func checkLuhn(order string) bool {
	sum := 0
	len := len([]rune(order))
	parity := len % 2
	for i := 0; i < len; i++ {
		num, err := strconv.Atoi(string(order[i]))
		if err != nil {
			return false
		}
		if i%2 == parity {
			num *= 2
			if num > 9 {
				num -= 9
			}
		}
		sum += num
	}
	return sum%10 == 0
}

func (s *BalanceOperationService) GetListOrders(ctx context.Context, userID int) ([]*http.OrderResponse, error) {
	entityArr, err := s.FindOrdersByUser(ctx, userID)
	if err != nil {
		return nil, customerr.NewError(err, nethttp.StatusInternalServerError)
	}
	responseArr := make([]*http.OrderResponse, len(entityArr))
	for i, entity := range entityArr {
		response := &http.OrderResponse{
			Number:     entity.Order,
			Status:     string(entity.Status),
			Accrual:    float32(entity.Sum) / 100,
			UploadedAt: entity.CreatedAt.Format(time.RFC3339),
		}
		responseArr[i] = response
	}
	return responseArr, nil
}

func (s *BalanceOperationService) GetBalance(ctx context.Context, userID int) (*http.BalanceResponse, error) {
	current, withdrawn, err := s.GetBalanceByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := &http.BalanceResponse{
		Current:   float32(current) / 100,
		Withdrawn: float32(withdrawn) / 100,
	}
	return result, nil
}

func (s *BalanceOperationService) CreateWithdraw(ctx context.Context, userID int, withdraw *http.WithdrawRequest) error {
	if !checkLuhn(withdraw.Order) {
		return customerr.NewError(errors.New("luhn alg validation failed"), nethttp.StatusUnprocessableEntity)
	}
	balanceOperation := &entity.BalanceOperation{
		Order:  withdraw.Order,
		Sum:    int(withdraw.Sum*100) * (-1),
		UserID: userID,
		Status: entity.PROCESSED,
		Type:   entity.WITHDRAW,
	}
	return s.SaveWithdraw(ctx, balanceOperation)
}

func (s *BalanceOperationService) GetWithdrawals(ctx context.Context, userID int) ([]*http.WithdrawResponse, error) {
	entityArr, err := s.FindWithdrawsByUser(ctx, userID)
	if err != nil {
		return nil, customerr.NewError(err, nethttp.StatusInternalServerError)
	}
	responseArr := make([]*http.WithdrawResponse, len(entityArr))
	for i, entity := range entityArr {
		response := &http.WithdrawResponse{
			Order:       entity.Order,
			Sum:         float32(entity.Sum) / 100 * (-1),
			ProcessedAt: entity.CreatedAt.Format(time.RFC3339),
		}
		responseArr[i] = response
	}
	return responseArr, nil
}

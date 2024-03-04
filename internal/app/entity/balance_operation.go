package entity

import "time"

type ProcessStatus string

type BalanceOperationType string

const (
	NEW        ProcessStatus = ProcessStatus("NEW")
	PROCESSING               = "PROCESSING"
	INVALID                  = "INVALID"
	PROCESSED                = "PROCESSED"
)

const (
	ACCRUAL  BalanceOperationType = BalanceOperationType("ACCRUAL")
	WITHDRAW                      = "WITHDRAW"
)

// Операция с балансом пользователя
type BalanceOperation struct {
	ID        int
	Sum       int
	Order     string
	Status    ProcessStatus
	Type      BalanceOperationType
	UserID    int
	CreatedAt time.Time
	DeletedAt time.Time
}

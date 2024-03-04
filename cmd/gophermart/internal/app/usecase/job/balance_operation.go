package job

import (
	"context"
	"gophermart/cmd/gophermart/internal/app/config"
	"gophermart/cmd/gophermart/internal/app/entity"
	"gophermart/cmd/gophermart/internal/app/infrastructure/repository"
	"gophermart/cmd/gophermart/internal/app/infrastructure/webapi"
	"time"
)

const SIZE_ARRAY_TO_UPDATE int = 1000

type BalanceOperationJob struct {
	chToUpdateAccrual chan *entity.BalanceOperation
	*webapi.AccrualWebAPI
	repository.BalanceOperationRepository
}

func NewBalanceOperationJob(config *config.Config, r repository.BalanceOperationRepository) *BalanceOperationJob {
	return &BalanceOperationJob{
		make(chan *entity.BalanceOperation, 1024),
		&webapi.AccrualWebAPI{C: config},
		r,
	}
}

func (j *BalanceOperationJob) ProduceOrder(ctx context.Context) {
	defer close(j.chToUpdateAccrual)
	ticker := time.NewTicker(5000 * time.Millisecond)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			orders, err := j.FindOrdersToProcess(ctx)
			if err != nil {
				continue
			}
			for _, el := range orders {
				j.chToUpdateAccrual <- el
			}
		}
	}
}

func (j *BalanceOperationJob) ConsumeOrder(ctx context.Context) {
	arrayToUpdate := make([]*entity.BalanceOperation, 0)
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer func() {
		if len(arrayToUpdate) > 0 {
			j.UpdateOrders(ctx, arrayToUpdate)
			arrayToUpdate = arrayToUpdate[:]
		}
	}()
loop:
	for {
		select {
		case el := <-j.chToUpdateAccrual:
			response, err := j.GetAccrualRequest(el.Order)
			if err != nil {
				continue
			}
			el.Sum = int(response.Accrual * 100)
			el.Status = entity.ProcessStatus(response.Status)
			arrayToUpdate = append(arrayToUpdate, el)
			if len(arrayToUpdate) > SIZE_ARRAY_TO_UPDATE {
				j.UpdateOrders(ctx, arrayToUpdate)
				arrayToUpdate = arrayToUpdate[:]
			}
		case <-ticker.C:
			if len(arrayToUpdate) > 0 {
				j.UpdateOrders(ctx, arrayToUpdate)
				arrayToUpdate = arrayToUpdate[:]
			}
		case <-ctx.Done():
			break loop
		}
	}
}

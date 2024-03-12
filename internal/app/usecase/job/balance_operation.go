package job

import (
	"context"
	"time"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/infrastructure/repository"
)

const MaxArraySize int = 1000

type AccrualWebAPI interface {
	GetAccrualRequest(order string) (*entity.AccrualResponse, error)
}

type BalanceOperationJob struct {
	chToUpdateAccrual chan *entity.BalanceOperation
	AccrualWebAPI
	repository.BalanceOperationRepository
}

func NewBalanceOperationJob(config *config.Config, r repository.BalanceOperationRepository, webAPI AccrualWebAPI) *BalanceOperationJob {
	return &BalanceOperationJob{
		make(chan *entity.BalanceOperation, 1024),
		webAPI,
		r,
	}
}

func (j *BalanceOperationJob) ProduceOrder(ctx context.Context) {
	defer close(j.chToUpdateAccrual)
	ticker := time.NewTicker(500 * time.Millisecond)
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
	ticker := time.NewTicker(500 * time.Millisecond)
	defer func() {
		if len(arrayToUpdate) > 0 {
			j.UpdateOrders(ctx, arrayToUpdate)
			arrayToUpdate = arrayToUpdate[:0]
		}
	}()
loop:
	for {
		select {
		case el := <-j.chToUpdateAccrual:
			response, err := j.GetAccrualRequest(el.Order)
			if err != nil {
				el.Sum = 0
				el.Status = entity.ProcessStatus("NEW")
			} else {
				el.Sum = int(response.Accrual * 100)
				el.Status = entity.ProcessStatus(response.Status)
			}
			arrayToUpdate = append(arrayToUpdate, el)
			if len(arrayToUpdate) > MaxArraySize {
				j.UpdateOrders(ctx, arrayToUpdate)
				arrayToUpdate = arrayToUpdate[:0]
			}
		case <-ticker.C:
			if len(arrayToUpdate) > 0 {
				j.UpdateOrders(ctx, arrayToUpdate)
				arrayToUpdate = arrayToUpdate[:0]
			}
		case <-ctx.Done():
			break loop
		}
	}
}

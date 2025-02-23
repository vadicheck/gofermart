package sync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vadicheck/gofermart/internal/app/accrualservice"
	"github.com/vadicheck/gofermart/internal/app/client/http"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

const delay = 5000
const maxQueueLength = 500
const httpClientTimeout = 10000

type App struct {
	accrualAddress string
	orderService   orderService
	logger         logger.LogClient
}

type orderService interface {
	GetOrderByID(ctx context.Context, orderID int) (gofermart.Order, error)
	GetOrdersIdsByStatus(ctx context.Context, statuses []constants.OrderStatus, logger logger.LogClient) ([]int, error)
	UpdateOrder(ctx context.Context, orderID int, newStatus constants.OrderStatus, accrual int, logger logger.LogClient) error
}

func New(
	accrualAddress string,
	orderService orderService,
	logger logger.LogClient,
) *App {
	return &App{
		accrualAddress: accrualAddress,
		orderService:   orderService,
		logger:         logger,
	}
}

func (sa *App) Run(
	ctx context.Context,
	orderService orderService,
	wg *sync.WaitGroup,
) error {
	sa.logger.Info(fmt.Sprintf("sync app starting: %s", sa.accrualAddress))

	transport := http.NewHTTPClient(
		"accrual-service",
		nil,
		time.Millisecond*time.Duration(httpClientTimeout),
		sa.logger,
	)
	accrualService := accrualservice.New(transport, sa.accrualAddress, sa.logger)

	wg.Add(1)
	go func() {
		jobs := make(chan int, maxQueueLength)
		results := make(chan int, maxQueueLength)

		defer close(jobs)
		defer close(results)

		for w := 1; w <= 3; w++ {
			go sa.handleOrder(ctx, accrualService, jobs)
		}

		statuses := []constants.OrderStatus{
			constants.StatusNew,
			constants.StatusProcessing,
		}

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Exit sync")
				wg.Done()
				return
			default:
				orderIds, err := orderService.GetOrdersIdsByStatus(ctx, statuses, sa.logger)
				if err != nil {
					sa.logger.Error(fmt.Errorf("failed to get order ids by status. err: %w", err))
				} else {
					for _, orderID := range orderIds {
						jobs <- orderID
					}
				}
			}
			time.Sleep(delay * time.Millisecond)
		}
	}()

	return nil
}

func (sa *App) handleOrder(
	ctx context.Context,
	accrualService accrualservice.Service,
	jobs <-chan int,
) {
	for orderID := range jobs {
		order, err := sa.orderService.GetOrderByID(ctx, orderID)
		if err != nil {
			sa.logger.Error(fmt.Errorf("failed to get order by id. err: %w", err))
			continue
		}

		if order.ID == 0 {
			sa.logger.Error(errors.New("order not found"))
			continue
		}

		orderStatus := constants.OrderStatus(order.Status)

		if orderStatus != constants.StatusNew && orderStatus != constants.StatusProcessing {
			continue
		}

		orderResponse, err := accrualService.GetOrder(ctx, orderID)
		if err != nil {
			sa.logger.Error(fmt.Errorf("failed to get order from accrual system. err: %w", err))
		}

		err = sa.orderService.UpdateOrder(
			ctx,
			orderID,
			constants.OrderStatus(orderResponse.Status),
			orderResponse.Accrual,
			sa.logger,
		)
		if err != nil {
			sa.logger.Error(fmt.Errorf("failed to update order. err: %w", err))
		}

		time.Sleep(time.Second)
	}
}

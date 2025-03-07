package sync

import (
	"context"
	"database/sql"
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
	orderService   OrderService
	storage        storage
	logger         logger.LogClient
}

type storage interface {
	Apply(ctx context.Context, orderID string, userID int, orderResponse *accrualservice.GetOrderResponse) error
}

type OrderService interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)

	GetUserByID(ctx context.Context, userID int) (gofermart.User, error)
	ChangeUserBalance(ctx context.Context, userID int, balance float32) error

	GetOrderByID(ctx context.Context, orderID string) (gofermart.Order, error)
	GetOrdersIdsByStatus(ctx context.Context, statuses []constants.OrderStatus) ([]string, error)
	UpdateOrder(ctx context.Context, orderID string, newStatus constants.OrderStatus, accrual float32) error
}

func New(
	accrualAddress string,
	orderService OrderService,
	storage storage,
	logger logger.LogClient,
) *App {
	return &App{
		accrualAddress: accrualAddress,
		orderService:   orderService,
		storage:        storage,
		logger:         logger,
	}
}

func (sa *App) Run(ctx context.Context, wg *sync.WaitGroup) error {
	sa.logger.Info(fmt.Sprintf("sync app starting: %s", sa.accrualAddress))

	transport := http.NewHTTPClient(
		"accrual-service",
		nil,
		time.Millisecond*time.Duration(httpClientTimeout),
		sa.logger,
	)
	accrualService := accrualservice.New(transport, sa.accrualAddress, sa.logger)

	var m sync.Mutex

	wg.Add(1)
	go func() {
		jobs := make(chan string, maxQueueLength)
		results := make(chan int, maxQueueLength)

		defer close(jobs)
		defer close(results)

		for w := 1; w <= 5; w++ {
			go sa.handleOrder(ctx, accrualService, &m, jobs)
		}

		statuses := []constants.OrderStatus{
			constants.StatusNew,
			constants.StatusProcessing,
		}

		for {
			select {
			case <-ctx.Done():
				sa.logger.Info("Exit sync")
				wg.Done()
				return
			default:
				orderIds, err := sa.orderService.GetOrdersIdsByStatus(ctx, statuses)
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
	accrualService accrualservice.Client,
	m *sync.Mutex,
	jobs <-chan string,
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

		if orderResponse == nil {
			continue
		}

		m.Lock()
		err = sa.storage.Apply(ctx, orderID, order.UserID, orderResponse)
		m.Unlock()

		if err != nil {
			sa.logger.Error(fmt.Errorf("failed to apply order changes. err: %w", err))
		}
	}
}

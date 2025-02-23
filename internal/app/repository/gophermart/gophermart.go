package gophermart

import (
	"context"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Gophermart interface {
	CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int, error)
	GetUserByLogin(ctx context.Context, login string) (gofermart.User, error)

	GetOrders(ctx context.Context, userID int, logger logger.LogClient) ([]gofermart.Order, error)
	CreateOrder(ctx context.Context, orderID, userID int, logger logger.LogClient) (int, error)
	GetOrderByID(ctx context.Context, orderID int) (gofermart.Order, error)
	GetOrdersIdsByStatus(ctx context.Context, statuses []constants.OrderStatus, logger logger.LogClient) ([]int, error)
	UpdateOrder(ctx context.Context, orderID int, newStatus constants.OrderStatus, accrual int, logger logger.LogClient) error
}

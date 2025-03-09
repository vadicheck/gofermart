package gophermart

import (
	"context"
	"database/sql"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
)

type Gophermart interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)

	CreateUser(ctx context.Context, login, password string) (int, error)
	GetUserByID(ctx context.Context, userID int) (gofermart.User, error)
	GetUserByLogin(ctx context.Context, login string) (gofermart.User, error)
	ChangeUserBalance(ctx context.Context, userID int, balance float32) error

	GetOrders(ctx context.Context, userID int) ([]gofermart.Order, error)
	CreateOrder(ctx context.Context, orderID string, userID int) (int, error)
	GetOrderByID(ctx context.Context, orderID string) (gofermart.Order, error)
	GetOrdersIdsByStatus(ctx context.Context, statuses []constants.OrderStatus) ([]string, error)
	UpdateOrder(ctx context.Context, orderID string, newStatus constants.OrderStatus, accrual float32) error

	CreateTransaction(ctx context.Context, userID int, orderID string, sum float32) error
	CreateTransactionAndChangeBalance(ctx context.Context, userID int, orderID string, sum float32, newBalance float32) error
	GetTransactionsByUserID(ctx context.Context, userID int) ([]gofermart.Transaction, error)

	GetTotalWithdrawn(ctx context.Context, userID int) (float32, error)
}

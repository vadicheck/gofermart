package balance

import (
	"context"

	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/services"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type withdrawStorage interface {
	CreateTransactionAndChangeBalance(ctx context.Context, userID int, orderID string, sum float32, newBalance float32) error
}

type Service interface {
	Withdraw(ctx context.Context, user gofermart.User, orderID string, sum float32) error
}

type service struct {
	storage withdrawStorage
	logger  logger.LogClient
}

func New(storage withdrawStorage, logger logger.LogClient) Service {
	return &service{
		storage: storage,
		logger:  logger,
	}
}

func (s *service) Withdraw(
	ctx context.Context,
	user gofermart.User,
	orderID string,
	sum float32,
) error {
	if user.Balance < sum {
		return services.ErrInsufficientBalance
	}

	return s.storage.CreateTransactionAndChangeBalance(ctx, user.ID, orderID, sum, user.Balance-sum)
}

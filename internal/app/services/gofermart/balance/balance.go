package balance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/services"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type withdrawStorage interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
	ChangeUserBalance(ctx context.Context, userID int, balance float32) error
	CreateTransaction(ctx context.Context, userID int, orderID string, sum float32) error
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

	tx, err := s.storage.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				s.logger.Error(fmt.Errorf("transaction rollback error: %w", err))
			}
		}
	}()

	err = s.storage.CreateTransaction(ctx, user.ID, orderID, sum)
	if err != nil {
		return err
	}

	err = s.storage.ChangeUserBalance(ctx, user.ID, user.Balance-sum)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

package balance

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/services"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Storage interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
	ChangeUserBalance(ctx context.Context, userID int, balance int, logger logger.LogClient) error
	CreateTransaction(ctx context.Context, userID int, orderID int, sum int, logger logger.LogClient) error
}

type Service struct {
	storage Storage
	logger  logger.LogClient
}

func New(storage Storage, logger logger.LogClient) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) Withdraw(
	ctx context.Context,
	user gofermart.User,
	orderID int,
	sum int,
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
			s.logger.Error(fmt.Errorf("transaction rollback error: %w", err))
		}
	}()

	err = s.storage.CreateTransaction(ctx, user.ID, orderID, sum, s.logger)
	if err != nil {
		return err
	}

	err = s.storage.ChangeUserBalance(ctx, user.ID, user.Balance-sum, s.logger)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

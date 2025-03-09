package sync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vadicheck/gofermart/internal/app/accrualservice"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/sync"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Storage struct {
	logger       logger.LogClient
	orderService sync.OrderService
}

func New(orderService sync.OrderService, logger logger.LogClient) (*Storage, error) {
	return &Storage{
		logger:       logger,
		orderService: orderService,
	}, nil
}

func (s *Storage) Apply(
	ctx context.Context,
	orderID string,
	userID int,
	orderResponse *accrualservice.GetOrderResponse,
) error {
	tx, err := s.orderService.BeginTransaction(ctx)
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

	if constants.OrderStatus(orderResponse.Status) == constants.StatusProcessed {
		user, getErr := s.orderService.GetUserByID(ctx, userID)
		if getErr != nil {
			return errors.New("user not found")
		}

		err = s.orderService.ChangeUserBalance(ctx, user.ID, user.Balance+orderResponse.Accrual)
		if err != nil {
			return fmt.Errorf("failed to change user balance. err: %w", err)
		}
	}

	err = s.orderService.UpdateOrder(
		ctx,
		orderID,
		constants.OrderStatus(orderResponse.Status),
		orderResponse.Accrual,
	)

	if err != nil {
		return fmt.Errorf("failed to update order. err: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

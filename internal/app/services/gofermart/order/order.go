package order

import (
	"context"

	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Service struct {
	storage gophermart.Gophermart
}

func New(storage gophermart.Gophermart) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) CreateOrder(ctx context.Context, orderID, userID int, logger logger.LogClient) (int, error) {
	return s.storage.CreateOrder(ctx, orderID, userID, logger)
}

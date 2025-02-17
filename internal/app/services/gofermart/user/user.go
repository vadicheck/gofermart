package user

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

func (s *Service) CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int, error) {
	return s.storage.CreateUser(ctx, login, password, logger)
}

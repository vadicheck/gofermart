package gophermart

import (
	"context"

	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Gophermart interface {
	CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int, error)
	GetUserByLogin(ctx context.Context, login string) (gofermart.User, error)
}

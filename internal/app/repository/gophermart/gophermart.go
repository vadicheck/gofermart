package gophermart

import (
	"context"

	"github.com/vadicheck/gofermart/pkg/logger"
)

type Gophermart interface {
	CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int64, error)
}

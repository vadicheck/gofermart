package gophermart

import "context"

type Gophermart interface {
	CreateUser(ctx context.Context, userID int64) (int64, error)
}

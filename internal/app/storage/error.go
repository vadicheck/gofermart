package storage

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginAlreadyExists = errors.New("login already exists")

	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")

	ErrOrderTransactionAlreadyExists = errors.New("the order transaction already exists")
)

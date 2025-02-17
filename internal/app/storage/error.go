package storage

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginAlreadyExists = errors.New("login already exists")
)

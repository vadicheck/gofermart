package gofermart

import "time"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   int64     `json:"order_id"`
	Accrual   int       `json:"accrual"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package gofermart

import "time"

type User struct {
	ID       int     `json:"id"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Balance  float32 `json:"balance"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   string    `json:"order_id"`
	Accrual   float32   `json:"accrual"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   string    `json:"order_id"`
	Sum       float32   `json:"sum"`
	CreatedAt time.Time `json:"created_at"`
}

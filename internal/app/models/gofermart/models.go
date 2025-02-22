package gofermart

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	OrderID   int64  `json:"order_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

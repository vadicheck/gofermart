package accrualservice

type GetOrderResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual,omitempty"`
}

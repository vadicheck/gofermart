package transactions

type TransactionResponse struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAT string  `json:"processed_at"`
}

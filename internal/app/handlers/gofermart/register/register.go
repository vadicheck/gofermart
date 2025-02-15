package register

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
)

func New(ctx context.Context, repository gophermart.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request body")
	}
}

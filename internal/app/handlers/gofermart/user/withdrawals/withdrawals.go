package withdrawals

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart"
	transactions2 "github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart/transactions"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type storage interface {
	GetTransactionsByUserID(ctx context.Context, userID int) ([]models.Transaction, error)
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage storage,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		transactions, err := storage.GetTransactionsByUserID(ctx, userID)

		if err != nil {
			logger.Error(fmt.Errorf("failed to get transactions. userID: %d, err: %w", userID, err))
			response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "Failed to get withdrawals history"), logger)
			return
		}

		if len(transactions) == 0 {
			response.ResponseError(w, gofermart.NewError(http.StatusNoContent, "No withdrawals"), logger)
			return
		}

		responseOrders := make([]transactions2.TransactionResponse, 0)

		for _, transaction := range transactions {
			responseOrders = append(responseOrders, transactions2.TransactionResponse{
				Order:       strconv.FormatInt(transaction.OrderID, 10),
				Sum:         float32(transaction.Sum),
				ProcessedAT: transaction.CreatedAt.Format(time.RFC3339),
			})
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, responseOrders); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

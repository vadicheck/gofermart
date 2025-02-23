package withdrawals

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
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
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		transactions, err := storage.GetTransactionsByUserID(ctx, userID)

		if err != nil {
			logger.Error(fmt.Errorf("failed to get transactions. userID: %d, err: %w", userID, err))
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Failed to get withdrawals history"), logger)
			return
		}

		if len(transactions) == 0 {
			response.Error(w, resmodels.NewError(http.StatusNoContent, "No withdrawals"), logger)
			return
		}

		responseOrders := make([]resmodels.TransactionResponse, 0)

		for _, transaction := range transactions {
			responseOrders = append(responseOrders, resmodels.TransactionResponse{
				Order:       transaction.OrderID,
				Sum:         transaction.Sum,
				ProcessedAT: transaction.CreatedAt.Format(time.RFC3339),
			})
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, responseOrders); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

package orders

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	"github.com/vadicheck/gofermart/pkg/logger"
)

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage gophermart.Gophermart,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		orders, err := storage.GetOrders(ctx, userID)

		if err != nil {
			logger.Error(fmt.Errorf("failed to get orders. userID: %d, err: %w", userID, err))
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Failed to get orders"), logger)
			return
		}

		if len(orders) == 0 {
			response.Error(w, resmodels.NewError(http.StatusNoContent, "No orders found"), logger)
			return
		}

		responseOrders := make([]resmodels.OrderResponse, 0)

		for _, order := range orders {
			responseOrders = append(responseOrders, resmodels.OrderResponse{
				Number:     order.OrderID,
				Status:     order.Status,
				Accrual:    order.Accrual,
				UploadedAt: order.CreatedAt.Format(time.RFC3339),
			})
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, responseOrders); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

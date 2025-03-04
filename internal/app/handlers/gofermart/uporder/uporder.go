package uporder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	apppstorage "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/luhn"
)

type orderStorage interface {
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	CreateOrder(ctx context.Context, orderID string, userID int) (int, error)
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage orderStorage,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, "Invalid request"), logger)
			logger.Error(err)
			return
		}
		defer r.Body.Close()

		orderID := string(body)

		luhOrderID, err := strconv.Atoi(orderID)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusUnprocessableEntity, "Invalid order number"), logger)
			return
		}

		if !luhn.Valid(luhOrderID) {
			response.Error(w, resmodels.NewError(http.StatusUnprocessableEntity, "Invalid order number (luhn)"), logger)
			return
		}

		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		order, err := storage.GetOrderByID(ctx, orderID)
		if errors.Is(err, apppstorage.ErrOrderNotFound) {
			if _, err = storage.CreateOrder(ctx, orderID, userID); err != nil {
				response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Invalid create order"), logger)
				logger.Error(err)
				return
			}

			if responseErr := response.RespondWithJSON(w, http.StatusAccepted, nil); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}

			return
		} else if err != nil {
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Invalid create order"), logger)
			logger.Error(err)
			return
		}

		if order.UserID == userID {
			if responseErr := response.RespondWithJSON(w, http.StatusOK, nil); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}
			return
		}

		if responseErr := response.RespondWithJSON(w, http.StatusConflict, nil); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

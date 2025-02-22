package uporder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	pstorage "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/luhn"
)

type orderService interface {
	CreateOrder(ctx context.Context, orderID int, userID int, logger logger.LogClient) (int, error)
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage gophermart.Gophermart,
	orderService orderService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusBadRequest, "Invalid request"), logger)
			logger.Error(err)
			return
		}
		defer r.Body.Close()

		orderID, err := strconv.Atoi(string(body))
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusUnprocessableEntity, "Invalid order number"), logger)
			return
		}

		if !luhn.Valid(orderID) {
			response.ResponseError(w, gofermart.NewError(http.StatusUnprocessableEntity, "Invalid order number (luhn)"), logger)
			return
		}

		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusBadRequest, "Invalid user number"), logger)
			return
		}

		order, err := storage.GetOrderByID(ctx, orderID)
		if errors.Is(err, pstorage.ErrOrderNotFound) {
			if _, err = orderService.CreateOrder(ctx, orderID, userID, logger); err != nil {
				response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "Invalid create order"), logger)
				logger.Error(err)
				return
			}

			if responseErr := response.RespondWithJSON(w, http.StatusAccepted, nil); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}

			return
		} else if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "Invalid create order"), logger)
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

package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	"github.com/vadicheck/gofermart/internal/app/services"
	appstorage "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/luhn"
)

type balanceService interface {
	Withdraw(ctx context.Context, user models.User, orderID string, sum float32) error
}

type storage interface {
	GetUserByID(ctx context.Context, id int) (models.User, error)
}

type Request struct {
	OrderID string  `json:"order" validate:"required"`
	Sum     float32 `json:"sum" validate:"required"`
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage storage,
	validator validator.Validate,
	balanceService balanceService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request

		dec := json.NewDecoder(r.Body)
		if decodeErr := dec.Decode(&request); decodeErr != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, "invalid JSON body"), logger)
			logger.Error(decodeErr)
			return
		}

		err := validator.Struct(request)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, err.Error()), logger)
			return
		}

		luhOrderID, err := strconv.Atoi(request.OrderID)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, "invalid order number"), logger)
			return
		}

		if !luhn.Valid(luhOrderID) {
			response.Error(w, resmodels.NewError(http.StatusUnprocessableEntity, "invalid order number (luhn)"), logger)
			return
		}

		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "unauthorized"), logger)
			return
		}

		user, err := storage.GetUserByID(ctx, userID)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, "user not found"), logger)
			return
		}

		err = balanceService.Withdraw(ctx, user, request.OrderID, request.Sum)
		if err != nil {
			if errors.Is(err, services.ErrInsufficientBalance) {
				response.Error(w, resmodels.NewError(http.StatusPaymentRequired, "insufficient funds"), logger)
			} else if errors.Is(err, appstorage.ErrOrderTransactionAlreadyExists) {
				response.Error(w, resmodels.NewError(http.StatusUnprocessableEntity, "order has been processed"), logger)
			} else {
				response.Error(w, resmodels.NewError(http.StatusInternalServerError, "can't withdraw money"), logger)
			}
			return
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, nil); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

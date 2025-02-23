package currentbalance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart/users"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	storage2 "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Storage interface {
	GetUserByID(ctx context.Context, userID int) (models.User, error)
	GetTotalWithdrawn(ctx context.Context, userID int) (float32, error)
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	storage Storage,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.Header.Get(string(constants.XUserID)))
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		user, err := storage.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, storage2.ErrUserNotFound) {
				response.ResponseError(w, gofermart.NewError(http.StatusNotFound, "user not found"), logger)
			} else {
				response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "can't find user"), logger)
			}
			return
		}

		withdrawn, err := storage.GetTotalWithdrawn(ctx, userID)
		if err != nil {
			logger.Error(fmt.Errorf("failed to get total withdrawn. userID: %d, err: %w", userID, err))
			response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "Failed to get total withdrawn"), logger)
			return
		}

		responseBalance := users.BalanceResponse{
			Current:   float32(user.Balance),
			Withdrawn: withdrawn,
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, responseBalance); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

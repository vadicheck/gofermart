package currentbalance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	models "github.com/vadicheck/gofermart/internal/app/models/gofermart"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	appstorage "github.com/vadicheck/gofermart/internal/app/storage"
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
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
			return
		}

		user, err := storage.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, appstorage.ErrUserNotFound) {
				response.Error(w, resmodels.NewError(http.StatusNotFound, "User not found"), logger)
			} else {
				response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Can't find user"), logger)
			}
			return
		}

		withdrawn, err := storage.GetTotalWithdrawn(ctx, userID)
		if err != nil {
			logger.Error(fmt.Errorf("failed to get total withdrawn. userID: %d, err: %w", userID, err))
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Failed to get total withdrawn"), logger)
			return
		}

		responseBalance := resmodels.BalanceResponse{
			Current:   user.Balance,
			Withdrawn: withdrawn,
		}

		if responseErr := response.RespondWithJSON(w, http.StatusOK, responseBalance); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

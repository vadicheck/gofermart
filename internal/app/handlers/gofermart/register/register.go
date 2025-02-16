package register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	httperr "github.com/vadicheck/gofermart/internal/app/httpserver/error"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type RegRequest struct {
	Login    string `json:"login" validate:"required,alphanum,max=140"`
	Password string `json:"password" validate:"required,max=140"`
}

type regService interface {
	CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int64, error)
}

func New(
	ctx context.Context,
	logger logger.LogClient,
	validator validator.Validate,
	regService regService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RegRequest

		dec := json.NewDecoder(r.Body)
		if decodeErr := dec.Decode(&request); decodeErr != nil {
			respError := gofermart.NewError(http.StatusBadRequest, "Invalid JSON body")
			if responseErr := httperr.RespondWithJSON(w, respError.Code, respError); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}
			logger.Error(decodeErr)
			return
		}

		err := validator.Struct(request)
		if err != nil {
			respError := gofermart.NewError(http.StatusBadRequest, err.Error())
			if responseErr := httperr.RespondWithJSON(w, respError.Code, respError); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}
			return
		}

		if _, err = regService.CreateUser(ctx, request.Login, request.Password, logger); err != nil {
			respError := gofermart.NewError(http.StatusInternalServerError, "Error creating user")

			if errors.Is(err, storage.ErrLoginAlreadyExists) {
				respError = gofermart.NewError(http.StatusConflict, "Login already exists")
			}

			if responseErr := httperr.RespondWithJSON(w, respError.Code, respError); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}

			logger.Error(err)
		}
	}
}

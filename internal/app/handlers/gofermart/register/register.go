package register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/httpserver/models/gofermart"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	"github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/secure/jwt"
)

type RegRequest struct {
	Login    string `json:"login" validate:"required,alphanum,max=140"`
	Password string `json:"password" validate:"required,max=140"`
}

type regService interface {
	CreateUser(ctx context.Context, login, password string, logger logger.LogClient) (int, error)
}

func New(
	ctx context.Context,
	jwtConfig config.JwtConfig,
	logger logger.LogClient,
	validator validator.Validate,
	regService regService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request RegRequest

		dec := json.NewDecoder(r.Body)
		if decodeErr := dec.Decode(&request); decodeErr != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusBadRequest, "Invalid JSON body"), logger)
			logger.Error(decodeErr)
			return
		}

		err := validator.Struct(request)
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusBadRequest, err.Error()), logger)
			return
		}

		userID, err := regService.CreateUser(ctx, request.Login, request.Password, logger)
		if err != nil {
			respError := gofermart.NewError(http.StatusInternalServerError, "Error creating user")

			if errors.Is(err, storage.ErrLoginAlreadyExists) {
				respError = gofermart.NewError(http.StatusConflict, "Login already exists")
			}

			if responseErr := response.RespondWithJSON(w, respError.Code, respError); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}
			logger.Error(err)
			return
		}

		token, err := jwt.BuildJWTString(jwtConfig.JwtSecret, jwtConfig.JwtTokenExpire, userID)
		if err != nil {
			response.ResponseError(w, gofermart.NewError(http.StatusInternalServerError, "can't build jwt token"), logger)
			logger.Error(fmt.Errorf("can't build jwt token: %w", err))
			return
		}

		w.Header().Set("Authorization", token)

		if responseErr := response.RespondWithJSON(w, http.StatusCreated, nil); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

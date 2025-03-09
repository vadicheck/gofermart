package register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	appstorage "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	pass "github.com/vadicheck/gofermart/pkg/password"
	"github.com/vadicheck/gofermart/pkg/secure/jwt"
)

type regRequest struct {
	Login    string `json:"login" validate:"required,alphanum,max=140"`
	Password string `json:"password" validate:"required,max=140"`
}

type registerStorage interface {
	CreateUser(ctx context.Context, login, password string) (int, error)
}

func New(
	ctx context.Context,
	jwtConfig config.JwtConfig,
	logger logger.LogClient,
	storage registerStorage,
	validator *validator.Validate,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request regRequest

		dec := json.NewDecoder(r.Body)
		if decodeErr := dec.Decode(&request); decodeErr != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, "Invalid JSON body"), logger)
			logger.Error(decodeErr)
			return
		}

		err := validator.Struct(request)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusBadRequest, err.Error()), logger)
			return
		}

		hashPassword, err := pass.HashPassword(request.Password)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Error password"), logger)
			return
		}

		userID, err := storage.CreateUser(ctx, request.Login, hashPassword)
		if err != nil {
			respError := resmodels.NewError(http.StatusInternalServerError, "Error creating user")

			if errors.Is(err, appstorage.ErrLoginAlreadyExists) {
				respError = resmodels.NewError(http.StatusConflict, "Login already exists")
			}

			if responseErr := response.RespondWithJSON(w, respError.Code, respError); responseErr != nil {
				logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
			}
			logger.Error(err)
			return
		}

		token, err := jwt.BuildJWTString(jwtConfig.JwtSecret, jwtConfig.JwtTokenExpire, userID)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "can't build jwt token"), logger)
			logger.Error(fmt.Errorf("can't build jwt token: %w", err))
			return
		}

		w.Header().Set("Authorization", token)

		if responseErr := response.RespondWithJSON(w, http.StatusOK, nil); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

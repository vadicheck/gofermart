package login

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
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	appstorage "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/password"
	"github.com/vadicheck/gofermart/pkg/secure/jwt"
)

type Request struct {
	Login    string `json:"login" validate:"required,alphanum,max=140"`
	Password string `json:"password" validate:"required,max=140"`
}

func New(
	ctx context.Context,
	jwtConfig config.JwtConfig,
	logger logger.LogClient,
	validator *validator.Validate,
	storage gophermart.Gophermart,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request

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

		user, err := storage.GetUserByLogin(ctx, request.Login)
		if err != nil {
			if errors.Is(err, appstorage.ErrUserNotFound) {
				response.Error(w, resmodels.NewError(http.StatusNotFound, "User not found"), logger)
			} else {
				response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Can't find user"), logger)
				logger.Error(fmt.Errorf("can't find user: %w", err))
			}
			return
		}

		if !password.CheckPasswordHash(request.Password, user.Password) {
			response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Username or password is incorrect"), logger)
			return
		}

		token, err := jwt.BuildJWTString(jwtConfig.JwtSecret, jwtConfig.JwtTokenExpire, user.ID)
		if err != nil {
			response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Can't build jwt token"), logger)
			logger.Error(fmt.Errorf("can't build jwt token: %w", err))
			return
		}

		w.Header().Set("Authorization", token)

		if responseErr := response.RespondWithJSON(w, http.StatusOK, nil); responseErr != nil {
			logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
		}
	}
}

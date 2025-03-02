package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/httpserver/response"
	resmodels "github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	"github.com/vadicheck/gofermart/pkg/logger"
	securejwt "github.com/vadicheck/gofermart/pkg/secure/jwt"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

const loginURL = "/api/user/login"
const registerURL = "/api/user/register"

func New(logger logger.LogClient, jwtConfig config.JwtConfig) func(next http.Handler) http.Handler {
	logger.Info("jwt middleware enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			jwtToken := r.Header.Get("Authorization")

			if r.URL.String() == loginURL || r.URL.String() == registerURL {
				next.ServeHTTP(w, r)
				return
			}

			if jwtToken == "" {
				response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
				return
			}

			decodedJwtToken, err := securejwt.DecodeJwtToken(jwtToken, jwtConfig.JwtSecret)
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
					return
				}

				logger.Error(fmt.Errorf("can't decode jwt token: %w", err))
				response.Error(w, resmodels.NewError(http.StatusInternalServerError, "Auth error"), logger)
				return
			}

			if decodedJwtToken.UserID == 0 {
				logger.Error(fmt.Errorf("user_id is absent in jwt"))
				response.Error(w, resmodels.NewError(http.StatusUnauthorized, "Unauthorized"), logger)
				return
			}

			r.Header.Set(string(constants.XUserID), strconv.Itoa(decodedJwtToken.UserID))
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

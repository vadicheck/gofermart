package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/login"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/orders"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/register"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/uporder"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/user/balance/currentbalance"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/user/balance/withdraw"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/user/withdrawals"
	"github.com/vadicheck/gofermart/internal/app/middleware/gzip"
	"github.com/vadicheck/gofermart/internal/app/middleware/jwt"
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	"github.com/vadicheck/gofermart/internal/app/services/gofermart/balance"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type HTTPServer struct {
	router        *chi.Mux
	serverAddress string
}

func (hs *HTTPServer) Run(logger logger.LogClient) (*http.Server, error) {
	server := &http.Server{
		Addr:         hs.serverAddress,
		Handler:      hs.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Info(fmt.Sprintf("server starting: %s", hs.serverAddress))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(fmt.Errorf("error starting server: %w", err))
		}
	}()

	return server, nil
}

func New(
	ctx context.Context,
	cfg *config.Config,
	logger logger.LogClient,
	storage gophermart.Gophermart,
	validator *validator.Validate,
) *HTTPServer {
	r := chi.NewRouter()

	r.Use(gzip.New())
	r.Use(jwt.New(logger, cfg.Jwt))

	balanceService := balance.New(storage, logger)

	r.Post("/api/user/register", register.New(
		ctx,
		cfg.Jwt,
		logger,
		storage,
		validator,
	))

	r.Post("/api/user/login", login.New(
		ctx,
		cfg.Jwt,
		logger,
		validator,
		storage,
	))

	r.Post("/api/user/balance/withdraw", withdraw.New(
		ctx,
		logger,
		storage,
		validator,
		balanceService,
	))

	r.Post("/api/user/orders", uporder.New(ctx, logger, storage))
	r.Get("/api/user/orders", orders.New(ctx, logger, storage))

	r.Get("/api/user/balance", currentbalance.New(ctx, logger, storage))
	r.Get("/api/user/withdrawals", withdrawals.New(ctx, logger, storage))

	return &HTTPServer{
		router:        r,
		serverAddress: cfg.HTTPAddress,
	}
}

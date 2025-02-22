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
	"github.com/vadicheck/gofermart/internal/app/middleware/jwt"
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	"github.com/vadicheck/gofermart/internal/app/services/gofermart/order"
	"github.com/vadicheck/gofermart/internal/app/services/gofermart/user"
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
	validator validator.Validate,
) *HTTPServer {
	r := chi.NewRouter()

	r.Use(jwt.New(logger, cfg.Jwt))

	userService := user.New(storage)
	orderService := order.New(storage)

	r.Post("/api/user/register", register.New(
		ctx,
		cfg.Jwt,
		logger,
		validator,
		userService,
	))

	r.Post("/api/user/login", login.New(
		ctx,
		cfg.Jwt,
		logger,
		validator,
		storage,
	))

	r.Post("/api/user/orders", uporder.New(ctx, logger, storage, orderService))
	r.Get("/api/user/orders", orders.New(ctx, logger, storage))

	return &HTTPServer{
		router:        r,
		serverAddress: cfg.HTTPAddress,
	}
}

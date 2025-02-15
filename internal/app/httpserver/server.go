package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/handlers/gofermart/register"
	"github.com/vadicheck/gofermart/internal/app/repository/gophermart"
	"github.com/vadicheck/gofermart/pkg/logger"
	"github.com/vadicheck/gofermart/pkg/logger/sl"
)

type HTTPServer struct {
	router        *chi.Mux
	serverAddress string
}

func (hs *HTTPServer) Run() (*http.Server, error) {
	server := &http.Server{
		Addr:         hs.serverAddress,
		Handler:      hs.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	slog.Info(fmt.Sprintf("Server starting: %s", hs.serverAddress))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", sl.Err(err))
		}
	}()

	return server, nil
}

func New(
	ctx context.Context,
	cfg *config.Config,
	logger logger.LogClient,
	storage gophermart.Gophermart,
) *HTTPServer {
	r := chi.NewRouter()

	r.Post("/api/user/register", register.New(ctx, storage))

	return &HTTPServer{
		router:        r,
		serverAddress: cfg.HTTPAddress,
	}
}

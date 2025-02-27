package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/httpserver"
	"github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	appsync "github.com/vadicheck/gofermart/internal/app/sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("config read err %w", err))
	}

	logger, err := log.New(*cfg)
	if err != nil {
		panic(err)
	}

	storage, err := postgres.New(cfg, logger)
	if err != nil {
		panic(err)
	}

	httpApp := httpserver.New(
		ctx,
		cfg,
		logger,
		storage,
		*validator.New(),
	)

	syncApp := appsync.New(cfg.AccrualSystemAddress, storage, logger)

	httpServer, err := httpApp.Run(logger)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	err = syncApp.Run(ctx, &wg)
	if err != nil {
		panic(err)
	}

	logger.Info("app is ready")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-exit:
		logger.Info(fmt.Sprintf("signal.Notify: %v\n\n", v))
	case done := <-ctx.Done():
		logger.Info(fmt.Sprintf("ctx.Done: %v", done))
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Info(err.Error())
	}

	cancel()
	wg.Wait()

	logger.Info("Server Exited Properly")
}

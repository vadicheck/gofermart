package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-playground/validator/v10"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/httpserver"
	appLog "github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	storSync "github.com/vadicheck/gofermart/internal/app/storage/sync"
	appsync "github.com/vadicheck/gofermart/internal/app/sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Panic(fmt.Errorf("config read err %w", err))
	}

	logger, err := appLog.New(*cfg)
	if err != nil {
		log.Panic(err)
	}

	storage, err := postgres.New(cfg, logger)
	if err != nil {
		logger.Panic(err)
	}

	httpApp := httpserver.New(
		ctx,
		cfg,
		logger,
		storage,
		validator.New(),
	)

	syncStorage, err := storSync.New(storage, logger)
	if err != nil {
		logger.Panic(err)
	}

	syncApp := appsync.New(cfg.AccrualSystemAddress, storage, syncStorage, logger)

	httpServer, err := httpApp.Run(logger)
	if err != nil {
		logger.Panic(err)
	}

	var wg sync.WaitGroup

	err = syncApp.Run(ctx, &wg)
	if err != nil {
		logger.Panic(err)
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

package main

import (
	"context"
	"fmt"

	"github.com/vadicheck/gofermart/internal/app/config"
	pkglog "github.com/vadicheck/gofermart/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("config read err %w", err))
	}

	fmt.Println(cfg)

	logger := pkglog.NewLogger(cfg.Env)

	logger.Info("INFO")

	ctx.Done()
}

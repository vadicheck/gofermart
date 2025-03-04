package log

import (
	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/pkg/logger"
)

func New(cfg config.Config) (logger.LogClient, error) {
	log, err := logger.New(logger.Options{
		ConsoleOptions: logger.ConsoleOptions{
			Level: cfg.LogLevel,
		},
	})
	if err != nil {
		return nil, err
	}

	return log, err
}

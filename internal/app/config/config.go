package config

import (
	"flag"
	"os"
)

type Config struct {
	HTTPAddress          string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.HTTPAddress, "a", "localhost:8082", "HTTP server startup address")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&cfg.LogLevel, "ll", "", "level of logs")

	flag.Parse()

	if httpAddress := os.Getenv("RUN_ADDRESS"); httpAddress != "" {
		cfg.HTTPAddress = httpAddress
	}

	if databaseDSN := os.Getenv("DATABASE_URI"); databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	if accrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSystemAddress != "" {
		cfg.AccrualSystemAddress = accrualSystemAddress
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	return &cfg, nil
}

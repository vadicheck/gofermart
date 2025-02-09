package config

import (
	"flag"
	"os"
)

type Config struct {
	HTTPAddress          string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Env                  string `env:"env" env-default:"local"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.HTTPAddress, "a", "localhost:8082", "HTTP server startup address")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&cfg.Env, "e", "local", "environment[local|prod]")

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

	if env := os.Getenv("ENV"); env != "" {
		cfg.Env = env
	}

	return &cfg, nil
}

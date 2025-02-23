package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	AccrualAddress       string `env:"ACCRUAL_ADDRESS"`
	HTTPAddress          string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
	MigrationsFilePath   string `env:"MIGRATIONS_FILE_PATH"`
	Jwt                  JwtConfig
}

type JwtConfig struct {
	JwtSecret      string `env:"JWT_SECRET"`
	JwtTokenExpire time.Duration
}

func NewConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.AccrualAddress, "aa", "http://localhost:8080", "Accrual server address")
	flag.StringVar(&cfg.HTTPAddress, "a", "localhost:8082", "HTTP server startup address")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&cfg.LogLevel, "ll", "", "level of logs")
	flag.StringVar(&cfg.MigrationsFilePath, "mp", "file://internal/app/migration/migrations", "migrations file path")

	flag.Parse()

	cfg.Jwt.JwtTokenExpire = time.Hour * 24

	if accrualAddress := os.Getenv("ACCRUAL_ADDRESS"); accrualAddress != "" {
		cfg.AccrualAddress = accrualAddress
	}

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

	if migrationsFilePath := os.Getenv("MIGRATIONS_FILE_PATH"); migrationsFilePath != "" {
		cfg.MigrationsFilePath = migrationsFilePath
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.Jwt.JwtSecret = jwtSecret
	} else {
		cfg.Jwt.JwtSecret = "secretkey"
	}

	return &cfg, nil
}

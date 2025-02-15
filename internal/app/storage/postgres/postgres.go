package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/migration"
	"github.com/vadicheck/gofermart/pkg/logger"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config, logger logger.LogClient) (*Storage, error) {
	err := migration.ExecuteMigrations(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	m, err := migrate.New("file://internal/app/migration/migrations", cfg.DatabaseDSN)
	if err != nil {
		log.Panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No migrations needed")
		} else {
			log.Panic(err)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateUser(ctx context.Context, userID int64) (int64, error) {
	return userID, nil
}

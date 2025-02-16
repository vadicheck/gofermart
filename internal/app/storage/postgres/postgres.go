package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/migration"
	"github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/pkg/logger"
	pass "github.com/vadicheck/gofermart/pkg/password"
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

	return &Storage{db: db}, nil
}

func (s *Storage) CreateUser(
	ctx context.Context,
	login, password string,
	logger logger.LogClient,
) (int64, error) {
	const op = "storage.postgres.CreateUser"
	const insertURL = "INSERT INTO public.users (login, password) VALUES ($1,$2) RETURNING id"

	stmt, err := s.db.Prepare(insertURL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	hashPassword, err := pass.HashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64

	err = stmt.QueryRowContext(ctx, login, hashPassword).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, storage.ErrLoginAlreadyExists
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

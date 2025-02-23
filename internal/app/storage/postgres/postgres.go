package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/migration"
	"github.com/vadicheck/gofermart/internal/app/models/gofermart"
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
) (int, error) {
	const op = "storage.postgres.CreateUser"
	const insertSQL = "INSERT INTO public.users (login, password) VALUES ($1,$2) RETURNING id"

	stmt, err := s.db.Prepare(insertSQL)
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

	var id int

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

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (gofermart.User, error) {
	const op = "storage.postgres.GetUserByLogin"
	const selectSQL = "SELECT id, login, password FROM users WHERE login = $1"

	var user gofermart.User

	row := s.db.QueryRowContext(ctx, selectSQL, login)

	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return user, storage.ErrUserNotFound
	} else if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) DeleteAllUsers(ctx context.Context, logger logger.LogClient) error {
	const op = "storage.postgres.DeleteAllUsers"
	const deleteSQL = "DELETE FROM users WHERE id <> 0"

	stmt, err := s.db.Prepare(deleteSQL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	_, err = stmt.ExecContext(ctx)

	return err
}

func (s *Storage) GetOrders(ctx context.Context, userID int, logger logger.LogClient) ([]gofermart.Order, error) {
	const op = "storage.postgres.GetOrders"
	const selectSQL = "SELECT id, user_id, order_id, accrual, status, created_at, updated_at FROM orders WHERE user_id = $1"

	rows, err := s.db.QueryContext(ctx, selectSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders [%s]: %w", op, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(fmt.Errorf("rows close error: %w", err))
		}
	}()

	var orders []gofermart.Order

	for rows.Next() {
		var order gofermart.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderID,
			&order.Accrual,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row[%s]: %w", op, err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error encountered during rows iteration [%s]: %w", op, err)
	}

	return orders, nil
}

func (s *Storage) CreateOrder(
	ctx context.Context,
	orderID, userID int,
	logger logger.LogClient,
) (int, error) {
	const op = "storage.postgres.CreateOrder"
	const insertSQL = "INSERT INTO public.orders (user_id, order_id, status) VALUES ($1,$2,$3) RETURNING id"

	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	var id int

	err = stmt.QueryRowContext(ctx, userID, orderID, constants.StatusNew).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, storage.ErrOrderAlreadyExists
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, orderID int) (gofermart.Order, error) {
	const op = "storage.postgres.GetOrderByID"
	const selectSQL = "SELECT id, user_id, order_id, accrual, status, created_at, updated_at FROM orders WHERE order_id = $1"

	var order gofermart.Order

	row := s.db.QueryRowContext(ctx, selectSQL, orderID)

	err := row.Scan(&order.ID, &order.UserID, &order.OrderID, &order.Accrual, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return order, storage.ErrOrderNotFound
	} else if err != nil {
		return order, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (s *Storage) GetOrdersIdsByStatus(
	ctx context.Context,
	statuses []constants.OrderStatus,
	logger logger.LogClient,
) ([]int, error) {
	const op = "storage.postgres.GetOrdersIdsByStatus"
	const selectSQL = "SELECT order_id FROM orders WHERE status = ANY($1::order_status[])"

	rows, err := s.db.QueryContext(ctx, selectSQL, pq.Array(statuses))

	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status [%s] [%s]: %w", statuses, op, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(fmt.Errorf("rows close error: %w", err))
		}
	}()

	var ordersIds []int

	for rows.Next() {
		var orderID int
		if err := rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("failed to scan row[%s]: %w", op, err)
		}
		ordersIds = append(ordersIds, orderID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error encountered during rows iteration [%s]: %w", op, err)
	}

	return ordersIds, nil
}

func (s *Storage) UpdateOrder(
	ctx context.Context,
	orderID int,
	newStatus constants.OrderStatus,
	accrual int,
	logger logger.LogClient,
) error {
	const op = "storage.postgres.UpdateOrder"
	const updateSQL = "UPDATE orders SET status = $1, accrual = $2 WHERE order_id = $3"

	stmt, err := s.db.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	_, err = stmt.ExecContext(ctx, newStatus, accrual, orderID)

	return err
}

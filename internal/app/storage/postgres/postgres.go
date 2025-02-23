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
)

type Storage struct {
	logger logger.LogClient
	db     *sql.DB
}

func New(cfg *config.Config, logger logger.LogClient) (*Storage, error) {
	if err := migration.ExecuteMigrations(cfg, logger); err != nil {
		logger.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &Storage{
		db:     db,
		logger: logger,
	}, nil
}

func (s *Storage) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction: %w", err)
	}

	return tx, nil
}

func (s *Storage) CreateUser(ctx context.Context, login, password string) (int, error) {
	const op = "storage.postgres.CreateUser"
	const insertSQL = "INSERT INTO public.users (login, password) VALUES ($1,$2) RETURNING id"

	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err = stmt.Close(); err != nil {
			s.logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	var id int

	err = stmt.QueryRowContext(ctx, login, password).Scan(&id)
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
	const selectSQL = "SELECT id, login, password, balance FROM users WHERE login = $1"

	row := s.db.QueryRowContext(ctx, selectSQL, login)

	return s.fillUser(row, "storage.postgres.GetUserByLogin")
}

func (s *Storage) GetUserByID(ctx context.Context, userID int) (gofermart.User, error) {
	const selectSQL = "SELECT id, login, password, balance FROM users WHERE id = $1"

	row := s.db.QueryRowContext(ctx, selectSQL, userID)

	return s.fillUser(row, "storage.postgres.GetUserByID")
}

func (s *Storage) ChangeUserBalance(ctx context.Context, userID int, balance float32) error {
	const op = "storage.postgres.ChangeUserBalance"
	const updateSQL = "UPDATE users SET balance = $1 WHERE id = $2"

	stmt, err := s.db.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	_, err = stmt.ExecContext(ctx, balance, userID)

	return err
}

func (s *Storage) GetOrders(ctx context.Context, userID int) ([]gofermart.Order, error) {
	const op = "storage.postgres.GetOrders"
	const selectSQL = "SELECT id, user_id, order_id, accrual, status, created_at, updated_at FROM orders WHERE user_id = $1"

	rows, err := s.db.QueryContext(ctx, selectSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders [%s]: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(fmt.Errorf("rows close error: %w", err))
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

func (s *Storage) CreateOrder(ctx context.Context, orderID string, userID int) (int, error) {
	const op = "storage.postgres.CreateOrder"
	const insertSQL = "INSERT INTO public.orders (user_id, order_id, status) VALUES ($1,$2,$3) RETURNING id"

	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.logger.Error(fmt.Errorf("prepare sql error: %w", err))
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

func (s *Storage) GetOrderByID(ctx context.Context, orderID string) (gofermart.Order, error) {
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

func (s *Storage) GetOrdersIdsByStatus(ctx context.Context, statuses []constants.OrderStatus) ([]string, error) {
	const op = "storage.postgres.GetOrdersIdsByStatus"
	const selectSQL = "SELECT order_id FROM orders WHERE status = ANY($1::order_status[])"

	rows, err := s.db.QueryContext(ctx, selectSQL, pq.Array(statuses))

	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status [%s] [%s]: %w", statuses, op, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(fmt.Errorf("rows close error: %w", err))
		}
	}()

	var ordersIds []string

	for rows.Next() {
		var orderID string
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
	orderID string,
	newStatus constants.OrderStatus,
	accrual float32,
) error {
	const op = "storage.postgres.UpdateOrder"
	const updateSQL = "UPDATE orders SET status = $1, accrual = $2 WHERE order_id = $3"

	stmt, err := s.db.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	_, err = stmt.ExecContext(ctx, newStatus, accrual, orderID)

	return err
}

func (s *Storage) CreateTransaction(
	ctx context.Context,
	userID int,
	orderID string,
	sum float32,
) error {
	const op = "storage.postgres.CreateTransaction"
	const insertSQL = "INSERT INTO public.transactions (user_id, order_id, sum) VALUES ($1,$2,$3) RETURNING id"

	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.logger.Error(fmt.Errorf("prepare sql error: %w", err))
		}
	}()

	var id int

	err = stmt.QueryRowContext(ctx, userID, orderID, sum).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return storage.ErrOrderTransactionAlreadyExists
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTransactionsByUserID(ctx context.Context, userID int) ([]gofermart.Transaction, error) {
	const op = "storage.postgres.GetTransactionsByUserID"
	const selectSQL = "SELECT id, user_id, order_id, sum, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, selectSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions [%s]: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error(fmt.Errorf("rows close error: %w", err))
		}
	}()

	var transactions []gofermart.Transaction

	for rows.Next() {
		var transaction gofermart.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.OrderID,
			&transaction.Sum,
			&transaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row[%s]: %w", op, err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error encountered during rows iteration [%s]: %w", op, err)
	}

	return transactions, nil
}

func (s *Storage) GetTotalWithdrawn(ctx context.Context, userID int) (float32, error) {
	const op = "storage.postgres.GetTotalWithdrawn"
	const selectSQL = "SELECT COALESCE(SUM(sum), 0) FROM transactions t WHERE t.user_id = $1"

	var sum float32

	row := s.db.QueryRowContext(ctx, selectSQL, userID)

	err := row.Scan(&sum)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return sum, nil
}

func (s *Storage) fillUser(row *sql.Row, op string) (gofermart.User, error) {
	var user gofermart.User

	err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return user, storage.ErrUserNotFound
	} else if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

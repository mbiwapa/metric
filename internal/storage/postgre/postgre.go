// Package postgre provides a PostgreSQL storage implementation for metrics.
// It includes methods for connecting to the database, performing CRUD operations on metrics,
// and handling retries with backoff strategies in case of failures.
package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/retry/backoff"
	"github.com/mbiwapa/metric/internal/storage"
)

// Storage structure for storage
// The Storage struct encapsulates a connection to a PostgreSQL database.
// It provides methods to interact with the database, such as creating, updating, and retrieving metrics.
type Storage struct {
	db *sql.DB // db is a pointer to the sql.DB instance representing the database connection.
}

// New returns a new Storage instance.
// It attempts to connect to the PostgreSQL database using the provided DSN (Data Source Name).
// If the connection is successful, it creates the 'metric' table if it does not already exist.
// The function retries the connection and table creation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - dsn: The Data Source Name for connecting to the PostgreSQL database.
//
// Returns:
// - A pointer to the Storage instance.
// - An error if the connection or table creation fails.
func New(dsn string) (*Storage, error) {
	const op = "storage.postgre.New"

	var storage Storage

	action := func(attempt uint) error {

		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS metric (
        name TEXT PRIMARY KEY,
        gauge DOUBLE PRECISION NOT NULL DEFAULT 0,
        counter BIGINT NOT NULL DEFAULT 0);`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_, err = stmt.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		storage.db = db
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &storage, nil
}

// Ping checks the connection to the database.
// It retries the ping operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the ping operation.
//
// Returns:
// - An error if the ping operation fails.
func (s *Storage) Ping(ctx context.Context) error {
	const op = "storage.postgre.Ping"
	action := func(attempt uint) error {

		err := s.db.PingContext(ctx)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Close closes the connection to the database.
func (s *Storage) Close() {
	s.db.Close()
}

// UpdateGauge saves the given Gauge metric to the database.
// It retries the update operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the update operation.
// - key: The name of the metric.
// - value: The value of the gauge metric.
//
// Returns:
// - An error if the update operation fails.
func (s *Storage) UpdateGauge(ctx context.Context, key string, value float64) error {
	const op = "storage.postgre.UpdateGauge"
	action := func(attempt uint) error {
		stmt, err := s.db.PrepareContext(ctx, `INSERT INTO metric (name, gauge) VALUES ($1,$2) ON CONFLICT (name) DO UPDATE SET gauge=$2`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_, err = stmt.ExecContext(ctx, key, value)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return fmt.Errorf("%s: %s", op, pgErr.Message)
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// UpdateCounter saves the given Counter metric to the database.
// It retrieves the current value of the counter metric, adds the new value to it, and updates the database.
// It retries the update operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the update operation.
// - key: The name of the metric.
// - value: The value of the counter metric.
//
// Returns:
// - An error if the update operation fails.
func (s *Storage) UpdateCounter(ctx context.Context, key string, value int64) error {
	const op = "storage.postgre.UpdateCounter"
	action := func(attempt uint) error {
		originalValue, err := s.GetMetric(ctx, format.Counter, key)
		if err != nil {
			if !errors.Is(err, storage.ErrMetricNotFound) {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		stmt, err := s.db.PrepareContext(ctx, `INSERT INTO metric (name, counter) VALUES ($1,$2) ON CONFLICT (name) DO UPDATE SET counter=$2`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if originalValue != "" {
			val, errStr := strconv.ParseInt(originalValue, 0, 64)
			if errStr != nil {
				return fmt.Errorf("%s: %w", op, errStr)
			}
			value = value + val
		}
		_, err = stmt.ExecContext(ctx, key, value)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return fmt.Errorf("%s: %s", op, pgErr.Message)
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetAllMetrics returns all metrics of types gauge and counter from the database.
// It retries the retrieval operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the retrieval operation.
//
// Returns:
// - A slice of slices containing gauge metrics.
// - A slice of slices containing counter metrics.
// - An error if the retrieval operation fails.
func (s *Storage) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	const op = "storage.postgre.GetAllMetrics"
	var gauges, counters [][]string
	var notFound bool
	action := func(attempt uint) error {
		stmt, err := s.db.PrepareContext(ctx, `SELECT (name, gauge, counter) FROM metric`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rows, err := stmt.QueryContext(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			notFound = true
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		gauges = make([][]string, 0, 30)
		counters = make([][]string, 0, 5)

		for rows.Next() {
			var name string
			var gauge float64
			var counter int64
			err = rows.Scan(&name, &gauge, &counter)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			if counter > 0 {
				counters = append(counters, []string{name, strconv.FormatInt(counter, 10)})
			}
			if gauge > 0 {
				gauges = append(gauges, []string{name, strconv.FormatFloat(gauge, 'f', -1, 64)})
			}
		}
		err = rows.Err()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	if notFound {
		return nil, nil, storage.ErrMetricNotFound
	}
	return gauges, counters, nil
}

// GetMetric returns a metric by key from the database.
// It retries the retrieval operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the retrieval operation.
// - typ: The type of the metric (gauge or counter).
// - key: The name of the metric.
//
// Returns:
// - The value of the metric as a string.
// - An error if the retrieval operation fails or the metric is not found.
func (s *Storage) GetMetric(ctx context.Context, typ string, key string) (string, error) {
	const op = "storage.postgre.GetMetric"
	var result string
	var notFound bool
	action := func(attempt uint) error {
		stmt, err := s.db.PrepareContext(ctx, `SELECT name, gauge, counter FROM metric WHERE name=$1`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		var name string
		var gauge float64
		var counter int64
		err = stmt.QueryRowContext(ctx, key).Scan(&name, &gauge, &counter)
		if errors.Is(err, sql.ErrNoRows) {
			notFound = true
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if typ == format.Gauge {
			result = strconv.FormatFloat(gauge, 'f', -1, 64)
			return nil
		}
		if typ == format.Counter {
			result = strconv.FormatInt(counter, 10)
			return nil
		}
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if notFound {
		return "", storage.ErrMetricNotFound
	}
	return result, nil
}

// UpdateBatch saves the given Gauge and Counter metrics to the PostgreSQL database in a batch operation.
// It retries the update operation up to 4 times with a backoff strategy in case of failure.
//
// Parameters:
// - ctx: The context for the update operation.
// - gauges: A slice of slices containing gauge metrics to be updated.
// - counters: A slice of slices containing counter metrics to be updated.
//
// Returns:
// - An error if the update operation fails.
func (s *Storage) UpdateBatch(ctx context.Context, gauges [][]string, counters [][]string) error {
	const op = "storage.postgre.UpdateBatch"

	action := func(attempt uint) error {

		tx, errTx := s.db.BeginTx(ctx, &sql.TxOptions{})
		if errTx != nil {
			return fmt.Errorf("%s: %w", op, errTx)
		}
		defer tx.Rollback()

		stmt, errTx := tx.PrepareContext(ctx, `INSERT INTO metric (name, gauge, counter) VALUES ($1,$2,$3) ON CONFLICT (name) DO UPDATE SET gauge=$2, counter=$3`)
		if errTx != nil {
			return fmt.Errorf("%s: %w", op, errTx)
		}
		for _, gauge := range gauges {
			newVal, err := strconv.ParseFloat(gauge[1], 64)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			_, err = stmt.ExecContext(ctx, gauge[0], newVal, 0)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
		for _, counter := range counters {
			stmtSelect, err := tx.PrepareContext(ctx, `SELECT name, counter FROM metric WHERE name=$1`)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			var name string
			var getCounter int64
			err = stmtSelect.QueryRowContext(ctx, counter[0]).Scan(&name, &getCounter)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, err)
				}
			}

			newVal, err := strconv.ParseInt(counter[1], 0, 64)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			newVal = newVal + getCounter
			_, err = stmt.ExecContext(ctx, counter[0], 0, newVal)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		errTx = tx.Commit()
		if errTx != nil {
			return fmt.Errorf("%s: %w", op, errTx)
		}

		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()),
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

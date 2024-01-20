package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/storage"
)

// Storage structure for storage
type Storage struct {
	db *sql.DB
}

// New return a new Storage instance.
func New(dsn string) (*Storage, error) {
	const op = "storage.postgre.New"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS metric (
		name TEXT PRIMARY KEY,
		gauge DOUBLE PRECISION NOT NULL DEFAULT 0,
		counter BIGINT NOT NULL DEFAULT 0);`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Ping checks connection to the database
func (s *Storage) Ping(ctx context.Context) error {
	const op = "storage.postgre.Ping"

	err := s.db.PingContext(ctx)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Close closes the connection to the database
func (s *Storage) Close() {
	s.db.Close()
}

// UpdateGauge saves the given Gauge metric to the memory.
func (s *Storage) UpdateGauge(ctx context.Context, key string, value float64) error {
	const op = "storage.postgre.UpdateGauge"
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

// UpdateCounter saves the given Counter metric to the memory.
func (s *Storage) UpdateCounter(ctx context.Context, key string, value int64) error {
	const op = "storage.postgre.UpdateCounter"
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
		val, err := strconv.ParseInt(originalValue, 0, 64)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
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

// GetAllMetrics Возвращает слайс метрик 2 типов gauge и counter
func (s *Storage) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	//TODO implement me
	const op = "storage.postgre.GetAllMetrics"

	stmt, err := s.db.PrepareContext(ctx, `SELECT (name, gauge, counter) FROM metric`)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("%s: %w", op, storage.ErrMetricNotFound)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	gauges := make([][]string, 0, 30)
	counters := make([][]string, 0, 5)

	for rows.Next() {
		var name string
		var gauge float64
		var counter int64
		err = rows.Scan(&name, &gauge, &counter)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
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
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return gauges, counters, nil
}

// GetMetric Возвращает метрику по ключу
func (s *Storage) GetMetric(ctx context.Context, typ string, key string) (string, error) {

	const op = "storage.postgre.GetMetric"
	stmt, err := s.db.PrepareContext(ctx, `SELECT name, gauge, counter FROM metric WHERE name=$1`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var name string
	var gauge float64
	var counter int64
	err = stmt.QueryRowContext(ctx, key).Scan(&name, &gauge, &counter)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrMetricNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if typ == format.Gauge {
		return strconv.FormatFloat(gauge, 'f', -1, 64), nil
	}
	if typ == format.Counter {
		return strconv.FormatInt(counter, 10), nil
	}
	return "", nil
}

// UpdateBatch saves the given Gauge and Counter metrics to the PG.
func (s *Storage) UpdateBatch(ctx context.Context, gauges [][]string, counters [][]string) error {
	const op = "storage.postgre.UpdateBatch"

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO metric (name, gauge, counter) VALUES ($1,$2,$3) ON CONFLICT (name) DO UPDATE SET gauge=$2, counter=$3`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
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
		originalValue, err := s.GetMetric(ctx, format.Counter, counter[0])
		if err != nil {
			if !errors.Is(err, storage.ErrMetricNotFound) {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		newVal, err := strconv.ParseInt(counter[1], 0, 64)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if originalValue != "" {
			val, err := strconv.ParseInt(originalValue, 0, 64)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			newVal = newVal + val
		}
		_, err = stmt.ExecContext(ctx, counter[0], 0, newVal)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

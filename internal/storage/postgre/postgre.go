package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
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
		gauge DOUBLE PRECISION,
		counter INTEGER);`)
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
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO metric (name, counter) VALUES ($1,$2) ON CONFLICT (name) DO UPDATE SET counter=$2`)
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
	stmt, err := s.db.PrepareContext(ctx, `SELECT $1 FROM metric WHERE name=$2`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var result string
	err = stmt.QueryRowContext(ctx, typ, key).Scan(result)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%s: %w", op, storage.ErrMetricNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

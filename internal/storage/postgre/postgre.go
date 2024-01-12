package postgre

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PgStorage structure for storage
type PgStorage struct {
	db *sql.DB
}

// New return a new PgStorage instance.
func New(dsn string) (*PgStorage, error) {
	const op = "storage.postgre.New"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// defer db.Close()

	return &PgStorage{db: db}, nil
}

// Ping checks connection to the database
func (s *PgStorage) Ping(ctx context.Context) error {
	const op = "storage.postgre.Ping"

	err := s.db.PingContext(ctx)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Close closes the connection to the database
func (s *PgStorage) Close() {
	s.db.Close()
}

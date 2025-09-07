package store

import (
	// Standard
	"database/sql"
	"fmt"

	// Register pgx driver for database/sql usage
	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenPostgres validates DSN, opens a pgx-backed *sql.DB and pings it.
// Caller is responsible for closing the returned *sql.DB when done.
func OpenPostgres(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set; provide via -db flag or env variable")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

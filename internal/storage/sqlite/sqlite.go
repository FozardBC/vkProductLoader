package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/storage"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrExecStmt    = fmt.Errorf("failed to exec statement")
	ErrPrepareStmt = fmt.Errorf("failed to prepare statement")
)

type Storage struct {
	log *slog.Logger
	db  *sql.DB
}

func New(log *slog.Logger, storagePath string) (*Storage, error) {

	if storagePath == "" {
		return nil, fmt.Errorf("storage path is empty")
	}

	log.Debug("Initializing SQLite storage", "path", storagePath)

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(
		`
	CREATE TABLE IF NOT EXISTS product_IDs(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_ID INTEGER NOT NULL,
    product_ID TEXT NOT NULL UNIQUE,
    created_At TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_productID ON product_IDs(product_ID);

		`)

	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrPrepareStmt, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrExecStmt, err)
	}

	return &Storage{log: log, db: db}, nil
}

func (s *Storage) SaveID(ProductID int, CategoryID int) error {

	stmt, err := s.db.Prepare(`INSERT INTO product_IDs(product_ID, category_ID) VALUES (?,?)`)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt.Exec(ProductID, CategoryID)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return storage.ErrProductIDExists
		}
		return fmt.Errorf("%w: %w", ErrExecStmt, err)
	}

	return nil
}

func (s *Storage) Close() error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		s.log.Debug("SQLite storage closed")
	}
	return nil
}

func (s *Storage) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database is not initialized")
	}

	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

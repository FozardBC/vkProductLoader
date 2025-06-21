package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/storage"
	"strings"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrExecStmt    = fmt.Errorf("failed to exec statement")
	ErrPrepareStmt = fmt.Errorf("failed to prepare statement")
)

type Storage struct {
	log *slog.Logger
	db  *sql.DB
}

var (
	productsTableName      = "product_IDs"
	idColumn               = "id"
	categoryIDColumn       = "category_ID"
	productIDColumn        = "product_ID"
	isVKPublishedColumn    = "is_VK_published"
	isAvitoPublishedColumn = "is_avito_published"
	createdAtColumn        = "created_at"
)

func New(log *slog.Logger, storagePath string) (*Storage, error) {

	if storagePath == "" {
		return nil, fmt.Errorf("storage path is empty")
	}

	log.Debug("Initializing SQLite storage", "path", storagePath)

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare( // CHANGE DEFAULT is_avito_published NAME, when add avito support
		`
	CREATE TABLE IF NOT EXISTS product_IDs(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_ID INTEGER NOT NULL,
    product_ID TEXT NOT NULL UNIQUE,TRUE
	is_VK_published BOOLEAN DEFAULT TRUE,
	is_avito_is_VK_published BOOLEAN DEFAULT FALSE, 
    created_at TEXT DEFAULT (datetime('now'))
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

	query := fmt.Sprintf(
		`INSERT INTO %s(%s, %s) VALUES (?, ?)`,
		productsTableName,
		productIDColumn,
		categoryIDColumn,
	)

	stmt, err := s.db.Prepare(query)
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

func (s *Storage) GetProdIDs(options *filters.Options) ([]int, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s `,
		productIDColumn,
		productsTableName,
	)

	query, args := filter(query, options)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrProductIDnotFound
		}
		return nil, fmt.Errorf("%w: %w", ErrExecStmt, err)
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var productID int
		if err := rows.Scan(&productID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		productIDs = append(productIDs, productID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return productIDs, nil
}

func (s *Storage) Delete(productID int) error {

	query := fmt.Sprintf(
		`DELETE FROM %s WHERE %s = ?`,
		productsTableName,
		productIDColumn,
	)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt.Exec(productID)
	if err != nil {
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

func filter(query string, options *filters.Options) (string, []interface{}) {
	var whereClauses []string
	var args []interface{}
	argNum := 1

	if options.CategoryID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", categoryIDColumn, argNum))
		args = append(args, *options.CategoryID)
		argNum++
	}
	if options.ProductID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", productIDColumn, argNum))
		args = append(args, *options.ProductID)
		argNum++
	}
	if options.Created_at != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", createdAtColumn, argNum))
		args = append(args, *options.Created_at)
		argNum++
	}

	// Добавляем WHERE если есть условия
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return query, args
}

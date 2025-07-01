package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/domain/models"
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

var (
	productsTable             = "products"
	productsIdColumn          = "id"
	productsTitleColumm       = "title"
	productsPriceColumn       = "price"
	productsDescripColumn     = "description"
	productsUcozLoadedColumn  = "ucoz_loaded"
	productsVKLoadedColumn    = "vk_loaded"
	productsAvitoLoadedColumn = "avito_loaded"

	productsIDkey = "product_id"

	productsPlatformIDsTable = "product_platforms_ids"
	productsPlatformIDsVK    = "vk_product_id"
	productsPlatformIDsUcoz  = "ucoz_product_id"
	productsPlatformIDsAvito = "avito_product_id"

	productImagesTable          = "product_images"
	productImagesTelegramFileID = "telegram_file_id"
	productImagesTelegramUrl    = "telegram_url"
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

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS products(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        price INTEGER NOT NULL,
        description TEXT,
        ucoz_loaded BOOLEAN NOT NULL DEFAULT FALSE,
        vk_loaded BOOLEAN NOT NULL DEFAULT FALSE,
        avito_loaded BOOLEAN NOT NULL DEFAULT FALSE, 
        created_at TEXT DEFAULT (datetime('now'))
    );`)
	if err != nil {
		return nil, fmt.Errorf("failed to create products table: %w", err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS product_platforms_ids(
        product_id INTEGER PRIMARY KEY,
        vk_product_id INTEGER,
        ucoz_product_id INTEGER,
        avito_product_id INTEGER,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
    );`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_platforms_ids table: %w", err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS product_images (
        product_id INTEGER PRIMARY KEY,
        telegram_file_id TEXT UNIQUE NOT NULL,
        telegram_url TEXT NOT NULL,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
    );`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_images table: %w", err)
	}

	return &Storage{log: log, db: db}, nil
}

func (s *Storage) Save(ctx context.Context, product *models.Product) (int64, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%w:%w", storage.ErrBeginTx, err)
	}

	defer tx.Rollback()

	query := fmt.Sprintf(
		`INSERT INTO %s(%s, %s, %s) VALUES (?, ?, ?) RETURNING id`,
		productsTable,
		productsTitleColumm,
		productsPriceColumn,
		productsDescripColumn,
	)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	result, err := stmt.Exec(product.Title, product.Price, product.Description)
	if err != nil {

		return 0, fmt.Errorf("%w: %w", ErrExecStmt, err)
	}

	id, err := result.LastInsertId()
	if err != nil {

		return 0, fmt.Errorf("%w:%w", storage.ErrReturnId, err)
	}

	stmt.Close()

	//2
	query2 := fmt.Sprintf("INSERT INTO %s(%s) VALUES (?)", productsPlatformIDsTable, productsIDkey)

	stmt2, err := tx.Prepare(query2)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt2.Exec(id)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, storage.ErrProductIDExists
		} //убрать в другой запрос ПРО СУЩЕСТВУЮЩИЙ АЙДИ ПРОДУКТА НА ПЛОЩАДКЕ
		return 0, fmt.Errorf("%w: %w", ErrExecStmt, err)
	}

	stmt2.Close()

	//3
	query3 := fmt.Sprintf("INSERT INTO %s(%s,%s,%s) VALUES (?,?,?)", productImagesTable, productsIDkey, productImagesTelegramFileID, productImagesTelegramUrl)

	stmt3, err := tx.Prepare(query3)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt3.Exec(id, product.TelegramFileID, product.TelegramUrlPic)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrExecStmt, err)
	}

	stmt3.Close()

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", storage.ErrCommitTx, err)
	}
	return id, nil
}

func (s *Storage) VkLoaded(productID int64, vkProductID int) error {
	tx, err := s.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("%w:%w", storage.ErrBeginTx, err)
	}

	query1 := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", productsTable, productsVKLoadedColumn, productsIdColumn)
	stmt, err := tx.Prepare(query1)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt.Exec(true, productID)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrExecStmt, err)
	}

	query2 := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", productsPlatformIDsTable, productsPlatformIDsVK, productsIDkey)

	stmt, err = tx.Prepare(query2)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}
	_, err = stmt.Exec(vkProductID, productID)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrExecStmt, err)
	}

	stmt.Close()

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w: %w", storage.ErrCommitTx, err)
	}

	defer tx.Rollback()

	return nil
}

func (s *Storage) Delete(ctx context.Context, productID int) error {
	return nil
}

func (s *Storage) GetProdIDs(options *filters.Options) ([]int, error) {
	return nil, nil
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

// func filter(query string, options *filters.Options) (string, []interface{}) {
// 	var whereClauses []string
// 	var args []interface{}
// 	argNum := 1

// 	if options.CategoryID != nil {
// 		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", categoryIDColumn, argNum))
// 		args = append(args, *options.CategoryID)
// 		argNum++
// 	}
// 	if options.ProductID != nil {
// 		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", productIDColumn, argNum))
// 		args = append(args, *options.ProductID)
// 		argNum++
// 	}
// 	if options.Created_at != nil {
// 		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", createdAtColumn, argNum))
// 		args = append(args, *options.Created_at)
// 		argNum++
// 	}

// 	// Добавляем WHERE если есть условия
// 	if len(whereClauses) > 0 {
// 		query += " WHERE " + strings.Join(whereClauses, " AND ")
// 	}

// 	return query, args
// }

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/storage"
	"strings"

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
	productsFtsTable = "products_fts"

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
        telegram_file_id TEXT NOT NULL,
        telegram_url TEXT NOT NULL,
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
    );`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_images table: %w", err)
	}

	_, err = db.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS products_fts 
	USING fts4(
	title,
	ucoz_loaded,
    vk_loaded,
    avito_loaded, 
    content='products',
    tokenize='simple');`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_images table: %w", err)
	}

	_, err = db.Exec(`
	CREATE TRIGGER IF NOT EXISTS products_ai AFTER INSERT ON products BEGIN
    INSERT INTO products_fts(rowid, title, ucoz_loaded, vk_loaded, avito_loaded)
    VALUES (new.id, new.title, new.ucoz_loaded, new.vk_loaded, new.avito_loaded);
	END;`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_images table: %w", err)
	}

	_, err = db.Exec(`
	CREATE TRIGGER IF NOT EXISTS products_au AFTER UPDATE OF name ON products BEGIN
    UPDATE products_fts SET
        title = new.title,
		ucoz_loaded = new.ucoz_loaded,
		vk_loaded = new.vk_loaded, 
		avito_loaded = new.avito_loaded
       
    WHERE rowid = old.id;
END;`)
	if err != nil {
		return nil, fmt.Errorf("failed to create product_images table: %w", err)
	}

	_, err = db.Exec(`
	CREATE TRIGGER IF NOT EXISTS products_ad AFTER DELETE ON products BEGIN
    DELETE FROM products_fts WHERE rowid = old.id;
END;`)
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

	product.Title = strings.ToLower(product.Title)

	query := fmt.Sprintf(
		`INSERT INTO %s(%s, %s, %s) VALUES (LOWER(?), ?, ?) RETURNING id`,
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

	_, err = stmt3.Exec(id, product.TelegramFileID, product.MainPictureURL)
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

func (s *Storage) UcozLoaded(productID int64, ucozProductID int) error {
	tx, err := s.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("%w:%w", storage.ErrBeginTx, err)
	}

	query1 := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", productsTable, productsUcozLoadedColumn, productsIdColumn)
	stmt, err := tx.Prepare(query1)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}

	_, err = stmt.Exec(true, productID)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrExecStmt, err)
	}

	query2 := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", productsPlatformIDsTable, productsPlatformIDsUcoz, productsIDkey)

	stmt, err = tx.Prepare(query2)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPrepareStmt, err)
	}
	_, err = stmt.Exec(ucozProductID, productID)
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

func (s *Storage) countProducts(searchQuery string) (int, error) {

	var count int

	searchQuery = strings.ToLower(searchQuery)

	query := fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM %s 
        WHERE %s MATCH ?
        AND (%s = TRUE OR %s = TRUE OR %s = TRUE)`,
		productsFtsTable,
		productsTitleColumm,
		productsUcozLoadedColumn,
		productsVKLoadedColumn,
		productsAvitoLoadedColumn,
	)

	err := s.db.QueryRow(query, searchQuery).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("no products found found")

		}

		return 0, fmt.Errorf("failed to make query:%w", err)
	}

	return count, nil

}

func (s *Storage) Search(ctx context.Context, searchQuery string, offset int, limit int) (products []*models.Product, count int, err error) {

	tx, err := s.db.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})
	if err != nil {
		return nil, 0, fmt.Errorf("%w:%w", storage.ErrBeginTx, err)
	}

	defer tx.Rollback()

	searchQuery = strings.ToLower(searchQuery)

	querySearchIDs := fmt.Sprintf(
		` 
		SELECT rowid 
        FROM %s
        WHERE %s MATCH ?
        AND (%s = TRUE OR %s = TRUE OR %s = TRUE)
		LIMIT %d OFFSET %d
`,
		productsFtsTable,
		productsTitleColumm,
		productsAvitoLoadedColumn, productsVKLoadedColumn, productsUcozLoadedColumn,
		limit, offset,
	)

	count, err = s.countProducts(searchQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check count: %w", err)
	}

	if count < 1 {
		return nil, 0, sql.ErrNoRows
	}

	IDsRows, err := tx.Query(querySearchIDs, searchQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", ErrExecStmt, err)
	}

	defer IDsRows.Close()

	if err := IDsRows.Err(); err != nil {
		log.Printf("ROWS ITERATION ERROR: %v", err)
	}

	var list []*models.Product

	var id int

	// TODO  ВОТ ТУТ ПОЧЕМУ ТО НЕ СКАНИУРЕТСЯ НИХУЯ БЛЯТЬ И НЕ ПЕРЕХОДИТ, ХОТЯ Я СДЕЛАЛ ТАКОЙ ЖЕ ЗАПРОС В КВЕРИ ТУЛЕ И ТАМ ВСЕ НОРМ почини блтяь
	for IDsRows.Next() {

		err = IDsRows.Scan(&id)
		if err != nil {
			s.log.Error("failed to get product by id", "error", err)
			continue
		}

		query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s 
		FROM %s
		WHERE %s = ?`,
			productsTitleColumm, productsDescripColumn, productsVKLoadedColumn, productsAvitoLoadedColumn, productsUcozLoadedColumn,
			productsTable,
			productsIdColumn)

		var p models.Product

		p.Id = int64(id)

		err = tx.QueryRow(query, id).Scan(
			&p.Title,
			&p.Description,

			&p.VK.ToLoad,
			&p.Avito.ToLoad,
			&p.Ucoz.ToLoad,
		)

		query = fmt.Sprintf(`
		SELECT %s, %s
		FROM %s
		WHERE product_id = ? `,
			productImagesTelegramFileID,
			productImagesTelegramUrl,
			productImagesTable,
		)

		err = tx.QueryRow(query, id).Scan(
			&p.TelegramFileID,
			&p.MainPictureURL,
		)

		if err != nil {
			s.log.Error("can't scan row", "err", err.Error())
			return nil, 0, fmt.Errorf("can't scan row: %w", err)
		}

		list = append(list, &p)

	}

	if IDsRows.Err() != nil {
		return nil, 0, err
	}

	tx.Commit()

	return list, count, nil

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

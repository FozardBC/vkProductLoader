package storage

import "errors"

type Storage interface {
	SaveID(ProductID int, CategoryID int) error
	Close() error
	Ping() error
}

var (
	ErrProductIDExists = errors.New("product ID already exists in storage")
)

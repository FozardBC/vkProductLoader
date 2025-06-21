package storage

import (
	"errors"
	"prodLoaderREST/internal/domain/filters"
)

type Storage interface {
	SaveID(ProductID int, CategoryID int) error
	GetProdIDs(options *filters.Options) ([]int, error)
	Delete(productID int) error
	Close() error
	Ping() error
}

var (
	ErrProductIDExists   = errors.New("product ID already exists in storage")
	ErrProductIDnotFound = errors.New("product ID not found in storage")
)

package storage

import (
	"context"
	"errors"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/domain/models"
)

type Storage interface {
	Save(ctx context.Context, product *models.Product) (int64, error)
	GetProdIDs(options *filters.Options) ([]int, error)
	Delete(ctx context.Context, productID int) error
	Close() error
	Ping() error
}

var (
	ErrProductIDExists   = errors.New("product ID already exists in storage")
	ErrProductIDnotFound = errors.New("product ID not found in storage")
	ErrReturnId          = errors.New("failed to return id of product ")
	ErrBeginTx           = errors.New("failed to begin transaction")
	ErrCommitTx          = errors.New("failed to commit transaction")
)

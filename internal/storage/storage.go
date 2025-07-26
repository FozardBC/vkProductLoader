package storage

import (
	"context"
	"errors"
	"prodLoaderREST/internal/domain/models"
)

type Storage interface {
	Save(ctx context.Context, product *models.Product) (int64, error)
	VkProductID(productID int64) (int, error)
	Search(ctx context.Context, searchQuery string, offset int, limit int) (products []*models.Product, count int, err error)
	UcozLoaded(productID int64, ucozProductID int) error
	VkLoaded(productID int64, vkProductID int) error
	VkDeleted(productID int64) error
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

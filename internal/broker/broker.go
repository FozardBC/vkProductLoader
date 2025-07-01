package broker

import (
	"context"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/storage"
)

var VKaddProductChannel = make(chan *models.Product, 100)
var VKdeleteProductChannel = make(chan int, 100)

var UcozAddProductChannel = make(chan *models.Product, 100)
var UcozDeleteProductChannel = make(chan *models.Product, 100)

type Exchanger struct {
	log     *slog.Logger
	storage storage.Storage
}

func New(log *slog.Logger, storage storage.Storage) *Exchanger {
	return &Exchanger{
		log:     log,
		storage: storage,
	}
}

func (e *Exchanger) WriteAdd(ctx context.Context, product *models.Product) error {

	if product == nil {
		return fmt.Errorf("nil product")
	}

	id, err := e.storage.Save(ctx, product)
	if err != nil {

		return err
	}

	product.Id = id

	go func() {
		if product.VK.ToLoad {
			VKaddProductChannel <- product
		}
	}()

	go func() {
		if product.Ucoz.ToLoad {
			UcozAddProductChannel <- product
		}
	}()

	go func() {
		if product.Ucoz.ToLoad {

		}
	}()

	return nil
}

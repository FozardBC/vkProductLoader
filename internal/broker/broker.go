package broker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/storage"
	"prodLoaderREST/internal/storage/pictureManager"
)

var VKaddProductChannel = make(chan *models.Product, 100)
var VKdeleteProductChannel = make(chan *VkToDelete, 100)

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

	err = pictureManager.SavePicture(int(id), product.MainPictureURL)
	if err != nil {
		e.log.Warn("failed to save picture", "err", err.Error())
	}

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

	// go func() {
	// 	if product.Avito.ToLoad {

	// 	}
	// }()

	return nil
}

func (e *Exchanger) WriteDelete(productID int) error {

	DeleteID := VkToDelete{
		ProductID: productID,
	}

	if DeleteID.ProductID < 1 {
		return fmt.Errorf("ProductID can't be less 1")
	}

	var err error

	DeleteID.VkProductID, err = e.storage.VkProductID(int64(DeleteID.ProductID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("product is not exists")
		}
		return fmt.Errorf("failed to get vk prod id:%w", err)

	}

	go func() {
		VKdeleteProductChannel <- &DeleteID
	}()
	e.log.Debug("productID written to delete VK", "VKproductID", DeleteID.VkProductID)

	// тут также потом будет писать в юкоз

	return nil

}

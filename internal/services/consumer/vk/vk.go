package vk

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/domain/models"

	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/v3/api"
)

var (
	ErrNotAllProductsDeleted = errors.New("not all products were deleted")
)

type StatusChanger interface {
	VkLoaded(productID int64, vkProductID int) error
}

type VkConsumer struct {
	log           *slog.Logger
	VK            *api.VK
	StatusChanger StatusChanger
	groupID       int
}

func New(log *slog.Logger, vk *api.VK, statusChanger StatusChanger, groupID int) *VkConsumer {
	return &VkConsumer{
		log:           log,
		VK:            vk,
		StatusChanger: statusChanger,
		groupID:       groupID,
	}
}

func (v *VkConsumer) GetClientName() string {

	p := params.NewAccountGetInfoBuilder()

	info, err := v.VK.AccountGetProfileInfo(api.Params(p.Params))
	if err != nil {
		log.Fatal(err)
	}

	return info.FirstName + " " + info.LastName
}

func (v *VkConsumer) Load(products chan *models.Product) {
	for p := range products {

		go func() {

			v.log.Debug("Received product", "product", p)

			if p == nil {
				v.log.Error("Received nil product")
				return
			}

			log := v.log.With("Title", p.Title)

			MainPicResponse, err := v.loadMainPicture(p.MainPictureURL)
			if err != nil {
				log.Error("Failed to load main picture", "err", err.Error())
				return
			}

			log.Debug("Main picture loaded")

			PicturesIDs, err := v.loadPictures(log, p.PicturesURL)
			if err != nil {
				log.Error("Failed to load pictures", "err", err.Error())
				return
			}

			pars := params.NewMarketAddBuilder()

			pars.OwnerID(-v.groupID)
			pars.Name(p.Title)
			pars.MainPhotoID(MainPicResponse[0].ID)
			pars.PhotoIDs(PicturesIDs)
			pars.Description(p.Description)
			pars.Price(float64(p.Price))
			pars.CategoryID(p.VK.CategoryID)

			response, err := v.VK.MarketAdd(api.Params(pars.Params))
			if err != nil {
				log.Error("Failed to add product to market", "err", err.Error(), "respone", response)
				return
			}

			err = v.StatusChanger.VkLoaded(p.Id, response.MarketItemID)
			if err != nil {
				log.Error("failed to change status", "error", err)
			}

			log.Debug("Product added to market. ID Saved in storage")
		}()
	}
}

func (v *VkConsumer) Delete(options *filters.Options) (int, error) {

	//	pars := params.NewMarketDeleteBuilder()
	//
	// pars.OwnerID(-v.groupID)
	//
	// ProductIDs, err := v.storage.GetProdIDs(options)
	// if err != nil {
	// if errors.Is(err, storage.ErrProductIDnotFound) {
	// v.log.Warn("no product founds for given filters", "options", options)
	//
	// return 0, err
	// }
	// v.log.Error("Failed to get product IDs from storage", "err", err.Error())
	// return 0, fmt.Errorf("failed to get product IDs from storage: %w", err)
	// }
	//
	// wg := sync.WaitGroup{}
	//
	// for i, productID := range ProductIDs {
	// wg.Add(1)
	// go func(wg *sync.WaitGroup) {
	//
	// if productID == 0 {
	// v.log.Debug("No products found for deletion", "count", len(ProductIDs))
	// wg.Done()
	// return
	// }
	//
	// v.log.Debug("Deleting product", "productID", productID)
	//
	// pars.ItemID(productID)
	//
	// _, err := v.VK.MarketDelete(api.Params(pars.Params))
	// if err != nil {
	// v.log.Error("Failed to delete product from market", "productID", productID, "err", err.Error())
	//
	// return
	// }
	//
	// ProductIDs[i] = 0 // Удаляем ID из слайса, чтобы не удалять его повторно
	//
	// wg.Done()
	//
	// v.log.Debug("Product deleted successfully", "productID", productID)
	// }(&wg)
	// }

	// wg.Wait()

	// var NotDeletedCount int

	// errNotDeleted := "Not deleted products: "

	// for _, productID := range ProductIDs {
	// 	if productID != 0 {
	// 		NotDeletedCount++

	// 		errNotDeleted += fmt.Sprintf("%d; ", productID)
	// 	}
	// }

	// if NotDeletedCount > 0 {
	// 	v.log.Error("Some products were not deleted", "count", NotDeletedCount, "details", errNotDeleted)
	// 	return len(ProductIDs) - NotDeletedCount, fmt.Errorf("%w: %s", ErrNotAllProductsDeleted, errNotDeleted)
	// }

	// v.log.Info("All products deleted successfully", "count", len(ProductIDs))

	return 0, nil

}

func (v *VkConsumer) loadPictures(log *slog.Logger, picURLs []string) ([]int, error) {

	PicturesIDs := make([]int, 0)

	for _, pic := range picURLs {
		if len(PicturesIDs) == 4 {
			break
		}

		if pic == "" {
			log.Warn("Picture URL is empty, skipping")
			continue
		}

		resp, err := http.Get(pic)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexepeted status: %s", resp.Status)
		}

		respPhoto, err := v.VK.UploadMarketPhoto(v.groupID, false, resp.Body)
		if err != nil {
			v.log.Warn("failed to upload picture", "err", err.Error())
		}

		if len(respPhoto) >= 1 {
			PicturesIDs = append(PicturesIDs, respPhoto[0].ID)
		}
	}

	return PicturesIDs, nil
}

func (v *VkConsumer) loadMainPicture(picURL string) (api.PhotosSaveMarketPhotoResponse, error) {

	if picURL == "" {
		return nil, fmt.Errorf("main picture URL is empty")
	}

	resp, err := http.Get(picURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	vkMainPicResp, err := v.VK.UploadMarketPhoto(v.groupID, true, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't load MainPhoto to VK: %s", err.Error())
	}

	v.log.Debug("Main picture loaded", "url", picURL, "content-length", resp.ContentLength)

	if len(vkMainPicResp) == 0 {
		return nil, fmt.Errorf("no response from VK when uploading main picture")
	}

	return vkMainPicResp, nil
}

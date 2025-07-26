package vk

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/broker"
	"prodLoaderREST/internal/domain/models"
	"strings"

	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/v3/api"
)

var (
	ErrNotAllProductsDeleted = errors.New("not all products were deleted")
)

type StatusChanger interface {
	VkLoaded(productID int64, vkProductID int) error
	VkDeleted(productID int64) error
}

type Consumer struct {
	log           *slog.Logger
	vk            *api.VK
	statusChanger StatusChanger
	groupID       int
}

func New(log *slog.Logger, vk *api.VK, groupID int, StatusChanger StatusChanger) *Consumer {
	return &Consumer{
		log:           log,
		vk:            vk,
		statusChanger: StatusChanger,
		groupID:       groupID,
	}
}

func (v *Consumer) GetClientName() string {

	p := params.NewAccountGetInfoBuilder()

	info, err := v.vk.AccountGetProfileInfo(api.Params(p.Params))
	if err != nil {
		log.Fatal(err)
	}

	return info.FirstName + " " + info.LastName
}

func (v *Consumer) ListenLoad(products chan *models.Product) {
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

			parts := strings.Split(p.Description, "\n") //надо для корректного отображения названия товарв

			pars := params.NewMarketAddBuilder()

			pars.OwnerID(-v.groupID)
			pars.Name(p.Title)
			pars.MainPhotoID(MainPicResponse[0].ID)
			pars.PhotoIDs(PicturesIDs)
			pars.Description(p.Description)
			pars.Price(float64(p.Price))
			pars.CategoryID(p.VK.CategoryID)

			if len(parts) > 1 {
				pars.Name(parts[0])
			}

			response, err := v.vk.MarketAdd(api.Params(pars.Params))
			if err != nil {
				log.Error("Failed to add product to market", "err", err.Error(), "respone", response)
				return
			}

			err = v.statusChanger.VkLoaded(p.Id, response.MarketItemID)
			if err != nil {
				log.Error("failed to change status", "error", err)
			}

			log.Debug("Product added to market. ID Saved in storage")
		}()
	}
}

func (v *Consumer) ListenDelete(products chan *broker.VkToDelete) {

	for id := range products {
		v.log.Debug("Recived product to delete", "productID", id.ProductID)

		pars := params.NewMarketDeleteBuilder()

		pars.OwnerID(-v.groupID)

		pars.ItemID(id.VkProductID)

		_, err := v.vk.MarketDelete(api.Params(pars.Params))
		if err != nil {
			v.log.Error("Failed to delete product from market", "productID", id, "err", err.Error())

			return
		}

		v.log.Debug("product deleted from VK", "productID", id.ProductID)

		err = v.statusChanger.VkDeleted(int64(id.ProductID))
		if err != nil {
			v.log.Error("Failed to delete product from storage", "productID", id.ProductID, "err", err)

		}

	}

}

func (v *Consumer) loadPictures(log *slog.Logger, picURLs []string) ([]int, error) {

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

		respPhoto, err := v.vk.UploadMarketPhoto(v.groupID, false, resp.Body)
		if err != nil {
			v.log.Warn("failed to upload picture", "err", err.Error())
		}

		if len(respPhoto) >= 1 {
			PicturesIDs = append(PicturesIDs, respPhoto[0].ID)
		}
	}

	return PicturesIDs, nil
}

func (v *Consumer) loadMainPicture(picURL string) (api.PhotosSaveMarketPhotoResponse, error) {

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

	vkMainPicResp, err := v.vk.UploadMarketPhoto(v.groupID, true, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't load MainPhoto to VK: %s", err.Error())
	}

	v.log.Debug("Main picture loaded", "url", picURL, "content-length", resp.ContentLength)

	if len(vkMainPicResp) == 0 {
		return nil, fmt.Errorf("no response from VK when uploading main picture")
	}

	return vkMainPicResp, nil
}

package vk

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/domain/models"

	"prodLoaderREST/internal/storage"

	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/v3/api"
)

type VkConsumer struct {
	log     *slog.Logger
	VK      *api.VK
	Storage storage.Storage
	groupID int
}

func New(log *slog.Logger, vk *api.VK, storage storage.Storage, groupID int) *VkConsumer {
	return &VkConsumer{
		log:     log,
		VK:      vk,
		Storage: storage,
		groupID: groupID,
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

			MainPicResponse, err := v.loadMainPicture(p.MainPicture)
			if err != nil {
				log.Error("Failed to load main picture", "err", err.Error())
				return
			}

			log.Debug("Main picture loaded")

			PicturesIDs, err := v.loadPictures(log, p.Pictures)
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
			pars.CategoryID(p.VKCategoryID)

			response, err := v.VK.MarketAdd(api.Params(pars.Params))
			if err != nil {
				log.Error("Failed to add product to market", "err", err.Error())
				return
			}

			err = v.Storage.SaveID(response.MarketItemID, p.VKCategoryID)
			if err != nil {
				if errors.Is(err, storage.ErrProductIDExists) {
					log.Warn("Product ID already exists in storage, skipping", "err", err.Error())
					return
				}
				log.Warn("Failed to save product ID in storage", "err", err.Error())
			}

			log.Debug("Product added to market. ID Saved in storage")
		}()
	}
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

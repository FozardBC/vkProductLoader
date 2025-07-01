package add

import (
	"context"
	"log/slog"
	"net/http"

	"prodLoaderREST/internal/api/middlewares/requestid"
	"prodLoaderREST/internal/api/types"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/lib/api/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Avito struct {
	ToLoad bool `json:"toLoad"`
}

type DanisaBot struct {
	ToLoad bool `json:"toLoad"`
}
type VK struct {
	ToLoad     bool `json:"toLoad"`
	CategoryID int  `json:"categoryID"`
}

type Request struct {
	Title          string    `json:"title" validate:"required"`
	Description    string    `json:"description" validate:"required"`
	Size           string    `json:"size" validate:"required"`
	Status         string    `json:"status" validate:"required"`
	Price          int       `json:"price" validate:"required"`
	TelegramFileID string    `json:"tg_fileID" validate:"required"`
	MainPictureURL string    `json:"mainPictureURL" validate:"required"`
	PicturesURL    []string  `json:"picturesURL" validate:"required"`
	VK             VK        `json:"vk"`
	Avito          Avito     `json:"avito"`
	DanisaBot      DanisaBot `json:"danisa_bot"`
}

type Exchanger interface {
	WriteAdd(ctx context.Context, product *models.Product) error
}

func New(log *slog.Logger, exchanger Exchanger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logHandler := log.With("requestID", requestid.Get(c))

		ctx := c.Request.Context()

		var Product *models.Product

		if err := c.BindJSON(&Product); err != nil {
			logHandler.Error(types.ErrDecodeReqBody.Error(), "err", err.Error())

			c.JSON(http.StatusBadRequest, response.Error(types.ErrDecodeReqBody.Error()))
			return
		}

		if err := validator.New().Struct(Product); err != nil {
			validatorErr := err.(validator.ValidationErrors)

			logHandler.Error("invalid request", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.ValidationError(validatorErr))

			return
		}

		logHandler.Debug("received product", "product", Product)

		if err := exchanger.WriteAdd(ctx, Product); err != nil {
			logHandler.Error("failed to write to broker", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal Error"))

			return
		}

		logHandler.Info("product added to queue", "product", Product.Title)

		c.JSON(http.StatusOK, response.OK)

	}
}

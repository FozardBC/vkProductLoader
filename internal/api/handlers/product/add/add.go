package add

import (
	"log/slog"
	"net/http"

	"prodLoaderREST/internal/api/middlewares/requestid"
	"prodLoaderREST/internal/api/types"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/lib/api/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Title          string   `json:"title" validate:"required"`
	Description    string   `json:"description" validate:"required"`
	Size           string   `json:"size" validate:"required"`
	Status         string   `json:"status" validate:"required"`
	Price          int      `json:"price" validate:"required"`
	MainPictureURL string   `json:"mainPictureURL" validate:"required"`
	PicturesURL    []string `json:"picturesURL" validate:"required"`
	CategoryID     int      `json:"categoryID" validate:"required"`
}

func New(log *slog.Logger, queue chan *models.Product) gin.HandlerFunc {
	return func(c *gin.Context) {
		logHandler := log.With("requestID", requestid.Get(c))

		var Request Request

		if err := c.BindJSON(&Request); err != nil {
			logHandler.Error(types.ErrDecodeReqBody.Error(), "err", err.Error())

			c.JSON(http.StatusBadRequest, response.Error(types.ErrDecodeReqBody.Error()))
			return
		}

		if err := validator.New().Struct(Request); err != nil {
			validatorErr := err.(validator.ValidationErrors)

			logHandler.Error("invalid request", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.ValidationError(validatorErr))

			return
		}

		logHandler.Debug("received product", "product", Request)

		product := &models.Product{
			Title:        Request.Title,
			Description:  Request.Description,
			Size:         Request.Size,
			Status:       Request.Status,
			Price:        Request.Price,
			MainPicture:  Request.MainPictureURL,
			Pictures:     Request.PicturesURL,
			VKCategoryID: Request.CategoryID,
		}

		queue <- product

		logHandler.Info("product added to queue", "product", Request)

		c.JSON(http.StatusOK, response.OK)

	}
}

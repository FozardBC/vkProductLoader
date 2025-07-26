package delete

import (
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/api/types"
	"prodLoaderREST/internal/lib/api/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductDeleteWriter interface {
	WriteDelete(productID int) error
}

func New(log *slog.Logger, Deleter ProductDeleteWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		logHandler := log.With("requestID", c.GetString("requestID"))

		var prodIDint int

		var err error

		productID := c.Query("product_id")
		if productID != "" {
			prodIDint, err = strconv.Atoi(productID)
			if err != nil {
				log.Error(types.ErrConvertParam.Error(), "param", "product_ID", "query", productID)
				c.JSON(http.StatusBadRequest, response.Error("productID is not integer"))

				return
			}

		}

		err = Deleter.WriteDelete(prodIDint)
		if err != nil {
			log.Error("failed to write to delete", "err", err.Error())

			c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		}

		logHandler.Info("Product deleted successfully")
		c.JSON(200, response.OK())
	}
}

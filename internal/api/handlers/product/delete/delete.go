package delete

import (
	"errors"
	"fmt"
	"log/slog"
	"prodLoaderREST/internal/api/types"
	"prodLoaderREST/internal/domain/filters"
	"prodLoaderREST/internal/services/consumer/vk"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductDeleter interface {
	Delete(options *filters.Options) (int, error)
}

func New(log *slog.Logger, deleter ProductDeleter) gin.HandlerFunc {
	return func(c *gin.Context) {
		logHandler := log.With("requestID", c.GetString("requestID"))

		options, err := setFilterQueries(logHandler, c)
		if err != nil {
			logHandler.Error("Failed to set filter queries", "error", err.Error())
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to set filter queries: %s", err.Error())})
			return
		}

		count, err := deleter.Delete(options)
		if err != nil {
			if errors.Is(err, vk.ErrNotAllProductsDeleted) {
				logHandler.Error("Not all products were deleted", "error", err.Error(), "count", count)
				c.JSON(500, gin.H{"error": fmt.Sprintf("Not all products were deleted: %s", err.Error())})
				return
			}
			logHandler.Error("Failed to delete products", "error", err.Error())
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete products: %s", err.Error())})
			return
		}

		logHandler.Info("Product deleted successfully", "count", count)
		c.JSON(200, gin.H{"message": "Product deleted successfully"})
	}
}

func setFilterQueries(log *slog.Logger, c *gin.Context) (*filters.Options, error) {

	var op filters.Options

	productID := c.Query("product_id")
	if productID != "" {
		prodIDint, err := strconv.Atoi(productID)
		if err != nil {
			log.Error(types.ErrConvertParam.Error(), "param", "product_ID", "query", productID)
			return nil, fmt.Errorf("%w:%s", types.ErrConvertParam, productID)
		}

		op.ProductID = &prodIDint

	}

	categoryID := c.Query("category_id")
	if categoryID != "" {
		categoryIDint, err := strconv.Atoi(categoryID)
		if err != nil {
			log.Error(types.ErrConvertParam.Error(), "param", "product_ID", "query", categoryID)
			return nil, fmt.Errorf("%w:%s", types.ErrConvertParam, categoryID)
		}

		op.CategoryID = &categoryIDint

	}

	return &op, nil
}

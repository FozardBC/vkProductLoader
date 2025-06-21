package delete

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func New(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logHandler := log.With("requestID", c.GetString("requestID"))

		productID := c.Param("productID")

		logHandler.Info("Product deleted successfully", "productID", productID)
		c.JSON(200, gin.H{"message": "Product deleted successfully"})
	}
}

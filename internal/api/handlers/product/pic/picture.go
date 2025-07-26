package pic

import (
	"errors"
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/api/middlewares/requestid"
	"prodLoaderREST/internal/lib/api/response"
	"prodLoaderREST/internal/storage/pictureManager"

	"github.com/gin-gonic/gin"
)

func New(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		id := c.Param("id")
		if id == "" {
			logHandler.Error("empty product ID")
			c.JSON(http.StatusBadRequest, response.Error("empty product ID"))
			return
		}

		file, err := pictureManager.Picture(id)
		if err != nil {
			if errors.Is(err, pictureManager.ErrNotFound) {
				logHandler.Error("picture not found", "id", id)
				c.JSON(http.StatusNotFound, response.Error("Picture not found"))
				return
			}
			logHandler.Error("failed to get picture", "err", err.Error())
			c.JSON(http.StatusInternalServerError, response.Error("Internal Error"))
			return
		}

		if file == nil {
			logHandler.Error("picture not found", "id", id)
			c.JSON(http.StatusNotFound, response.Error("Picture not found"))
			return
		}

		c.Header("Content-Type", "image/jpeg")
		c.Status(http.StatusOK)

		c.Writer.Write(file)

	}
}

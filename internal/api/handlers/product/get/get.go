package get

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prodLoaderREST/internal/api/middlewares/requestid"
	"prodLoaderREST/internal/api/types"
	"prodLoaderREST/internal/domain/models"
	"prodLoaderREST/internal/lib/api/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductSearcher interface {
	Search(ctx context.Context, query string, offset int, limit int) (products []*models.Product, count int, err error)
}

var ErrConvertParam = errors.New("can't convernt int query parameter")

const defaultPage = "1"
const defaultLimit = "10"

func New(log *slog.Logger, searcher ProductSearcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		var pag types.Pagination

		var err error

		searchQuery := c.Query("search")
		if searchQuery == "" {
			logHandler.Error("empty search query")

			c.JSON(http.StatusBadRequest, response.Error("empty search query"))

			return
		}

		pageQuery := c.DefaultQuery("page", defaultPage)
		if pageQuery == "0" {
			logHandler.Error("page can't be 0")

			c.JSON(http.StatusBadRequest, "page parameter can't be 0")
			return
		}

		pag.Page, err = strconv.Atoi(pageQuery)
		if err != nil {
			logHandler.Error(ErrConvertParam.Error(), "param", "page", "query", pageQuery)

			c.JSON(http.StatusBadRequest, response.Error(fmt.Sprintf("Invalid parameter:%s", pageQuery)))

			return
		}

		limitQurey := c.DefaultQuery("limit", defaultLimit)

		pag.Limit, err = strconv.Atoi(limitQurey)
		if err != nil {
			logHandler.Error(ErrConvertParam.Error(), "param", "limit", "query", limitQurey)

			c.JSON(http.StatusBadRequest, response.Error(fmt.Sprintf("Invalid parameter:%s", limitQurey)))

			return
		}

		logHandler.Debug("Pagination query", "page", pag.Page, "limit", pag.Limit)

		products, count, err := searcher.Search(ctx, searchQuery, pag.Offset(), pag.Limit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logHandler.Error("no data found", "query", searchQuery)

				c.JSON(http.StatusNoContent, "")
				return
			}
			logHandler.Error("can't get list of persons", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal Server Error"))
			return
		}

		meta := &types.Meta{
			Total:  count,
			Limit:  pag.Limit,
			Offset: pag.Offset(),
			Next:   (pag.Offset() + pag.Limit) < count,
		}

		c.JSON(http.StatusOK, response.OKWithPayload(map[string]interface{}{"data": products, "meta": meta}))

	}
}

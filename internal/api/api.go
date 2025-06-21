package api

import (
	"log/slog"

	"prodLoaderREST/internal/broker"

	"prodLoaderREST/internal/api/handlers/product/add"
	"prodLoaderREST/internal/api/middlewares/requestid"
	"prodLoaderREST/internal/api/types"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	Router *gin.Engine
	Log    *slog.Logger
}

func New(log *slog.Logger, servs ...types.Consumer) *API {
	return &API{
		Router: gin.New(),
		Log:    log,
	}
}

func (api *API) Setup() {
	v1 := api.Router.Group("api/v1/")

	v1.Use(requestid.RequestIdMidlleware())
	//v1.Use(gin.LoggerWithFormatter(log.Logging))

	v1.POST("/products", add.New(api.Log, broker.VKProductChannel))

	v1.GET("/swagger/*any", gin.WrapH(httpSwagger.Handler()))

}

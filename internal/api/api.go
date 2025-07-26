package api

import (
	"log/slog"

	"prodLoaderREST/internal/broker"
	"prodLoaderREST/internal/lib/api/log"
	"prodLoaderREST/internal/services/productManager"
	"prodLoaderREST/internal/storage"

	"prodLoaderREST/internal/api/handlers/product/add"
	deleteHandler "prodLoaderREST/internal/api/handlers/product/delete"
	"prodLoaderREST/internal/api/handlers/product/get"
	"prodLoaderREST/internal/api/handlers/product/pic"
	"prodLoaderREST/internal/api/middlewares/requestid"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	Router         *gin.Engine
	Log            *slog.Logger
	productManager *productManager.Manager
	Exchanger      *broker.Exchanger
	Storage        storage.Storage
}

func New(log *slog.Logger, productManager *productManager.Manager, Exchanger *broker.Exchanger, storage storage.Storage) *API {
	return &API{
		Router:         gin.New(),
		Log:            log,
		productManager: productManager,
		Exchanger:      Exchanger,
		Storage:        storage,
	}
}

func (api *API) Setup() {
	v1 := api.Router.Group("api/v1/")

	v1.Use(requestid.RequestIdMidlleware())
	v1.Use(gin.LoggerWithFormatter(log.Logging))

	v1.POST("/products", add.New(api.Log, api.Exchanger))
	v1.DELETE("/products/", deleteHandler.New(api.Log, api.Exchanger))
	v1.GET("/products/", get.New(api.Log, api.Storage))
	v1.GET("/products/picture/:id", pic.New(api.Log))

	v1.GET("/swagger/*any", gin.WrapH(httpSwagger.Handler()))

}

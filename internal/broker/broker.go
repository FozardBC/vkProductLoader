package broker

import "prodLoaderREST/internal/domain/models"

var VKaddProductChannel = make(chan *models.Product, 100)
var VKdeleteProductChannel = make(chan int, 100)

var ShopAddProductChannel = make(chan *models.Product, 100)
var ShopDeleteProductChannel = make(chan *models.Product, 100)

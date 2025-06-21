package broker

import "prodLoaderREST/internal/domain/models"

var VKProductChannel = make(chan *models.Product, 100)

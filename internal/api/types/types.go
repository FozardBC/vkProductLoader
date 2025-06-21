package types

import (
	"errors"
	"prodLoaderREST/internal/domain/models"
)

var (
	ErrDecodeReqBody = errors.New("failed to decode request body")
)

type Consumer interface {
	Load(chan *models.Product)
}

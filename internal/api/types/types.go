package types

import (
	"errors"
	"prodLoaderREST/internal/domain/models"
)

var (
	ErrDecodeReqBody = errors.New("failed to decode request body")
	ErrConvertParam  = errors.New("can't convert int query parameter")
)

type Consumer interface {
	Load(chan *models.Product)
}

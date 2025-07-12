package models

type Product struct {
	Id             int64
	Title          string   `json:"title" validate:"required"`
	Description    string   `json:"description" validate:"required"`
	Size           string   `json:"size" validate:"required"`
	Status         string   `json:"status" validate:"required"`
	Price          int      `json:"price" validate:"required"`
	TelegramFileID string   `json:"telegram_fileID" validate:"required"`
	MainPictureURL string   `json:"mainPictureURL" validate:"required"`
	PicturesURL    []string `json:"picturesURL" validate:"required"`
	VK             VK       `json:"vk"`
	Avito          Avito    `json:"avito"`
	Ucoz           Ucoz     `json:"ucoz"`
}

type Avito struct {
	ToLoad bool `json:"toLoad"`
}

type Ucoz struct {
	ToLoad bool `json:"toLoad"`
}

type VK struct {
	ToLoad     bool `json:"toLoad"`
	CategoryID int  `json:"categoryID" validate:"required"`
}

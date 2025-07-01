package models

type Product struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Size        string    `json:"size" validate:"required"`
	Status      string    `json:"status" validate:"required"`
	Price       int       `json:"price" validate:"required"`
	MainPicture string    `json:"mainPictureURL" validate:"required"`
	Pictures    []string  `json:"picturesURL" validate:"required"`
	VK          VK        `json:"vk"`
	Avito       Avito     `json:"avito"`
	DanisaBot   DanisaBot `json:"danisa_bot"`
}

type Avito struct {
	ToLoad bool `json:"toLoad"`
}

type DanisaBot struct {
	ToLoad bool `json:"toLoad"`
}

type VK struct {
	ToLoad     bool `json:"toLoad"`
	CategoryID int  `json:"categoryID" validate:"required"`
}

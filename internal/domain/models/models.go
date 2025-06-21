package models

type Product struct {
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	Size         string   `json:"size" validate:"required"`
	Status       string   `json:"status" validate:"required"`
	Price        int      `json:"price" validate:"required"`
	MainPicture  string   `json:"mainPictureURL" validate:"required"`
	Pictures     []string `json:"picturesURL" validate:"required"`
	VKCategoryID int      `json:"VKcategoryID" validate:"required"`
}

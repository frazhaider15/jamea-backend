package dto

type AddDefaultActivityRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

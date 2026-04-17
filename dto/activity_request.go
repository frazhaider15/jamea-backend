package dto

import "github.com/jamea/models"

type ActivityRequest struct {
	Module     models.Module `json:"module" binding:"required"`
	Month      int           `json:"month" binding:"required"`
	Year       int           `json:"year" binding:"required"`
	Activities []string      `json:"activities" binding:"required"`
}

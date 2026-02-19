package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jamea/db"
	"github.com/jamea/dto"
	"github.com/jamea/models"
)

func Login(request dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	if err := db.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// In a real app, compare hashed password. Here plain text as per persistent storage requirement implies simplicity for now,
	// but using DB query.
	if user.Password != request.Password {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate a simple token
	token := uuid.New().String()

	return &dto.LoginResponse{
		Token:  token,
		Module: user.Module,
	}, nil
}

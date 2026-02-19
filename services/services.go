package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jamea/dto"
	"github.com/jamea/store"
)

func Login(request dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := store.FindUserByEmail(request.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

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

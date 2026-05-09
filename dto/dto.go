package dto

import "github.com/jamea/models"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string        `json:"token"`
	Module models.Module `json:"module"`
}

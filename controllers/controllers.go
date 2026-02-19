package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamea/dto"
	"github.com/jamea/services"
)

func Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}
	if err := request.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}

	response, err := services.Login(request)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, NewStandardResponse(false, 401, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, http.StatusOK, "Login successful", response))
}

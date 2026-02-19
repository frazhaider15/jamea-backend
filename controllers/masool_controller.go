package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamea/services"
)

func UploadMasool(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "file is required", nil))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, "failed to open file", nil))
		return
	}
	defer file.Close()

	err = services.UploadMasool(file)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool data uploaded successfully", nil))
}

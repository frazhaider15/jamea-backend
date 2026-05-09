package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

// GetMasools handles GET /masool?module=MODULE
func GetMasools(ctx *gin.Context) {
	moduleStr := ctx.Query("module")
	if moduleStr == "" {
		ctx.JSON(400, NewStandardResponse(false, 400, "module is required", nil))
		return
	}

	module, err := models.NewModuleString(moduleStr)
	if err != nil {
		ctx.JSON(400, NewStandardResponse(false, 400, "invalid module", nil))
		return
	}

	masools, err := services.GetMasoolsByModule(module)
	if err != nil {
		ctx.JSON(500, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	ctx.JSON(200, NewStandardResponse(true, 200, "Masools fetched successfully", masools))
}

func UploadMasool(ctx *gin.Context) {
	// Extract module from query parameter or context (assuming authenticated user's module)
	// For now, let's assume it's passed as a query param module for simplicity in testing,
	// OR better: get it from the authenticated user.
	// Since I haven't implemented middleware to set user context yet, I'll accept it as a query param.
	// But requirements say \"masools should be mapped against a module\".
	// The login returns a module. The frontend should send this module back or use the token.
	// I'll look for module in query string for now to unblock.

	moduleStr := ctx.Query("module")
	if moduleStr == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "module is required", nil))
		return
	}

	module, err := models.NewModuleString(moduleStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid module", nil))
		return
	}

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

	data, err := services.UploadMasool(file, module)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool data uploaded successfully", data))
}

func DeleteMasool(ctx *gin.Context) {
	moduleStr := ctx.Query("module")
	if moduleStr == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "module is required", nil))
		return
	}

	module, err := models.NewModuleString(moduleStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid module", nil))
		return
	}

	err = services.DeleteMasoolsByModule(module)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool data deleted successfully", nil))
}

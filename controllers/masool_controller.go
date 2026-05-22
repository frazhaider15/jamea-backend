package controllers

import (
	"fmt"
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

	services.LogUserActivity(module, "upload_masool", fmt.Sprintf("Uploaded %d Masool record(s)", len(data)))
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

	services.LogUserActivity(module, "delete_masool", "Deleted all Masool data")
	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool data deleted successfully", nil))
}

// CreateMasool handles POST /masool
func CreateMasool(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid request body", nil))
		return
	}

	name, _ := body["name"].(string)
	moduleStr, _ := body["module"].(string)

	if name == "" || moduleStr == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "name and module are required", nil))
		return
	}

	module, err := models.NewModuleString(moduleStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid module", nil))
		return
	}

	masool := models.Masool{
		Name:   name,
		Module: module,
		Data:   make([]models.MasoolData, 0),
	}

	// Capture all other fields as dynamic data
	for k, v := range body {
		if k == "name" || k == "module" {
			continue
		}
		masool.Data = append(masool.Data, models.MasoolData{
			Key: k,
			Val: v,
		})
	}

	if err := services.CreateMasool(&masool); err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	services.LogUserActivity(module, "create_masool", "Created Masool: "+name)
	ctx.JSON(http.StatusCreated, NewStandardResponse(true, 201, "Masool created successfully", masool))
}

// GetMasool handles GET /masool/:id
func GetMasool(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid ID", nil))
		return
	}

	masool, err := services.GetMasoolByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, NewStandardResponse(false, 404, "Masool not found", nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool fetched successfully", masool))
}

// UpdateMasool handles PUT /masool/:id
func UpdateMasool(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid ID", nil))
		return
	}

	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid request body", nil))
		return
	}

	var data []models.MasoolData
	for k, v := range body {
		data = append(data, models.MasoolData{
			Key: k,
			Val: v,
		})
	}

	masool, err := services.UpdateMasoolByID(id, data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	services.LogUserActivity(masool.Module, "update_masool", "Updated Masool: "+masool.Name)
	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool updated successfully", masool))
}

// DeleteMasoolByID handles DELETE /masool/:id
func DeleteMasoolByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid ID", nil))
		return
	}

	masool, err := services.GetMasoolByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, NewStandardResponse(false, 404, "Masool not found", nil))
		return
	}

	if err := services.DeleteMasoolByID(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	services.LogUserActivity(masool.Module, "delete_masool", "Deleted Masool: "+masool.Name)
	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool deleted successfully", nil))
}

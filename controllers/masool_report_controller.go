package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

func UploadMasoolReport(ctx *gin.Context) {
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

	data, err := services.UploadMasoolReport(file, module)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}

	services.LogUserActivity(module, "upload_masool_report", fmt.Sprintf("Uploaded %d Masool report record(s)", len(data)))
	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool report data uploaded successfully", data))
}

func GetMasoolReport(ctx *gin.Context) {
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

	month := ctx.Query("month")
	data, err := services.GetMasoolReport(module, month)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool report fetched successfully", data))
}

func DeleteMasoolReport(ctx *gin.Context) {
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

	err = services.DeleteMasoolReportsByModule(module)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	services.LogUserActivity(module, "delete_masool_report", "Deleted all Masool reports")
	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Masool reports deleted successfully", nil))
}

func GetPreviousMasoolReports(ctx *gin.Context) {
	masoolIDStr := ctx.Query("masool_id")
	if masoolIDStr == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "masool_id is required", nil))
		return
	}

	masoolID, err := strconv.ParseUint(masoolIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid masool_id", nil))
		return
	}

	data, err := services.GetPreviousMasoolReports(uint(masoolID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Previous masool reports fetched successfully", data))
}

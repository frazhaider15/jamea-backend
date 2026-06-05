package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	filters := make(map[string]string)
	for param, dataKey := range previousReportFilterKeys {
		if v := strings.TrimSpace(ctx.Query(param)); v != "" {
			filters[dataKey] = v
		}
	}
	if len(filters) > 1 {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "only one of province, division, district, tehsil, area may be passed at a time", nil))
		return
	}

	month := ctx.Query("month")
	data, err := services.GetMasoolReport(module, month, filters)
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

// previousReportFilterKeys maps query-param names to the actual key names
// stored inside masool.Data so callers can use lowercase params.
var previousReportFilterKeys = map[string]string{
	"province": "Province",
	"division": "Division",
	"district": "District",
	"tehsil":   "Tehsil",
	"area":     "Area",
}

func GetPreviousMasoolReports(ctx *gin.Context) {
	masoolIDStr := strings.TrimSpace(ctx.Query("masool_id"))

	filters := make(map[string]string)
	for param, dataKey := range previousReportFilterKeys {
		if v := strings.TrimSpace(ctx.Query(param)); v != "" {
			filters[dataKey] = v
		}
	}

	// Exactly one selector must be provided.
	provided := len(filters)
	if masoolIDStr != "" {
		provided++
	}
	if provided == 0 {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "exactly one of masool_id, province, division, district, tehsil, area is required", nil))
		return
	}
	if provided > 1 {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "only one of masool_id, province, division, district, tehsil, area may be passed at a time", nil))
		return
	}

	var masoolID uint
	if masoolIDStr != "" {
		parsed, err := strconv.ParseUint(masoolIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid masool_id", nil))
			return
		}
		masoolID = uint(parsed)
	}

	data, err := services.GetPreviousMasoolReports(masoolID, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Previous masool reports fetched successfully", data))
}

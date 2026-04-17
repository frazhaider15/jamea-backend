package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jamea/dto"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

func PostActivities(ctx *gin.Context) {
	var req dto.ActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid request body: "+err.Error(), nil))
		return
	}

	activity := models.Activity{
		Module:     req.Module,
		Month:      req.Month,
		Year:       req.Year,
		Activities: req.Activities,
	}

	data, err := services.SaveActivities(activity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, "failed to save activities: "+err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Activities saved successfully", data))
}

func GetActivities(ctx *gin.Context) {
	moduleStr := ctx.Query("module")
	monthStr := ctx.Query("month")
	yearStr := ctx.Query("year")

	if moduleStr == "" || monthStr == "" || yearStr == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "module, month, and year are required", nil))
		return
	}

	module, err := models.NewModuleString(moduleStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid module", nil))
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid month", nil))
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid year", nil))
		return
	}

	data, err := services.GetActivities(module, month, year)
	if err != nil {
		ctx.JSON(http.StatusNotFound, NewStandardResponse(false, 404, "activities not found: "+err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Activities fetched successfully", data))
}

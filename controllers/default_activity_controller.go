package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jamea/dto"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

func GetDefaultActivities(ctx *gin.Context) {
	data, err := services.GetDefaultActivities()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, "failed to fetch default activities: "+err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Default activities fetched successfully", data))
}

func AddDefaultActivity(ctx *gin.Context) {
	var req dto.AddDefaultActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid request body: "+err.Error(), nil))
		return
	}

	activity := models.DefaultActivity{
		Name:  strings.TrimSpace(req.Name),
		Color: strings.TrimSpace(req.Color),
	}

	if activity.Name == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "name is required", nil))
		return
	}

	data, err := services.AddDefaultActivity(activity)
	if err != nil {
		if err.Error() == "activity already exists" {
			ctx.JSON(http.StatusConflict, NewStandardResponse(false, 409, err.Error(), data))
			return
		}
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, "failed to add default activity: "+err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Default activity added successfully", data))
}

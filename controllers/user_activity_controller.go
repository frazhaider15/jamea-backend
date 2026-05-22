package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

const defaultLastActivitiesLimit = 10

// GetLastActivities handles GET /last-activities?module=MODULE&limit=N
// Returns the most recent actions performed by the user, newest first.
func GetLastActivities(ctx *gin.Context) {
	var module models.Module
	if moduleStr := ctx.Query("module"); moduleStr != "" {
		m, err := models.NewModuleString(moduleStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid module", nil))
			return
		}
		module = m
	}

	limit := defaultLastActivitiesLimit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "invalid limit", nil))
			return
		}
		limit = parsed
	}

	data, err := services.GetLastUserActivities(module, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewStandardResponse(false, 500, "failed to fetch last activities: "+err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Last activities fetched successfully", data))
}

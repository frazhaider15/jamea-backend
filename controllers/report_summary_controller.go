package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jamea/models"
	"github.com/jamea/services"
)

// GetMonthlyReportSummary handles
// GET /report-summary?module=MODULE&month=YYYY_MM[&province=...|&division=...|&district=...|&tehsil=...|&area=...]
//
// It returns the complete month-wise report: every masool of the module with
// the geography effective for them that month, their report (activities,
// remarks and remaining details), the module's planned activities, and a
// reported/not-reported tally. An optional single geographic filter narrows the
// masools to one area of the hierarchy.
func GetMonthlyReportSummary(ctx *gin.Context) {
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

	month := strings.TrimSpace(ctx.Query("month"))
	if month == "" {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, "month is required", nil))
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

	data, err := services.GetMonthlyReportSummary(module, month, filters)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, NewStandardResponse(false, 400, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, NewStandardResponse(true, 200, "Monthly report summary fetched successfully", data))
}

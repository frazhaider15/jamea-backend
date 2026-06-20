package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jamea/db"
	"github.com/jamea/models"
)

// monthRE matches an exact "YYYY_MM" month string.
var monthRE = regexp.MustCompile(`^(\d{4})_(\d{2})$`)

// MasoolMonthlyReport is one masool's slice of the monthly summary: who they
// are, the geography effective for them that month, whether they filed a report,
// and that report's activities, remarks and remaining details.
type MasoolMonthlyReport struct {
	ID         uint                   `json:"id"`
	Name       string                 `json:"name"`
	Location   map[string]string      `json:"location"`
	HasReport  bool                   `json:"has_report"`
	Activities []interface{}          `json:"activities"`
	Remarks    []interface{}          `json:"remarks"`
	Report     map[string]interface{} `json:"report"`
}

// ReportSummaryCounts is the headline tally for a month.
type ReportSummaryCounts struct {
	TotalMasools int `json:"total_masools"`
	Reported     int `json:"reported"`
	NotReported  int `json:"not_reported"`
}

// MonthlyReportSummary is the complete month-wise report: every masool of a
// module with their location and report, the module's planned activities for the
// month, and the reported/not-reported tally.
type MonthlyReportSummary struct {
	Month            string                `json:"month"`
	Module           string                `json:"module"`
	Summary          ReportSummaryCounts   `json:"summary"`
	ModuleActivities []string              `json:"module_activities"`
	Masools          []MasoolMonthlyReport `json:"masools"`
}

// GetMonthlyReportSummary builds a complete month-wise report for a module.
//
// For the given month ("YYYY_MM") it lists every masool of the module, resolves
// the geography effective for each masool that month (from its location history,
// so a later reassignment never rewrites the past), and attaches the report they
// filed that month — broken out into activities, remarks and the remaining
// report details. Masools that filed no report are still listed, with
// has_report=false, so the summary doubles as a reporting checklist. It also
// surfaces the module-level planned activities for the month.
//
// When filters is non-empty (a single geographic level -> value, case-insensitive)
// only masools whose resolved geography for the month matches the hierarchy are
// included.
func GetMonthlyReportSummary(module models.Module, month string, filters map[string]string) (*MonthlyReportSummary, error) {
	m := monthRE.FindStringSubmatch(strings.TrimSpace(month))
	if m == nil {
		return nil, fmt.Errorf("invalid month %q: expected YYYY_MM", month)
	}
	yearInt, _ := strconv.Atoi(m[1])
	monthInt, _ := strconv.Atoi(m[2])

	// 1. Every masool of the module.
	var masools []models.Masool
	if err := db.DB.Where("module = ?", module).Order("id asc").Find(&masools).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masools: %v", err)
	}

	// 2. The reports filed that month, keeping the latest per masool.
	var reports []models.MasoolReport
	if err := db.DB.Where("module = ? AND month = ?", module, month).Order("id desc").Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masool reports: %v", err)
	}
	reportByMasool := make(map[uint]models.MasoolReport, len(reports))
	for _, r := range reports {
		if _, seen := reportByMasool[r.MasoolID]; !seen {
			reportByMasool[r.MasoolID] = r
		}
	}

	// 3. Location history for those masools, grouped by masool id, so geography
	//    can be resolved for the requested month rather than the current area.
	var locs []models.MasoolLocation
	if err := db.DB.Where("module = ?", module).Find(&locs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masool locations: %v", err)
	}
	locsByMasool := make(map[uint][]models.MasoolLocation)
	for _, l := range locs {
		locsByMasool[l.MasoolID] = append(locsByMasool[l.MasoolID], l)
	}

	// 4. The module-level planned activities for the month (at most one row).
	var moduleActivities []string
	acts, err := GetActivities(module, monthInt, yearInt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %v", err)
	}
	for _, a := range acts {
		moduleActivities = append(moduleActivities, a.Activities...)
	}

	// 5. Assemble one entry per masool.
	summary := &MonthlyReportSummary{
		Month:            month,
		Module:           module.String(),
		ModuleActivities: moduleActivities,
		Masools:          []MasoolMonthlyReport{},
	}

	for _, masool := range masools {
		geo := resolveMasoolGeo(locsByMasool[masool.ID], month, masool)

		// Apply the optional geographic filter against the month's geography.
		if len(filters) > 0 && !geoMapMatchesHierarchy(geo, filters) {
			continue
		}

		entry := MasoolMonthlyReport{
			ID:         masool.ID,
			Name:       masool.Name,
			Location:   geo,
			Activities: []interface{}{},
			Remarks:    []interface{}{},
			Report:     map[string]interface{}{},
		}

		if report, ok := reportByMasool[masool.ID]; ok {
			entry.HasReport = true
			entry.Activities = toList(reportDataValue(report.Data, "Activities"))
			entry.Remarks = toList(reportDataValue(report.Data, "Remarks"))
			entry.Report = flattenReportDetails(report.Data)
		}

		summary.Masools = append(summary.Masools, entry)
		if entry.HasReport {
			summary.Summary.Reported++
		}
	}

	summary.Summary.TotalMasools = len(summary.Masools)
	summary.Summary.NotReported = summary.Summary.TotalMasools - summary.Summary.Reported

	return summary, nil
}

// reportDataValue returns the value stored under key (case-insensitive) in a
// report's Data, or nil when the key is absent.
func reportDataValue(data []models.MasoolData, key string) interface{} {
	for _, d := range data {
		if strings.EqualFold(strings.TrimSpace(d.Key), key) {
			return d.Val
		}
	}
	return nil
}

// flattenReportDetails turns a report's Data into a lowercase key -> value map,
// dropping the fields surfaced elsewhere (id/name and the broken-out
// activities/remarks) so Report holds only the remaining details.
func flattenReportDetails(data []models.MasoolData) map[string]interface{} {
	out := make(map[string]interface{})
	for _, d := range data {
		key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
		switch key {
		case "", "id", "name", "activities", "remarks":
			continue
		}
		out[key] = d.Val
	}
	return out
}

// toList normalises a stored value into a JSON array so activities/remarks are
// always arrays for the client: nil -> [], a slice is passed through, and any
// scalar is wrapped in a single-element array.
func toList(v interface{}) []interface{} {
	switch vv := v.(type) {
	case nil:
		return []interface{}{}
	case []interface{}:
		return vv
	case []string:
		out := make([]interface{}, len(vv))
		for i, s := range vv {
			out[i] = s
		}
		return out
	default:
		return []interface{}{vv}
	}
}

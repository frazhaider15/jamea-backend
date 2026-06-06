package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jamea/db"
	"github.com/jamea/models"
)

// monthRangeRE matches "YYYY_MM-YYYY_MM"; the two capture groups are the
// inclusive start and end months of a range filter.
var monthRangeRE = regexp.MustCompile(`^(\d{4}_\d{2})-(\d{4}_\d{2})$`)

// parseMonthFilter inspects a month query value. If it matches the range form
// "YYYY_MM-YYYY_MM" it returns the two endpoints with isRange=true; otherwise
// it returns isRange=false and the caller should treat the value as an exact
// match (existing behaviour).
func parseMonthFilter(s string) (start, end string, isRange bool) {
	m := monthRangeRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

func UploadMasoolReport(file multipart.File, module models.Module) ([]models.MasoolReport, error) {
	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Find the index of the "ID" column
	idIndex := -1
	for i, h := range headers {
		if strings.TrimSpace(strings.ToLower(h)) == "id" {
			idIndex = i
			break
		}
	}
	if idIndex == -1 {
		return nil, fmt.Errorf("CSV must contain an 'ID' column for masool_id")
	}

	var uploaded []models.MasoolReport
	var currentReport *models.MasoolReport
	currentMonth := time.Now().Format("2006_01")

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %v", err)
		}

		// Skip rows where all values are empty
		allEmpty := true
		for _, val := range record {
			if strings.TrimSpace(val) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		idStr := strings.TrimSpace(record[idIndex])
		if idStr != "" {
			// Save the previous report if it exists
			if currentReport != nil {
				if err := db.DB.Create(currentReport).Error; err != nil {
					return nil, fmt.Errorf("failed to save masool report to db: %v", err)
				}
				uploaded = append(uploaded, *currentReport)
			}

			// Parse masool_id
			masoolID, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid masool ID '%s': must be a valid number", idStr)
			}

			// Validate masool exists
			var masool models.Masool
			if err := db.DB.First(&masool, masoolID).Error; err != nil {
				return nil, fmt.Errorf("masool with id %d not found", masoolID)
			}

			// Start a new report
			currentReport = &models.MasoolReport{
				MasoolID: uint(masoolID),
				Module:   module,
				Month:    currentMonth,
				Data:     make([]models.MasoolData, 0),
			}

			// Add initial data
			for i, val := range record {
				val = strings.TrimSpace(val)
				if val == "" || i == idIndex {
					continue
				}
				key := strings.TrimSpace(headers[i])
				if key != "" {
					currentReport.Data = append(currentReport.Data, models.MasoolData{
						Key: key,
						Val: val,
					})
				}
			}
		} else if currentReport != nil {
			// Continuation row
			for i, val := range record {
				val = strings.TrimSpace(val)
				if val == "" || i == idIndex {
					continue
				}
				key := strings.TrimSpace(headers[i])
				if key == "" {
					continue
				}

				// Find if the key already exists
				found := false
				for j := range currentReport.Data {
					if strings.EqualFold(currentReport.Data[j].Key, key) {
						switch existingVal := currentReport.Data[j].Val.(type) {
						case []string:
							currentReport.Data[j].Val = append(existingVal, val)
						case string:
							currentReport.Data[j].Val = []string{existingVal, val}
						default:
							currentReport.Data[j].Val = []string{fmt.Sprint(existingVal), val}
						}
						found = true
						break
					}
				}

				if !found {
					currentReport.Data = append(currentReport.Data, models.MasoolData{
						Key: key,
						Val: val,
					})
				}
			}
		}
	}

	// Save the last report
	if currentReport != nil {
		if err := db.DB.Create(currentReport).Error; err != nil {
			return nil, fmt.Errorf("failed to save masool report to db: %v", err)
		}
		uploaded = append(uploaded, *currentReport)
	}

	return uploaded, nil
}

// GetMasoolReport fetches all masools along with their associated report data.
// It matches masool.id with masool_reports.masool_id and flattens the JSONB
// key-val data from both tables into a single flat response per report.
// When filters is non-empty, only masools whose Data matches every key/val
// pair (case-insensitive) are included.
func GetMasoolReport(module models.Module, month string, filters map[string]string) ([]map[string]interface{}, error) {
	// 1. Fetch all masools for the given module
	var masools []models.Masool
	if err := db.DB.Where("module = ?", module).Find(&masools).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masools: %v", err)
	}

	// 2. Apply geographic filters, then build a lookup map: masool ID -> Masool
	masoolMap := make(map[uint]models.Masool)
	for _, m := range masools {
		if len(filters) > 0 && !masoolMatchesFilters(m, filters) {
			continue
		}
		masoolMap[m.ID] = m
	}
	if len(masoolMap) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 3. Fetch all masool reports for the given module
	var reports []models.MasoolReport
	query := db.DB.Where("module = ?", module).Order("id desc")
	if month != "" {
		if start, end, isRange := parseMonthFilter(month); isRange {
			query = query.Where("month >= ?", start).Where("month <= ?", end)
		} else {
			query = query.Where("month = ?", month)
		}
	}
	if err := query.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masool reports: %v", err)
	}

	// 4. Join and flatten the data
	var result []map[string]interface{}
	processedMasools := make(map[uint]bool)

	for _, report := range reports {
		masool, exists := masoolMap[report.MasoolID]
		if !exists {
			continue // skip reports with no matching masool
		}

		if processedMasools[report.MasoolID] {
			continue // skip if we already added the latest report for this masool
		}
		processedMasools[report.MasoolID] = true

		entry := make(map[string]interface{})
		entry["id"] = masool.ID
		entry["name"] = masool.Name
		entry["month"] = report.Month

		// Add masool's key-val data
		for _, d := range masool.Data {
			key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if key != "" && key != "name" && key != "id" {
				entry[key] = d.Val
			}
		}

		// Add/override with report's key-val data
		for _, d := range report.Data {
			key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if key != "" {
				entry[key] = d.Val
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

func DeleteMasoolReportsByModule(module models.Module) error {
	return db.DB.Where("module = ?", module).Delete(&models.MasoolReport{}).Error
}

// GetPreviousMasoolReports fetches reports for previous months of the same year.
// Selectors are mutually exclusive (the controller enforces exactly one):
//   - masoolID > 0  : restrict to that masool
//   - filters       : a single geographic level -> value (case-insensitive);
//                     reports are matched by the geography that was effective for
//                     each report's own month, not the masool's current area.
//
// Geography is resolved per report month from the masool's location history, so
// reassigning a masool to a new area never rewrites the location of its past
// reports.
func GetPreviousMasoolReports(masoolID uint, filters map[string]string) ([]map[string]interface{}, error) {
	currentMonth := time.Now().Format("2006_01")
	parts := strings.Split(currentMonth, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid month format, expected YYYY_MM")
	}
	year := parts[0]

	// 1. Gather candidate reports for the current year, up to the current month.
	query := db.DB.Where("month LIKE ?", year+"_%").Where("month <= ?", currentMonth).Order("id desc")
	switch {
	case masoolID != 0:
		var m models.Masool
		if err := db.DB.First(&m, masoolID).Error; err != nil {
			return nil, fmt.Errorf("masool with id %d not found: %v", masoolID, err)
		}
		query = query.Where("masool_id = ?", masoolID)
	case len(filters) > 0:
		// Keep every report for the year; each is filtered below by the geography
		// effective for its own month.
	default:
		return nil, fmt.Errorf("masool_id or at least one filter is required")
	}

	var reports []models.MasoolReport
	if err := query.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch previous reports: %v", err)
	}
	if len(reports) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 2. Load the masools referenced by those reports (name + non-geographic data).
	idSet := make(map[uint]struct{}, len(reports))
	for _, r := range reports {
		idSet[r.MasoolID] = struct{}{}
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	var masools []models.Masool
	if err := db.DB.Where("id IN ?", ids).Find(&masools).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masools: %v", err)
	}
	masoolMap := make(map[uint]models.Masool, len(masools))
	for _, m := range masools {
		masoolMap[m.ID] = m
	}

	// 3. Load location history for those masools, grouped by masool id.
	var locs []models.MasoolLocation
	if err := db.DB.Where("masool_id IN ?", ids).Find(&locs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masool locations: %v", err)
	}
	locsByMasool := make(map[uint][]models.MasoolLocation)
	for _, l := range locs {
		locsByMasool[l.MasoolID] = append(locsByMasool[l.MasoolID], l)
	}

	// 4. Build the response, resolving each report's geography by its own month
	//    and deduping by (masoolID, month) — keep the latest per pair.
	type monthKey struct {
		masoolID uint
		month    string
	}
	processed := make(map[monthKey]bool)
	result := []map[string]interface{}{}

	for _, report := range reports {
		masool, ok := masoolMap[report.MasoolID]
		if !ok {
			continue
		}

		geo := resolveMasoolGeo(locsByMasool[report.MasoolID], report.Month, masool)

		// Filter against the geography effective for this report's month.
		if len(filters) > 0 && !geoMapMatchesHierarchy(geo, filters) {
			continue
		}

		k := monthKey{report.MasoolID, report.Month}
		if processed[k] {
			continue
		}
		processed[k] = true

		entry := make(map[string]interface{})
		entry["id"] = masool.ID
		entry["name"] = masool.Name
		entry["month"] = report.Month

		// Masool's non-geographic data first.
		for _, d := range masool.Data {
			kk := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if kk != "" && kk != "name" && kk != "id" {
				entry[kk] = d.Val
			}
		}
		// Override the geographic levels with the values effective for this month.
		for _, level := range geoHierarchy {
			entry[level] = geo[level]
		}
		// Report's own data takes final precedence.
		for _, d := range report.Data {
			kk := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if kk != "" {
				entry[kk] = d.Val
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

// geoHierarchy lists the geographic levels from broadest to narrowest.
// A masool selected at a given level must have every level below it empty, so
// that e.g. a province search returns only province-level masools and not the
// districts/tehsils/areas that also carry that province name.
var geoHierarchy = []string{"province", "division", "district", "tehsil", "area"}

// geoLevelIndex returns the position of key within geoHierarchy, or -1 if key
// is not a geographic level.
func geoLevelIndex(key string) int {
	for i, level := range geoHierarchy {
		if level == key {
			return i
		}
	}
	return -1
}

// flattenMasoolData turns a masool's Data into a lowercase key -> string value
// lookup so filters can be matched case-insensitively.
func flattenMasoolData(masool models.Masool) map[string]string {
	dataMap := make(map[string]string, len(masool.Data))
	for _, d := range masool.Data {
		key := strings.ToLower(strings.TrimSpace(d.Key))
		if key == "" {
			continue
		}
		switch v := d.Val.(type) {
		case string:
			dataMap[key] = v
		default:
			dataMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return dataMap
}

// geoMapMatchesHierarchy is the hierarchy-aware matcher used by the
// previous-reports endpoint. geo is a resolved lowercase level -> value map
// (see resolveMasoolGeo). The filtered geographic level must equal the requested
// value (case-insensitive) AND every level below it in the
// province > division > district > tehsil > area hierarchy must be empty.
// Levels above the filtered one may hold any value. Non-geographic filter keys
// fall back to a plain equality match.
func geoMapMatchesHierarchy(geo map[string]string, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}

	for fk, fv := range filters {
		lfk := strings.ToLower(fk)

		// The filtered level itself must match the requested value.
		got, ok := geo[lfk]
		if !ok || !strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(fv)) {
			return false
		}

		// For geographic filters, every level below the filtered one must be empty.
		if level := geoLevelIndex(lfk); level >= 0 {
			for _, lower := range geoHierarchy[level+1:] {
				if strings.TrimSpace(geo[lower]) != "" {
					return false
				}
			}
		}
	}
	return true
}

// masoolMatchesFilters returns true when the masool's Data contains every
// key/value pair in filters. Key and value matching is case-insensitive so
// callers can pass lowercase query params against capitalised stored keys.
func masoolMatchesFilters(masool models.Masool, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	dataMap := flattenMasoolData(masool)
	for fk, fv := range filters {
		got, ok := dataMap[strings.ToLower(fk)]
		if !ok || !strings.EqualFold(got, fv) {
			return false
		}
	}
	return true
}

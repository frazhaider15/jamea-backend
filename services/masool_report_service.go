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
// Selectors (one or more, mutually combinable):
//   - masoolID > 0  : restrict to that masool
//   - filters       : map of masool.Data key -> value (case-insensitive match);
//                     only masools whose Data contains ALL given key/val pairs match
//
// If only filters are given, every matching masool's previous reports are returned.
func GetPreviousMasoolReports(masoolID uint, filters map[string]string) ([]map[string]interface{}, error) {
	currentMonth := time.Now().Format("2006_01")
	parts := strings.Split(currentMonth, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid month format, expected YYYY_MM")
	}
	year := parts[0]

	// 1. Resolve the set of masools to include.
	var masools []models.Masool
	switch {
	case masoolID != 0:
		var m models.Masool
		if err := db.DB.First(&m, masoolID).Error; err != nil {
			return nil, fmt.Errorf("masool with id %d not found: %v", masoolID, err)
		}
		if len(filters) > 0 && !masoolMatchesFilters(m, filters) {
			return []map[string]interface{}{}, nil
		}
		masools = []models.Masool{m}
	case len(filters) > 0:
		var all []models.Masool
		if err := db.DB.Find(&all).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch masools: %v", err)
		}
		for _, m := range all {
			if masoolMatchesFilters(m, filters) {
				masools = append(masools, m)
			}
		}
		if len(masools) == 0 {
			return []map[string]interface{}{}, nil
		}
	default:
		return nil, fmt.Errorf("masool_id or at least one filter is required")
	}

	// 2. Fetch all previous reports for the selected masools.
	masoolMap := make(map[uint]models.Masool, len(masools))
	ids := make([]uint, 0, len(masools))
	for _, m := range masools {
		masoolMap[m.ID] = m
		ids = append(ids, m.ID)
	}

	var reports []models.MasoolReport
	if err := db.DB.Where("masool_id IN ?", ids).
		Where("month LIKE ?", year+"_%").
		Where("month <= ?", currentMonth).
		Order("id desc").
		Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch previous reports: %v", err)
	}

	// 3. Flatten, deduping by (masoolID, month) — keep the latest per pair.
	type monthKey struct {
		masoolID uint
		month    string
	}
	processed := make(map[monthKey]bool)
	var result []map[string]interface{}

	for _, report := range reports {
		masool, ok := masoolMap[report.MasoolID]
		if !ok {
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

		for _, d := range masool.Data {
			kk := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if kk != "" && kk != "name" && kk != "id" {
				entry[kk] = d.Val
			}
		}
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

// masoolMatchesFilters returns true when the masool's Data contains every
// key/value pair in filters. Key and value matching is case-insensitive so
// callers can pass lowercase query params against capitalised stored keys.
func masoolMatchesFilters(masool models.Masool, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
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
	for fk, fv := range filters {
		got, ok := dataMap[strings.ToLower(fk)]
		if !ok || !strings.EqualFold(got, fv) {
			return false
		}
	}
	return true
}

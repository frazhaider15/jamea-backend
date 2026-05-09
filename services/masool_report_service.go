package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/jamea/db"
	"github.com/jamea/models"
)

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
func GetMasoolReport(module models.Module, month string) ([]map[string]interface{}, error) {
	// 1. Fetch all masools for the given module
	var masools []models.Masool
	if err := db.DB.Where("module = ?", module).Find(&masools).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch masools: %v", err)
	}

	// 2. Build a lookup map: masool ID -> Masool
	masoolMap := make(map[uint]models.Masool)
	for _, m := range masools {
		masoolMap[m.ID] = m
	}

	// 3. Fetch all masool reports for the given module
	var reports []models.MasoolReport
	query := db.DB.Where("module = ?", module).Order("id desc")
	if month != "" {
		query = query.Where("month = ?", month)
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

// GetPreviousMasoolReports fetches reports for previous months of the same year for a specific masool
func GetPreviousMasoolReports(masoolID uint) ([]map[string]interface{}, error) {
	currentMonth := time.Now().Format("2006_01")
	parts := strings.Split(currentMonth, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid month format, expected YYYY_MM")
	}
	year := parts[0]

	// 1. Fetch the specific masool
	var masool models.Masool
	if err := db.DB.First(&masool, masoolID).Error; err != nil {
		return nil, fmt.Errorf("masool with id %d not found: %v", masoolID, err)
	}

	// 2. Fetch all reports for this year strictly less than currentMonth
	var reports []models.MasoolReport
	query := db.DB.Where("masool_id = ?", masoolID).
		Where("month LIKE ?", year+"_%").
		Where("month <= ?", currentMonth).
		Order("id desc")

	if err := query.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch previous reports: %v", err)
	}

	// 3. Join and flatten
	var result []map[string]interface{}
	processedMonths := make(map[string]bool)

	for _, report := range reports {
		if processedMonths[report.Month] {
			continue // skip if we already added the latest report for this month
		}
		processedMonths[report.Month] = true

		entry := make(map[string]interface{})
		entry["id"] = masool.ID
		entry["name"] = masool.Name
		entry["month"] = report.Month

		// Add masool's key-val data
		for _, d := range masool.Data {
			k := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if k != "" && k != "name" && k != "id" {
				entry[k] = d.Val
			}
		}

		// Add/override with report's key-val data
		for _, d := range report.Data {
			k := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d.Key), " ", "_"))
			if k != "" {
				entry[k] = d.Val
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

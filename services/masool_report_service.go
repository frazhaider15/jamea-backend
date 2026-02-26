package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/jamea/db"
	"github.com/jamea/models"
)

func UploadMasoolReport(file multipart.File, masoolID uint, module models.Module) ([]models.MasoolReport, error) {
	// Validate that the masool exists
	var masool models.Masool
	if err := db.DB.First(&masool, masoolID).Error; err != nil {
		return nil, fmt.Errorf("masool with id %d not found", masoolID)
	}

	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	var uploaded []models.MasoolReport

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

		report := models.MasoolReport{
			MasoolID: masoolID,
			Module:   module,
			Data:     make([]models.MasoolData, 0),
		}

		for i, val := range record {
			val = strings.TrimSpace(val)
			key := strings.TrimSpace(headers[i])
			if i < len(headers) && key != "" {
				report.Data = append(report.Data, models.MasoolData{
					Key: key,
					Val: val,
				})
			}
		}

		if err := db.DB.Create(&report).Error; err != nil {
			return nil, fmt.Errorf("failed to save masool report to db: %v", err)
		}

		uploaded = append(uploaded, report)
	}

	return uploaded, nil
}

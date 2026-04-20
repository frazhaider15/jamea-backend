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

func UploadMasool(file multipart.File, module models.Module) ([]models.Masool, error) {
	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Find the index of the "Name" column
	nameIndex := -1
	for i, h := range headers {
		if strings.EqualFold(h, "Name") {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		return nil, fmt.Errorf("CSV must contain a 'Name' column")
	}

	var uploaded []models.Masool

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %v", err)
		}

		name := strings.TrimSpace(record[nameIndex])
		if name == "" {
			continue // Skip records with empty name
		}

		masool := models.Masool{
			Name:   name,
			Module: module, // Set the module
			Data:   make([]models.MasoolData, 0),
		}

		for i, val := range record {
			val = strings.TrimSpace(val)
			key := strings.TrimSpace(headers[i])
			if i < len(headers) && key != "" {
				masool.Data = append(masool.Data, models.MasoolData{
					Key: key,
					Val: val,
				})
			}
		}

		if err := db.DB.Create(&masool).Error; err != nil {
			return nil, fmt.Errorf("failed to save masool to db: %v", err)
		}

		uploaded = append(uploaded, masool)
	}

	return uploaded, nil
}
func DeleteMasoolsByModule(module models.Module) error {
	return db.DB.Where("module = ?", module).Delete(&models.Masool{}).Error
}

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

// GetMasoolsByModule fetches all masools for a given module
func GetMasoolsByModule(module models.Module) ([]models.Masool, error) {
	var masools []models.Masool
	err := db.DB.Where("module = ?", module).Find(&masools).Error
	return masools, err
}

func UploadMasool(file multipart.File, module models.Module) ([]models.Masool, error) {
	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Find the index of the \"Name\" column and \"Longitude\" column
	nameIndex := -1
	longitudeIndex := -1
	for i, h := range headers {
		hTrimmed := strings.TrimSpace(h)
		if strings.EqualFold(hTrimmed, "Name") {
			nameIndex = i
		}
		if strings.EqualFold(hTrimmed, "Longitude") {
			longitudeIndex = i
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

		// Ensure we don't out-of-bounds if the record is short
		if len(record) <= nameIndex {
			continue
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
			if longitudeIndex != -1 && i > longitudeIndex {
				break
			}
			if i >= len(headers) {
				break
			}
			val = strings.TrimSpace(val)
			key := strings.TrimSpace(headers[i])
			if key != "" {
				masool.Data = append(masool.Data, models.MasoolData{
					Key: key,
					Val: val,
				})
			}
		}

		if err := db.DB.Create(&masool).Error; err != nil {
			return nil, fmt.Errorf("failed to save masool to db: %v", err)
		}
		if err := RecordMasoolLocation(masool); err != nil {
			return nil, fmt.Errorf("failed to record masool location: %v", err)
		}

		uploaded = append(uploaded, masool)
	}

	return uploaded, nil
}

func DeleteMasoolsByModule(module models.Module) error {
	return db.DB.Where("module = ?", module).Delete(&models.Masool{}).Error
}

// CreateMasool creates a new masool
func CreateMasool(masool *models.Masool) error {
	if err := db.DB.Create(masool).Error; err != nil {
		return err
	}
	return RecordMasoolLocation(*masool)
}

// GetMasoolByID fetches a masool by its ID
func GetMasoolByID(id uint) (*models.Masool, error) {
	var masool models.Masool
	if err := db.DB.First(&masool, id).Error; err != nil {
		return nil, err
	}
	return &masool, nil
}

// UpdateMasoolByID updates an existing masool's data by its ID
func UpdateMasoolByID(id uint, data []models.MasoolData) (*models.Masool, error) {
	var masool models.Masool
	if err := db.DB.First(&masool, id).Error; err != nil {
		return nil, err
	}
	masool.Data = data
	if err := db.DB.Save(&masool).Error; err != nil {
		return nil, err
	}
	if err := RecordMasoolLocation(masool); err != nil {
		return nil, err
	}
	return &masool, nil
}

// DeleteMasoolByID deletes a masool by its ID
func DeleteMasoolByID(id uint) error {
	return db.DB.Delete(&models.Masool{}, id).Error
}

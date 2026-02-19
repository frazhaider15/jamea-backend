package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/jamea/models"
	"github.com/jamea/store"
)

func UploadMasool(file multipart.File) error {
	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Find the index of the "Name" column
	nameIndex := -1
	for i, h := range headers {
		if h == "Name" {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		return fmt.Errorf("CSV must contain a 'Name' column")
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record: %v", err)
		}

		masool := models.Masool{
			Name: record[nameIndex],
			Data: make([]models.MasoolData, 0),
		}

		for i, val := range record {
			// Add all columns to the data list, including Name (as implied by "key data")
			// or we can choose to exclude it. Since the prompt said "additionally store Masool name seperately",
			// it implies it might be in the data too or just extracted.
			// Storing everything in Data is safer for full context.
			if i < len(headers) {
				masool.Data = append(masool.Data, models.MasoolData{
					Key: headers[i],
					Val: val,
				})
			}
		}

		store.AddMasool(masool)
	}

	return nil
}

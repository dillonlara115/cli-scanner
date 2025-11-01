package exporter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dillonlara115/baracuda/pkg/models"
)

// ExportJSON exports page results to a JSON file
func ExportJSON(results []*models.PageResult, filePath string, pretty bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}


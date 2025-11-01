package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dillonlara115/baracuda/pkg/models"
)

// ExportCSV exports page results to a CSV file
func ExportCSV(results []*models.PageResult, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"URL",
		"Status Code",
		"Response Time (ms)",
		"Title",
		"Meta Description",
		"Canonical",
		"H1",
		"H2",
		"H3",
		"H4",
		"H5",
		"H6",
		"Internal Links",
		"External Links",
		"Redirect Chain",
		"Error",
		"Crawled At",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, result := range results {
		row := []string{
			result.URL,
			strconv.Itoa(result.StatusCode),
			strconv.FormatInt(result.ResponseTime, 10),
			result.Title,
			result.MetaDesc,
			result.Canonical,
			strings.Join(result.H1, " | "),
			strings.Join(result.H2, " | "),
			strings.Join(result.H3, " | "),
			strings.Join(result.H4, " | "),
			strings.Join(result.H5, " | "),
			strings.Join(result.H6, " | "),
			strings.Join(result.InternalLinks, " | "),
			strings.Join(result.ExternalLinks, " | "),
			strings.Join(result.RedirectChain, " -> "),
			result.Error,
			result.CrawledAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}


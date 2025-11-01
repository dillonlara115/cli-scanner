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

// ImportCSV imports page results from a CSV file
func ImportCSV(filePath string) ([]*models.PageResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Parse header
	header := records[0]
	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	results := make([]*models.PageResult, 0, len(records)-1)

	// Parse data rows
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 {
			continue
		}

		result := &models.PageResult{
			H1:            make([]string, 0),
			H2:            make([]string, 0),
			H3:            make([]string, 0),
			H4:            make([]string, 0),
			H5:            make([]string, 0),
			H6:            make([]string, 0),
			InternalLinks: make([]string, 0),
			ExternalLinks: make([]string, 0),
			Images:        make([]models.Image, 0),
			RedirectChain: make([]string, 0),
		}

		// Helper to get field value safely
		getField := func(name string) string {
			if idx, ok := headerMap[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		// Parse fields
		result.URL = getField("url")
		if result.URL == "" {
			continue // Skip rows without URL
		}

		// Status code
		if statusStr := getField("status code"); statusStr != "" {
			if status, err := strconv.Atoi(statusStr); err == nil {
				result.StatusCode = status
			}
		}

		// Response time
		if timeStr := getField("response time (ms)"); timeStr != "" {
			if timeMs, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
				result.ResponseTime = timeMs
			}
		}

		// Simple fields
		result.Title = getField("title")
		result.MetaDesc = getField("meta description")
		result.Canonical = getField("canonical")
		result.Error = getField("error")

		// Parse array fields (pipe-separated)
		if h1Str := getField("h1"); h1Str != "" {
			result.H1 = strings.Split(h1Str, " | ")
		}
		if h2Str := getField("h2"); h2Str != "" {
			result.H2 = strings.Split(h2Str, " | ")
		}
		if h3Str := getField("h3"); h3Str != "" {
			result.H3 = strings.Split(h3Str, " | ")
		}
		if h4Str := getField("h4"); h4Str != "" {
			result.H4 = strings.Split(h4Str, " | ")
		}
		if h5Str := getField("h5"); h5Str != "" {
			result.H5 = strings.Split(h5Str, " | ")
		}
		if h6Str := getField("h6"); h6Str != "" {
			result.H6 = strings.Split(h6Str, " | ")
		}
		if internalStr := getField("internal links"); internalStr != "" {
			result.InternalLinks = strings.Split(internalStr, " | ")
		}
		if externalStr := getField("external links"); externalStr != "" {
			result.ExternalLinks = strings.Split(externalStr, " | ")
		}
		if redirectStr := getField("redirect chain"); redirectStr != "" {
			result.RedirectChain = strings.Split(redirectStr, " -> ")
		}

		// Parse crawled at timestamp
		if crawledStr := getField("crawled at"); crawledStr != "" {
			if t, err := time.Parse(time.RFC3339, crawledStr); err == nil {
				result.CrawledAt = t
			} else {
				// Try other common formats
				for _, layout := range []string{
					time.RFC3339Nano,
					"2006-01-02 15:04:05",
					"2006-01-02T15:04:05Z07:00",
				} {
					if t, err := time.Parse(layout, crawledStr); err == nil {
						result.CrawledAt = t
						break
					}
				}
			}
		} else {
			result.CrawledAt = time.Now()
		}

		results = append(results, result)
	}

	return results, nil
}


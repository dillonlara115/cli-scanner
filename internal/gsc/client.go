package gsc

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/searchconsole/v1"
	
	"github.com/dillonlara115/barracuda/internal/analyzer"
	"github.com/dillonlara115/barracuda/pkg/models"
)

// EnrichedIssue extends analyzer.Issue with GSC performance data
type EnrichedIssue struct {
	Issue              analyzer.Issue           `json:"issue"`
	GSCPerformance     *models.GSCPerformance  `json:"gsc_performance,omitempty"`
	EnrichedPriority   float64                 `json:"enriched_priority"`
	RecommendationReason string                `json:"recommendation_reason"`
}

// FetchPerformanceData fetches Search Analytics data for a property
func FetchPerformanceData(userID string, siteURL string, startDate, endDate time.Time) (map[string]*models.GSCPerformance, error) {
	service, err := GetService(userID)
	if err != nil {
		return nil, err
	}

	// Request Search Analytics data
	request := &searchconsole.SearchAnalyticsQueryRequest{
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
		Dimensions: []string{"page"},
		RowLimit:   25000, // Max allowed by API
	}

	response, err := service.Searchanalytics.Query(siteURL, request).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query search analytics: %w", err)
	}

	// Convert to our model
	performanceMap := make(map[string]*models.GSCPerformance)
	
	for _, row := range response.Rows {
		url := row.Keys[0] // First dimension is "page"
		
		// Normalize URL to match crawl results
		normalizedURL := normalizeURL(url)
		
		performanceMap[normalizedURL] = &models.GSCPerformance{
			URL:          normalizedURL,
			Impressions: int64(row.Impressions),
			Clicks:       int64(row.Clicks),
			CTR:          row.Ctr,
			Position:    row.Position,
			LastUpdated:  time.Now(),
		}
	}

	// Fetch query data for top pages
	if err := fetchQueryData(userID, siteURL, startDate, endDate, performanceMap); err != nil {
		// Log error but don't fail - query data is optional
		fmt.Printf("Warning: Failed to fetch query data: %v\n", err)
	}

	return performanceMap, nil
}

// fetchQueryData fetches top queries for pages
func fetchQueryData(userID string, siteURL string, startDate, endDate time.Time, performanceMap map[string]*models.GSCPerformance) error {
	service, err := GetService(userID)
	if err != nil {
		return err
	}

	// For each page with significant traffic, fetch top queries
	// Note: This is a simplified version - in production, you might want to batch this
	for url, perf := range performanceMap {
		if perf.Impressions < 100 {
			continue // Skip low-traffic pages
		}

		request := &searchconsole.SearchAnalyticsQueryRequest{
			StartDate:  startDate.Format("2006-01-02"),
			EndDate:    endDate.Format("2006-01-02"),
			Dimensions: []string{"query"},
			DimensionFilterGroups: []*searchconsole.ApiDimensionFilterGroup{
				{
					Filters: []*searchconsole.ApiDimensionFilter{
						{
							Dimension:  "page",
							Expression:  url,
							Operator:   "equals",
						},
					},
				},
			},
			RowLimit: 10, // Top 10 queries
		}

		response, err := service.Searchanalytics.Query(siteURL, request).Do()
		if err != nil {
			continue // Skip if query fails
		}

		queries := make([]models.Query, 0, len(response.Rows))
		for _, row := range response.Rows {
			queries = append(queries, models.Query{
				Query:       row.Keys[0],
				Impressions: int64(row.Impressions),
				Clicks:      int64(row.Clicks),
				CTR:         row.Ctr,
				Position:    row.Position,
			})
		}
		perf.TopQueries = queries
	}

	return nil
}

// normalizeURL normalizes URLs to match crawl results
func normalizeURL(url string) string {
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")
	// Ensure lowercase
	url = strings.ToLower(url)
	return url
}

// EnrichIssues merges GSC performance data with issues
func EnrichIssues(issues []analyzer.Issue, performanceMap map[string]*models.GSCPerformance) []EnrichedIssue {
	enriched := make([]EnrichedIssue, 0, len(issues))

	for _, issue := range issues {
		enrichedIssue := EnrichedIssue{
			Issue: issue,
		}

		// Normalize issue URL to match GSC data
		normalizedURL := normalizeURL(issue.URL)
		
		// Find matching performance data
		if perf, exists := performanceMap[normalizedURL]; exists {
			enrichedIssue.GSCPerformance = perf
			enrichedIssue.EnrichedPriority = calculateEnrichedPriority(issue, perf)
			enrichedIssue.RecommendationReason = generateRecommendationReason(issue, perf)
		} else {
			// No GSC data available - use base priority
			enrichedIssue.EnrichedPriority = float64(getSeverityWeight(issue.Severity))
		}

		enriched = append(enriched, enrichedIssue)
	}

	return enriched
}

// calculateEnrichedPriority calculates priority with GSC data
func calculateEnrichedPriority(issue analyzer.Issue, perf *models.GSCPerformance) float64 {
	basePriority := float64(getSeverityWeight(issue.Severity))

	// Traffic multiplier
	trafficMultiplier := 1.0
	if perf.Impressions > 10000 {
		trafficMultiplier = 3.0 // High traffic = 3x priority
	} else if perf.Impressions > 1000 {
		trafficMultiplier = 2.0 // Medium traffic = 2x priority
	} else if perf.Impressions < 100 {
		trafficMultiplier = 0.5 // Low traffic = 0.5x priority
	}

	// CTR opportunity multiplier
	ctrMultiplier := 1.0
	if perf.CTR < 2.0 && perf.Impressions > 1000 {
		ctrMultiplier = 1.5 // Low CTR with high impressions = opportunity
	}

	// Position multiplier (pages ranking but not in top 10)
	positionMultiplier := 1.0
	if perf.Position > 10 && perf.Position < 20 && perf.Impressions > 500 {
		positionMultiplier = 1.3 // Could improve ranking
	}

	return basePriority * trafficMultiplier * ctrMultiplier * positionMultiplier
}

// generateRecommendationReason creates contextual recommendation based on GSC data
func generateRecommendationReason(issue analyzer.Issue, perf *models.GSCPerformance) string {
	if perf.Impressions > 10000 {
		return fmt.Sprintf("This page has high search visibility (%d impressions/month). Fixing this issue could significantly impact your SEO performance.", perf.Impressions)
	} else if perf.Impressions > 1000 {
		if perf.CTR < 2.0 {
			return fmt.Sprintf("This page has moderate visibility (%d impressions/month) but low CTR (%.1f%%). Optimizing this could improve click-through rates.", perf.Impressions, perf.CTR)
		}
		return fmt.Sprintf("This page has moderate search visibility (%d impressions/month).", perf.Impressions)
	} else if perf.Impressions < 100 {
		return "This page has minimal search visibility. Consider fixing as part of broader technical SEO improvements."
	}
	return ""
}

// getSeverityWeight returns weight for severity level
func getSeverityWeight(severity string) int {
	switch severity {
	case "error":
		return 10
	case "warning":
		return 5
	case "info":
		return 1
	default:
		return 1
	}
}


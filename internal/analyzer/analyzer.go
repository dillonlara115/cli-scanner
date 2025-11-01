package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dillonlara115/baracuda/pkg/models"
)

// IssueType represents the type of SEO issue detected
type IssueType string

const (
	IssueMissingH1          IssueType = "missing_h1"
	IssueMissingMetaDesc    IssueType = "missing_meta_description"
	IssueMissingTitle       IssueType = "missing_title"
	IssueLongTitle          IssueType = "long_title"
	IssueLongMetaDesc       IssueType = "long_meta_description"
	IssueShortTitle         IssueType = "short_title"
	IssueShortMetaDesc      IssueType = "short_meta_description"
	IssueLargeImage         IssueType = "large_image"
	IssueMissingImageAlt    IssueType = "missing_image_alt"
	IssueSlowResponse       IssueType = "slow_response"
	IssueRedirectChain      IssueType = "redirect_chain"
	IssueNoCanonical        IssueType = "no_canonical"
	IssueBrokenLink         IssueType = "broken_link"
	IssueMultipleH1         IssueType = "multiple_h1"
	IssueEmptyH1            IssueType = "empty_h1"
)

// Issue represents a detected SEO issue
type Issue struct {
	Type           IssueType `json:"type"`
	Severity       string    `json:"severity"` // "error", "warning", "info"
	URL            string    `json:"url"`
	Message        string    `json:"message"`
	Value          string    `json:"value,omitempty"`
	Recommendation string    `json:"recommendation,omitempty"`
}

// Summary contains analysis results and statistics
type Summary struct {
	TotalPages           int                `json:"total_pages"`
	TotalIssues          int                `json:"total_issues"`
	IssuesByType         map[IssueType]int   `json:"issues_by_type"`
	Issues               []Issue            `json:"issues"`
	AverageResponseTime  int64              `json:"average_response_time_ms"`
	PagesWithErrors      int                `json:"pages_with_errors"`
	PagesWithRedirects   int                `json:"pages_with_redirects"`
	TotalInternalLinks   int                `json:"total_internal_links"`
	TotalExternalLinks   int                `json:"total_external_links"`
	SlowestPages         []PagePerformance  `json:"slowest_pages,omitempty"`
}

// PagePerformance tracks page performance metrics
type PagePerformance struct {
	URL          string `json:"url"`
	ResponseTime int64  `json:"response_time_ms"`
}

// Analyze analyzes crawl results and detects SEO issues
func Analyze(results []*models.PageResult) *Summary {
	summary := &Summary{
		TotalPages:   len(results),
		IssuesByType: make(map[IssueType]int),
		Issues:       make([]Issue, 0),
		SlowestPages: make([]PagePerformance, 0),
	}

	var totalResponseTime int64
	var slowPages []PagePerformance

	// Analyze basic issues first
	for _, result := range results {
		// Track response times
		totalResponseTime += result.ResponseTime
		if result.ResponseTime > 2000 { // Slower than 2 seconds
			slowPages = append(slowPages, PagePerformance{
				URL:          result.URL,
				ResponseTime: result.ResponseTime,
			})
		}

		// Track errors
		if result.Error != "" || result.StatusCode >= 400 {
			summary.PagesWithErrors++
			if result.StatusCode >= 400 {
				summary.Issues = append(summary.Issues, Issue{
					Type:           IssueBrokenLink,
					Severity:       "error",
					URL:            result.URL,
					Message:        fmt.Sprintf("HTTP %d", result.StatusCode),
					Value:          fmt.Sprintf("%d", result.StatusCode),
					Recommendation: "Fix broken link or redirect",
				})
				summary.IssuesByType[IssueBrokenLink]++
			}
		}

		// Track redirects
		if len(result.RedirectChain) > 0 {
			summary.PagesWithRedirects++
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueRedirectChain,
				Severity:       "warning",
				URL:            result.URL,
				Message:        fmt.Sprintf("Redirect chain: %s", strings.Join(result.RedirectChain, " -> ")),
				Value:          strings.Join(result.RedirectChain, " -> "),
				Recommendation: "Consider using direct links instead of redirect chains",
			})
			summary.IssuesByType[IssueRedirectChain]++
		}

		// Check title issues
		if result.Title == "" {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueMissingTitle,
				Severity:       "error",
				URL:            result.URL,
				Message:        "Missing page title",
				Recommendation: "Add a unique, descriptive title tag",
			})
			summary.IssuesByType[IssueMissingTitle]++
		} else {
			titleLen := len(result.Title)
			if titleLen < 30 {
				summary.Issues = append(summary.Issues, Issue{
					Type:           IssueShortTitle,
					Severity:       "warning",
					URL:            result.URL,
					Message:        fmt.Sprintf("Title too short (%d characters)", titleLen),
					Value:          result.Title,
					Recommendation: "Aim for 30-60 characters for optimal SEO",
				})
				summary.IssuesByType[IssueShortTitle]++
			} else if titleLen > 60 {
				summary.Issues = append(summary.Issues, Issue{
					Type:           IssueLongTitle,
					Severity:       "warning",
					URL:            result.URL,
					Message:        fmt.Sprintf("Title too long (%d characters)", titleLen),
					Value:          result.Title,
					Recommendation: "Keep titles under 60 characters to avoid truncation",
				})
				summary.IssuesByType[IssueLongTitle]++
			}
		}

		// Check meta description issues
		if result.MetaDesc == "" {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueMissingMetaDesc,
				Severity:       "warning",
				URL:            result.URL,
				Message:        "Missing meta description",
				Recommendation: "Add a unique meta description (120-160 characters)",
			})
			summary.IssuesByType[IssueMissingMetaDesc]++
		} else {
			descLen := len(result.MetaDesc)
			if descLen < 120 {
				summary.Issues = append(summary.Issues, Issue{
					Type:           IssueShortMetaDesc,
					Severity:       "info",
					URL:            result.URL,
					Message:        fmt.Sprintf("Meta description too short (%d characters)", descLen),
					Value:          result.MetaDesc,
					Recommendation: "Aim for 120-160 characters for optimal display",
				})
				summary.IssuesByType[IssueShortMetaDesc]++
			} else if descLen > 160 {
				summary.Issues = append(summary.Issues, Issue{
					Type:           IssueLongMetaDesc,
					Severity:       "warning",
					URL:            result.URL,
					Message:        fmt.Sprintf("Meta description too long (%d characters)", descLen),
					Value:          result.MetaDesc,
					Recommendation: "Keep under 160 characters to avoid truncation",
				})
				summary.IssuesByType[IssueLongMetaDesc]++
			}
		}

		// Check H1 issues
		if len(result.H1) == 0 {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueMissingH1,
				Severity:       "error",
				URL:            result.URL,
				Message:        "Missing H1 tag",
				Recommendation: "Add exactly one H1 tag per page",
			})
			summary.IssuesByType[IssueMissingH1]++
		} else if len(result.H1) > 1 {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueMultipleH1,
				Severity:       "warning",
				URL:            result.URL,
				Message:        fmt.Sprintf("Multiple H1 tags found (%d)", len(result.H1)),
				Value:          strings.Join(result.H1, ", "),
				Recommendation: "Use only one H1 tag per page for better SEO",
			})
			summary.IssuesByType[IssueMultipleH1]++
		} else if len(result.H1) == 1 && strings.TrimSpace(result.H1[0]) == "" {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueEmptyH1,
				Severity:       "error",
				URL:            result.URL,
				Message:        "H1 tag is empty",
				Recommendation: "Add meaningful content to H1 tag",
			})
			summary.IssuesByType[IssueEmptyH1]++
		}

		// Check canonical
		if result.Canonical == "" {
			summary.Issues = append(summary.Issues, Issue{
				Type:           IssueNoCanonical,
				Severity:       "info",
				URL:            result.URL,
				Message:        "No canonical tag found",
				Recommendation: "Consider adding canonical tag to prevent duplicate content issues",
			})
			summary.IssuesByType[IssueNoCanonical]++
		}

		// Count links
		summary.TotalInternalLinks += len(result.InternalLinks)
		summary.TotalExternalLinks += len(result.ExternalLinks)
	}

	// Calculate average response time
	if len(results) > 0 {
		summary.AverageResponseTime = totalResponseTime / int64(len(results))
	}

	// Sort slow pages
	sort.Slice(slowPages, func(i, j int) bool {
		return slowPages[i].ResponseTime > slowPages[j].ResponseTime
	})
	if len(slowPages) > 10 {
		summary.SlowestPages = slowPages[:10]
	} else {
		summary.SlowestPages = slowPages
	}

	summary.TotalIssues = len(summary.Issues)

	return summary
}

// AnalyzeWithImages analyzes results including image size checking
func AnalyzeWithImages(results []*models.PageResult, imageTimeout time.Duration) *Summary {
	summary := Analyze(results)

	// Add image analysis
	imageIssues := AnalyzeImages(results, imageTimeout)
	summary.Issues = append(summary.Issues, imageIssues...)

	// Update counts
	for _, issue := range imageIssues {
		summary.IssuesByType[issue.Type]++
	}
	summary.TotalIssues = len(summary.Issues)

	return summary
}

// GetIssueCountBySeverity returns counts grouped by severity
func (s *Summary) GetIssueCountBySeverity() map[string]int {
	counts := make(map[string]int)
	for _, issue := range s.Issues {
		counts[issue.Severity]++
	}
	return counts
}

// GetTopIssues returns the most common issues
func (s *Summary) GetTopIssues(limit int) []IssueType {
	type count struct {
		issueType IssueType
		count     int
	}
	counts := make([]count, 0, len(s.IssuesByType))
	for issueType, cnt := range s.IssuesByType {
		counts = append(counts, count{issueType: issueType, count: cnt})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})
	result := make([]IssueType, 0, limit)
	for i := 0; i < limit && i < len(counts); i++ {
		result = append(result, counts[i].issueType)
	}
	return result
}


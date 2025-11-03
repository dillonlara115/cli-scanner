package models

import "time"

// GSCPerformance represents Google Search Console performance data for a URL
type GSCPerformance struct {
	URL          string   `json:"url"`
	Impressions int64    `json:"impressions"`
	Clicks       int64    `json:"clicks"`
	CTR          float64  `json:"ctr"`
	Position     float64  `json:"position"`
	TopQueries   []Query  `json:"top_queries,omitempty"`
	LastUpdated  time.Time `json:"last_updated"`
}

// Query represents a search query from Google Search Console
type Query struct {
	Query       string  `json:"query"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Position    float64 `json:"position"`
}

// GSCProperty represents a Google Search Console property
type GSCProperty struct {
	URL           string `json:"url"`
	Type          string `json:"type"` // "URL_PREFIX" or "DOMAIN"
	Verified      bool   `json:"verified"`
}

// GSCAuthState represents OAuth state for security
type GSCAuthState struct {
	State        string    `json:"state"`
	RedirectURL  string    `json:"redirect_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// Note: EnrichedIssue is defined in internal/gsc/client.go to avoid circular dependency
// It extends analyzer.Issue with GSC performance data


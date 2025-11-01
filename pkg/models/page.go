package models

import "time"

// PageResult represents the SEO data extracted from a crawled page
type PageResult struct {
	URL           string    `json:"url"`
	StatusCode    int       `json:"status_code"`
	ResponseTime  int64     `json:"response_time_ms"` // Duration in milliseconds
	Title         string    `json:"title"`
	MetaDesc      string    `json:"meta_description"`
	Canonical     string    `json:"canonical"`
	H1            []string  `json:"h1"`
	H2            []string  `json:"h2"`
	H3            []string  `json:"h3"`
	H4            []string  `json:"h4"`
	H5            []string  `json:"h5"`
	H6            []string  `json:"h6"`
	InternalLinks []string  `json:"internal_links"`
	ExternalLinks []string  `json:"external_links"`
	Images        []Image   `json:"images,omitempty"`
	RedirectChain []string  `json:"redirect_chain,omitempty"`
	Error         string    `json:"error,omitempty"`
	CrawledAt     time.Time `json:"crawled_at"`
}

// Image represents an image found on a page
type Image struct {
	URL string `json:"url"`
	Alt string `json:"alt,omitempty"`
}



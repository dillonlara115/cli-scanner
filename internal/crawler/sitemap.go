package crawler

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	"github.com/dillonlara115/baracuda/internal/utils"
)

// SitemapIndex represents a sitemap index file
type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// Sitemap represents a single sitemap entry
type Sitemap struct {
	Loc string `xml:"loc"`
}

// URLSet represents a sitemap URL set
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// URL represents a single URL in a sitemap
type URL struct {
	Loc string `xml:"loc"`
}

// SitemapParser parses sitemap.xml files
type SitemapParser struct {
	fetcher *Fetcher
}

// NewSitemapParser creates a new SitemapParser instance
func NewSitemapParser(fetcher *Fetcher) *SitemapParser {
	return &SitemapParser{
		fetcher: fetcher,
	}
}

// ParseSitemap fetches and parses a sitemap URL, returning all URLs found
func (s *SitemapParser) ParseSitemap(sitemapURL string) ([]string, error) {
	result := s.fetcher.Fetch(sitemapURL)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", result.Error)
	}

	if result.PageResult.StatusCode != 200 {
		return nil, fmt.Errorf("sitemap returned HTTP %d", result.PageResult.StatusCode)
	}

	// Try parsing as sitemap index first
	var index SitemapIndex
	err := xml.Unmarshal(result.Body, &index)
	if err == nil && len(index.Sitemaps) > 0 {
		// It's a sitemap index, recursively parse each sitemap
		urls := make([]string, 0)
		for _, sitemap := range index.Sitemaps {
			subURLs, err := s.ParseSitemap(strings.TrimSpace(sitemap.Loc))
			if err != nil {
				utils.Debug("Failed to parse sub-sitemap", utils.NewField("url", sitemap.Loc), utils.NewField("error", err.Error()))
				continue
			}
			urls = append(urls, subURLs...)
		}
		return urls, nil
	}

	// Try parsing as URL set
	var urlSet URLSet
	err = xml.Unmarshal(result.Body, &urlSet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap XML: %w", err)
	}

	// Extract URLs and normalize them
	urls := make([]string, 0, len(urlSet.URLs))
	for _, u := range urlSet.URLs {
		normalized, err := utils.NormalizeURL(strings.TrimSpace(u.Loc))
		if err != nil {
			utils.Debug("Invalid URL in sitemap", utils.NewField("url", u.Loc), utils.NewField("error", err.Error()))
			continue
		}
		urls = append(urls, normalized)
	}

	return urls, nil
}

// DiscoverSitemapURL attempts to discover sitemap.xml URL from a base URL
func (s *SitemapParser) DiscoverSitemapURL(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s/sitemap.xml", u.Scheme, u.Host)
}


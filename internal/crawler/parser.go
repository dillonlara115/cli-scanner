package crawler

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dillonlara115/baracuda/internal/utils"
	"github.com/dillonlara115/baracuda/pkg/models"
)

// Parser extracts SEO data from HTML content
type Parser struct {
	baseURL string
	domain  string
}

// NewParser creates a new Parser instance
func NewParser(baseURL string) (*Parser, error) {
	domain, err := utils.ExtractDomain(baseURL)
	if err != nil {
		return nil, err
	}

	return &Parser{
		baseURL: baseURL,
		domain:  domain,
	}, nil
}

// Parse extracts SEO data from HTML content
func (p *Parser) Parse(htmlContent []byte) (*models.PageResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		return nil, err
	}

	result := &models.PageResult{
		URL:           p.baseURL,
		H1:            make([]string, 0),
		H2:            make([]string, 0),
		H3:            make([]string, 0),
		H4:            make([]string, 0),
		H5:            make([]string, 0),
		H6:            make([]string, 0),
		InternalLinks: make([]string, 0),
		ExternalLinks: make([]string, 0),
		Images:        make([]models.Image, 0),
	}

	// Extract title
	result.Title = strings.TrimSpace(doc.Find("title").First().Text())

	// Extract meta description
	doc.Find("meta[name='description']").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("content"); exists {
			result.MetaDesc = strings.TrimSpace(content)
		}
	})

	// Extract canonical link
	doc.Find("link[rel='canonical']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			result.Canonical = strings.TrimSpace(href)
		}
	})

	// Extract headings
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H1 = append(result.H1, text)
		}
	})
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H2 = append(result.H2, text)
		}
	})
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H3 = append(result.H3, text)
		}
	})
	doc.Find("h4").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H4 = append(result.H4, text)
		}
	})
	doc.Find("h5").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H5 = append(result.H5, text)
		}
	})
	doc.Find("h6").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result.H6 = append(result.H6, text)
		}
	})

	// Extract links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Resolve relative URLs
		resolvedURL, err := utils.ResolveURL(p.baseURL, href)
		if err != nil {
			return
		}

		// Normalize URL
		normalizedURL, err := utils.NormalizeURL(resolvedURL)
		if err != nil {
			return
		}

		// Skip fragments, javascript:, mailto:, etc.
		u, err := url.Parse(normalizedURL)
		if err != nil {
			return
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return
		}

		// Categorize as internal or external
		if utils.IsSameDomain(normalizedURL, p.baseURL) {
			// Avoid duplicates
			for _, existing := range result.InternalLinks {
				if existing == normalizedURL {
					return
				}
			}
			result.InternalLinks = append(result.InternalLinks, normalizedURL)
		} else {
			// Avoid duplicates
			for _, existing := range result.ExternalLinks {
				if existing == normalizedURL {
					return
				}
			}
			result.ExternalLinks = append(result.ExternalLinks, normalizedURL)
		}
	})

	// Extract images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		// Resolve relative URLs
		resolvedURL, err := utils.ResolveURL(p.baseURL, src)
		if err != nil {
			return
		}

		// Normalize URL
		normalizedURL, err := utils.NormalizeURL(resolvedURL)
		if err != nil {
			return
		}

		// Skip data URIs and non-HTTP schemes
		u, err := url.Parse(normalizedURL)
		if err != nil {
			return
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return
		}

		// Get alt text
		alt := s.AttrOr("alt", "")

		// Avoid duplicates
		for _, existing := range result.Images {
			if existing.URL == normalizedURL {
				return
			}
		}

		result.Images = append(result.Images, models.Image{
			URL: normalizedURL,
			Alt: alt,
		})
	})

	return result, nil
}

// ExtractLinks extracts all links from HTML content and returns them as a slice
func (p *Parser) ExtractLinks(htmlContent []byte) ([]string, error) {
	result, err := p.Parse(htmlContent)
	if err != nil {
		return nil, err
	}

	links := make([]string, 0, len(result.InternalLinks)+len(result.ExternalLinks))
	links = append(links, result.InternalLinks...)
	links = append(links, result.ExternalLinks...)

	return links, nil
}


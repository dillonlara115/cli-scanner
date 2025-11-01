package utils

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrEmptyStartURL   = errors.New("start URL cannot be empty")
	ErrInvalidMaxDepth = errors.New("max depth must be non-negative")
	ErrInvalidMaxPages = errors.New("max pages must be at least 1")
	ErrInvalidWorkers  = errors.New("workers must be at least 1")
	ErrInvalidExportFormat = errors.New("export format must be 'csv' or 'json'")
)

// NormalizeURL normalizes a URL by removing fragments and trailing slashes
func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	// Remove fragment
	u.Fragment = ""
	// Remove trailing slash unless it's root
	normalized := u.String()
	if normalized != u.Scheme+"://"+u.Host+"/" && strings.HasSuffix(normalized, "/") {
		normalized = normalized[:len(normalized)-1]
	}

	return normalized, nil
}

// ExtractDomain extracts the domain from a URL
func ExtractDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}
	return u.Host, nil
}

// IsSameDomain checks if two URLs belong to the same domain
func IsSameDomain(url1, url2 string) bool {
	domain1, err1 := ExtractDomain(url1)
	domain2, err2 := ExtractDomain(url2)
	if err1 != nil || err2 != nil {
		return false
	}
	return domain1 == domain2
}

// ResolveURL resolves a relative URL against a base URL
func ResolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	resolved := base.ResolveReference(rel)
	normalized, err := NormalizeURL(resolved.String())
	if err != nil {
		return "", err
	}

	return normalized, nil
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}


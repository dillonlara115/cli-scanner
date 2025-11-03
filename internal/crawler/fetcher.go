package crawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dillonlara115/barracuda/internal/utils"
	"github.com/dillonlara115/barracuda/pkg/models"
)

// Fetcher handles HTTP requests and response processing
type Fetcher struct {
	client    *http.Client
	userAgent string
}

// FetchResult contains the fetched page data
type FetchResult struct {
	PageResult *models.PageResult
	Body       []byte
	Error      error
}

// NewFetcher creates a new Fetcher instance
func NewFetcher(timeout time.Duration, userAgent string) *Fetcher {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirects up to 10 times
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &Fetcher{
		client:    client,
		userAgent: userAgent,
	}
}

// Fetch retrieves a URL and returns the response (single attempt, no retry)
func (f *Fetcher) Fetch(url string) *FetchResult {
	result := &FetchResult{
		PageResult: &models.PageResult{
			URL:       url,
			CrawledAt: time.Now(),
		},
	}

	startTime := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.PageResult.Error = result.Error.Error()
		return result
	}

	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Track redirect chain using CheckRedirect callback
	// CheckRedirect is called when the HTTP client encounters a redirect response
	var redirectChain []string
	originalCheckRedirect := f.client.CheckRedirect
	
	// Temporarily override CheckRedirect to capture redirect URLs
	f.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// When CheckRedirect is called:
		// - 'via' contains all previous requests (via[0] = original request)
		// - 'req' is the NEW request about to be made to follow the redirect
		// - The redirect response came from the last request in 'via'
		// 
		// We want to capture the redirect destinations (the URLs we're redirecting TO)
		// Each time CheckRedirect is called, we're following a redirect, so we capture req.URL
		if len(via) > 0 {
			// This is a redirect - capture the destination URL
			redirectChain = append(redirectChain, req.URL.String())
		}
		
		// Follow redirects up to 10 times
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}

	resp, err := f.client.Do(req)
	responseTime := time.Since(startTime)

	// Restore original CheckRedirect
	f.client.CheckRedirect = originalCheckRedirect

	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		result.PageResult.Error = result.Error.Error()
		result.PageResult.ResponseTime = responseTime.Milliseconds()
		return result
	}
	defer resp.Body.Close()

	result.PageResult.StatusCode = resp.StatusCode
	result.PageResult.ResponseTime = responseTime.Milliseconds()

	// Only add redirect chain if we actually had redirects (status code indicates redirects were followed)
	// If the final status is 3xx, it means we hit a redirect that wasn't followed, or
	// if we have redirectChain entries, we followed redirects
	if len(redirectChain) > 0 {
		result.PageResult.RedirectChain = redirectChain
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		result.PageResult.Error = result.Error.Error()
		return result
	}

	result.Body = body

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Error = fmt.Errorf("HTTP %d", resp.StatusCode)
		result.PageResult.Error = result.Error.Error()
	}

	return result
}

// isRetryableError checks if an error is retryable
func isRetryableError(result *FetchResult) bool {
	if result.Error == nil {
		return false
	}

	// Retry on 5xx errors, timeouts, and connection errors
	statusCode := result.PageResult.StatusCode
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// Check for timeout or connection errors in error message
	errMsg := result.Error.Error()
	if containsAny(errMsg, []string{"timeout", "connection refused", "no such host", "network is unreachable"}) {
		return true
	}

	return false
}

// containsAny checks if a string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// FetchWithRetry retrieves a URL with retry logic for transient errors
func (f *Fetcher) FetchWithRetry(url string, maxRetries int) *FetchResult {
	var lastResult *FetchResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: wait 2^attempt seconds
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		result := f.Fetch(url)
		lastResult = result

		// If successful or not retryable, return immediately
		if result.Error == nil || !isRetryableError(result) {
			return result
		}

		// Log retry attempt
		utils.Debug("Retrying fetch",
			utils.NewField("url", url),
			utils.NewField("attempt", attempt+1),
			utils.NewField("max_retries", maxRetries),
		)
	}

	return lastResult
}

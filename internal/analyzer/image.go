package analyzer

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dillonlara115/baracuda/pkg/models"
)

const (
	// MaxImageSizeKB is the threshold for considering images as "large"
	MaxImageSizeKB = 100
)

// ImageSizeInfo contains image size information
type ImageSizeInfo struct {
	URL    string
	SizeKB int64
	Size   int64
	Error  error
}

// CheckImageSize fetches image size using HEAD request
func CheckImageSize(imageURL string, timeout time.Duration) ImageSizeInfo {
	info := ImageSizeInfo{
		URL: imageURL,
	}

	client := &http.Client{
		Timeout: timeout,
	}

	// Try HEAD first (more efficient)
	req, err := http.NewRequest("HEAD", imageURL, nil)
	if err != nil {
		info.Error = err
		return info
	}

	resp, err := client.Do(req)
	if err != nil {
		info.Error = err
		return info
	}
	defer resp.Body.Close()

	// Check Content-Length header
	if resp.StatusCode == 200 {
		contentLength := resp.ContentLength
		if contentLength > 0 {
			info.Size = contentLength
			info.SizeKB = contentLength / 1024
			return info
		}
	}

	// If HEAD doesn't provide size, try GET but limit body read
	if resp.StatusCode == 200 && resp.ContentLength <= 0 {
		getReq, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			info.Error = err
			return info
		}

		// Only read first 1MB to get size
		getResp, err := client.Do(getReq)
		if err != nil {
			info.Error = err
			return info
		}
		defer getResp.Body.Close()

		if getResp.StatusCode == 200 {
			// Try to get size from Content-Length
			if getResp.ContentLength > 0 {
				info.Size = getResp.ContentLength
				info.SizeKB = getResp.ContentLength / 1024
			} else {
				// Read limited bytes to estimate
				limitedReader := io.LimitReader(getResp.Body, 1024*1024) // 1MB max
				bytesRead, _ := io.Copy(io.Discard, limitedReader)
				if bytesRead >= 1024*1024 {
					// If we hit the limit, it's at least 1MB
					info.Size = 1024 * 1024
					info.SizeKB = 1024
				} else {
					info.Size = bytesRead
					info.SizeKB = bytesRead / 1024
				}
			}
		}
	}

	return info
}

// AnalyzeImages analyzes images from page results and detects issues
func AnalyzeImages(results []*models.PageResult, timeout time.Duration) []Issue {
	var issues []Issue
	imageSizeCache := make(map[string]ImageSizeInfo)

	for _, result := range results {
		if result.StatusCode != 200 || result.Error != "" {
			continue
		}

		for _, img := range result.Images {
			// Check for missing alt text
			if img.Alt == "" {
				issues = append(issues, Issue{
					Type:           IssueMissingImageAlt,
					Severity:       "warning",
					URL:            result.URL,
					Message:        fmt.Sprintf("Image missing alt text: %s", img.URL),
					Value:          img.URL,
					Recommendation: "Add descriptive alt text for accessibility and SEO",
				})
			}

			// Check image size (with caching)
			if sizeInfo, cached := imageSizeCache[img.URL]; cached {
				if sizeInfo.SizeKB > MaxImageSizeKB {
					issues = append(issues, Issue{
						Type:           IssueLargeImage,
						Severity:       "warning",
						URL:            result.URL,
						Message:        fmt.Sprintf("Large image detected: %s (%d KB)", img.URL, sizeInfo.SizeKB),
						Value:          fmt.Sprintf("%s (%d KB)", img.URL, sizeInfo.SizeKB),
						Recommendation: fmt.Sprintf("Optimize image to reduce size below %d KB", MaxImageSizeKB),
					})
				}
			} else {
				// Fetch image size
				sizeInfo := CheckImageSize(img.URL, timeout)
				imageSizeCache[img.URL] = sizeInfo

				if sizeInfo.Error == nil && sizeInfo.SizeKB > MaxImageSizeKB {
					issues = append(issues, Issue{
						Type:           IssueLargeImage,
						Severity:       "warning",
						URL:            result.URL,
						Message:        fmt.Sprintf("Large image detected: %s (%d KB)", img.URL, sizeInfo.SizeKB),
						Value:          fmt.Sprintf("%s (%d KB)", img.URL, sizeInfo.SizeKB),
						Recommendation: fmt.Sprintf("Optimize image to reduce size below %d KB", MaxImageSizeKB),
					})
				}
			}
		}
	}

	return issues
}

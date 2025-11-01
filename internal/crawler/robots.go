package crawler

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/dillonlara115/baracuda/internal/utils"
	"github.com/temoto/robotstxt"
)

// RobotsChecker handles robots.txt checking and caching
type RobotsChecker struct {
	fetcher       *Fetcher
	cache         map[string]*robotstxt.Group
	cacheMu       sync.RWMutex
	userAgent     string
	respectRobots bool
}

// NewRobotsChecker creates a new RobotsChecker instance
func NewRobotsChecker(fetcher *Fetcher, userAgent string, respectRobots bool) *RobotsChecker {
	return &RobotsChecker{
		fetcher:       fetcher,
		cache:         make(map[string]*robotstxt.Group),
		userAgent:     userAgent,
		respectRobots: respectRobots,
	}
}

// IsAllowed checks if a URL is allowed by robots.txt
func (r *RobotsChecker) IsAllowed(targetURL string) (bool, error) {
	if !r.respectRobots {
		return true, nil
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)
	domain := u.Host

	// Check cache
	r.cacheMu.RLock()
	cached, exists := r.cache[domain]
	r.cacheMu.RUnlock()

	if exists && cached != nil {
		return cached.Test(targetURL), nil
	}

	// Fetch robots.txt
	robotsData, err := r.fetchRobotsTxt(robotsURL)
	if err != nil {
		// If robots.txt can't be fetched, allow by default
		utils.Debug("Could not fetch robots.txt", utils.NewField("url", robotsURL), utils.NewField("error", err.Error()))
		
		// Cache a permissive group to avoid repeated fetches
		r.cacheMu.Lock()
		r.cache[domain] = nil // nil means allow all
		r.cacheMu.Unlock()
		
		return true, nil
	}

	// Parse robots.txt
	robotsGroup, err := robotstxt.FromBytes(robotsData)
	if err != nil {
		utils.Debug("Could not parse robots.txt", utils.NewField("url", robotsURL), utils.NewField("error", err.Error()))
		
		// Cache a permissive group
		r.cacheMu.Lock()
		r.cache[domain] = nil
		r.cacheMu.Unlock()
		
		return true, nil
	}

	// Get group for user agent
	group := robotsGroup.FindGroup(r.userAgent)
	
	// Cache the group
	r.cacheMu.Lock()
	r.cache[domain] = group
	r.cacheMu.Unlock()

	return group.Test(targetURL), nil
}

// fetchRobotsTxt fetches robots.txt content
func (r *RobotsChecker) fetchRobotsTxt(robotsURL string) ([]byte, error) {
	result := r.fetcher.Fetch(robotsURL)
	if result.Error != nil {
		return nil, result.Error
	}
	
	if result.PageResult.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", result.PageResult.StatusCode)
	}
	
	return result.Body, nil
}


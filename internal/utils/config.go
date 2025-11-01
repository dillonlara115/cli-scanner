package utils

import (
	"time"
)

// Config holds all crawl configuration settings
type Config struct {
	StartURL      string
	MaxDepth      int
	MaxPages      int
	DomainFilter  string        // "same" or "all"
	Workers       int
	Delay         time.Duration
	Timeout       time.Duration
	UserAgent     string
	RespectRobots bool
	ParseSitemap  bool
	ExportFormat  string // "csv" or "json"
	ExportPath    string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		MaxDepth:      3,
		MaxPages:      1000,
		DomainFilter:  "same",
		Workers:       10,
		Delay:         0,
		Timeout:       30 * time.Second,
		UserAgent:     "baracuda/1.0.0",
		RespectRobots: true,
		ParseSitemap:  false,
		ExportFormat:  "csv",
		ExportPath:    "",
	}
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.StartURL == "" {
		return ErrEmptyStartURL
	}
	if c.MaxDepth < 0 {
		return ErrInvalidMaxDepth
	}
	if c.MaxPages < 1 {
		return ErrInvalidMaxPages
	}
	if c.Workers < 1 {
		return ErrInvalidWorkers
	}
	if c.ExportFormat != "csv" && c.ExportFormat != "json" {
		return ErrInvalidExportFormat
	}
	return nil
}


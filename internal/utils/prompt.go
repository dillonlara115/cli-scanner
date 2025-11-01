package utils

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PromptString prompts the user for a string input
func PromptString(prompt string, defaultValue string, required bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		
		input = strings.TrimSpace(input)
		
		// Use default if empty and default provided
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}
		
		// Check if required
		if required && input == "" {
			fmt.Println("⚠️  This field is required. Please try again.")
			continue
		}
		
		return input, nil
	}
}

// PromptInt prompts the user for an integer input
func PromptInt(prompt string, defaultValue int, required bool) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		defaultStr := ""
		if defaultValue > 0 {
			defaultStr = strconv.Itoa(defaultValue)
		}
		
		if defaultStr != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultStr)
		} else {
			fmt.Printf("%s: ", prompt)
		}
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		
		input = strings.TrimSpace(input)
		
		// Use default if empty and default provided
		if input == "" && defaultValue > 0 {
			return defaultValue, nil
		}
		
		// Allow empty for unlimited
		if input == "" && !required {
			return 0, nil
		}
		
		// Parse integer
		val, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("⚠️  Please enter a valid number. Try again.")
			continue
		}
		
		return val, nil
	}
}

// PromptBool prompts the user for a yes/no input
func PromptBool(prompt string, defaultValue bool) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Printf("%s", prompt)
		if defaultValue {
			fmt.Printf(" [Y/n]: ")
		} else {
			fmt.Printf(" [y/N]: ")
		}
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		
		// Use default if empty
		if input == "" {
			return defaultValue, nil
		}
		
		if input == "y" || input == "yes" {
			return true, nil
		}
		
		if input == "n" || input == "no" {
			return false, nil
		}
		
		fmt.Println("⚠️  Please enter 'y' for yes or 'n' for no.")
	}
}

// PromptChoice prompts the user to select from choices
func PromptChoice(prompt string, choices []string, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("%s\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}
	
	for {
		if defaultValue != "" {
			fmt.Printf("Select [%s]: ", defaultValue)
		} else {
			fmt.Printf("Select: ")
		}
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		
		input = strings.TrimSpace(input)
		
		// Use default if empty
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}
		
		// Try to parse as number
		if num, err := strconv.Atoi(input); err == nil {
			if num > 0 && num <= len(choices) {
				return choices[num-1], nil
			}
		}
		
		// Try to match by string
		for _, choice := range choices {
			if strings.EqualFold(input, choice) {
				return choice, nil
			}
		}
		
		fmt.Println("⚠️  Invalid choice. Please try again.")
	}
}

// PromptInteractive prompts the user for all crawl configuration interactively
func PromptInteractive() (*Config, string, string, bool, error) {
	fmt.Println("\n🐊 Baracuda - Interactive Crawl Setup")
	fmt.Println("═══════════════════════════════════════\n")
	
	// Get URL
	urlInput, err := PromptString("What is the URL you want to scan?", "", true)
	if err != nil {
		return nil, "", "", false, err
	}
	
	// Validate URL
	_, err = url.Parse(urlInput)
	if err != nil {
		return nil, "", "", false, fmt.Errorf("invalid URL: %w", err)
	}
	
	// Extract domain for directory naming
	parsedURL, _ := url.Parse(urlInput)
	domain := parsedURL.Hostname()
	if domain == "" {
		domain = "unknown"
	}
	
	// Create crawl directory
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	crawlDir := fmt.Sprintf("crawls/%s_%s", domain, timestamp)
	
	if err := os.MkdirAll(crawlDir, 0755); err != nil {
		return nil, "", "", false, fmt.Errorf("failed to create crawl directory: %w", err)
	}
	
	fmt.Printf("\n📁 Results will be saved to: %s/\n\n", crawlDir)
	
	// Get max pages
	maxPagesInput, err := PromptInt("How many pages do you want to scan? (leave blank for unlimited)", 1000, false)
	if err != nil {
		return nil, "", "", false, err
	}
	maxPages := maxPagesInput
	if maxPages == 0 {
		maxPages = 999999 // Very large number for "unlimited"
	}
	
	// Default to unlimited depth (no prompt)
	maxDepth := 9999 // Very large number for "unlimited"
	
	// Default to 10 workers (no prompt)
	workers := 10
	
	// Get export format
	format, err := PromptChoice("Export format?", []string{"csv", "json"}, "csv")
	if err != nil {
		return nil, "", "", false, err
	}
	
	// Get export path
	exportFilename := fmt.Sprintf("results.%s", format)
	exportPath := filepath.Join(crawlDir, exportFilename)
	
	// Ask if they want to customize export path
	customPath, err := PromptBool("Use custom export filename?", false)
	if err != nil {
		return nil, "", "", false, err
	}
	
	if customPath {
		customFilename, err := PromptString("Export filename", exportFilename, false)
		if err != nil {
			return nil, "", "", false, err
		}
		if customFilename != "" {
			exportPath = filepath.Join(crawlDir, customFilename)
		}
	}
	
	// Advanced options
	fmt.Println("\n📋 Advanced Options:")
	
	respectRobots, err := PromptBool("Respect robots.txt?", true)
	if err != nil {
		return nil, "", "", false, err
	}
	
	parseSitemap, err := PromptBool("Parse sitemap.xml for seed URLs?", false)
	if err != nil {
		return nil, "", "", false, err
	}
	
	// Always export link graph (no prompt)
	graphExport := filepath.Join(crawlDir, "graph.json")
	
	// Always open browser after crawl (no prompt)
	openBrowser := true
	
	// Build config
	config := &Config{
		StartURL:      urlInput,
		MaxDepth:      maxDepth,
		MaxPages:      maxPages,
		Workers:       workers,
		Delay:         0,
		Timeout:       30 * time.Second,
		UserAgent:     "baracuda/1.0.0",
		RespectRobots: respectRobots,
		ParseSitemap:  parseSitemap,
		ExportFormat:  format,
		ExportPath:    exportPath,
		DomainFilter:  "same",
	}
	
	return config, graphExport, crawlDir, openBrowser, nil
}


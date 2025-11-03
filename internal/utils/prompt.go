package utils

import (
	"bufio"
	"fmt"
	"io"
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

// PromptSelect prompts the user to select from choices using arrow keys
// Falls back to PromptChoice if terminal doesn't support ANSI codes
func PromptSelect(prompt string, choices []string, defaultValue string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices provided")
	}

	// Find default index
	defaultIndex := 0
	for i, choice := range choices {
		if choice == defaultValue {
			defaultIndex = i
			break
		}
	}

	// Check if we're in a terminal that supports ANSI codes
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		// Fallback to simple numbered selection
		return PromptChoice(prompt, choices, defaultValue)
	}

	// Check if stdin is a terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil || (fileInfo.Mode()&os.ModeCharDevice) == 0 {
		// Not a terminal, use fallback
		return PromptChoice(prompt, choices, defaultValue)
	}

	selected := defaultIndex
	
	fmt.Printf("%s\n", prompt)
	
	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Render function
	render := func() {
		// Move cursor up to start of menu (lines = len(choices) + 1 for prompt)
		for i := 0; i <= len(choices); i++ {
			fmt.Print("\033[1A\033[K") // Move up and clear line
		}
		
		// Print menu
		for i, choice := range choices {
			if i == selected {
				fmt.Printf("  \033[32m▶\033[0m %s\n", choice) // Green arrow for selected
			} else {
				fmt.Printf("    %s\n", choice)
			}
		}
	}

	// Initial render
	render()

	// Set terminal to raw mode (simplified - works on Unix)
	// For cross-platform, we'll try to read escape sequences
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		// Handle escape sequences (arrow keys)
		if char == '\x1b' {
			// Peek at next bytes without consuming
			buf := make([]byte, 2)
			n, _ := reader.Read(buf)
			if n == 2 {
				if buf[0] == '[' {
					if buf[1] == 'A' {
						// Up arrow
						if selected > 0 {
							selected--
							render()
						}
						continue
					} else if buf[1] == 'B' {
						// Down arrow
						if selected < len(choices)-1 {
							selected++
							render()
						}
						continue
					}
				}
			}
		} else if char == '\r' || char == '\n' {
			// Enter key
			fmt.Print("\n")
			return choices[selected], nil
		} else if char == '\x03' {
			// Ctrl+C
			fmt.Print("\033[?25h") // Show cursor
			return "", fmt.Errorf("interrupted")
		}
	}

	return choices[selected], nil
}

// PromptInteractive prompts the user for all crawl configuration interactively
func PromptInteractive() (*Config, string, string, bool, error) {
	fmt.Println()
	fmt.Println("🐊 Barracuda - Interactive Crawl Setup")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()
	
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
	
	// Default to unlimited pages (no prompt)
	maxPages := 0 // Will be set to 999999 for "unlimited"
	
	// Default to unlimited depth (no prompt)
	maxDepth := 9999 // Very large number for "unlimited"
	
	// Default to 10 workers (no prompt)
	workers := 10
	
	// Get export format (using arrow key selection, default to JSON)
	format, err := PromptSelect("Export format?", []string{"json", "csv"}, "json")
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
	
	// Default to respecting robots.txt (no prompt)
	respectRobots := true
	
	// Default to parsing sitemap.xml (no prompt)
	parseSitemap := true
	
	// Always export link graph (no prompt)
	graphExport := filepath.Join(crawlDir, "graph.json")
	
	// Always open browser after crawl (no prompt)
	openBrowser := true
	
	// Set maxPages to unlimited if 0
	if maxPages == 0 {
		maxPages = 999999 // Very large number for "unlimited"
	}
	
	// Build config
	config := &Config{
		StartURL:      urlInput,
		MaxDepth:      maxDepth,
		MaxPages:      maxPages,
		Workers:       workers,
		Delay:         0,
		Timeout:       30 * time.Second,
		UserAgent:     "barracuda/1.0.0",
		RespectRobots: respectRobots,
		ParseSitemap:  parseSitemap,
		ExportFormat:  format,
		ExportPath:    exportPath,
		DomainFilter:  "same",
	}
	
	return config, graphExport, crawlDir, openBrowser, nil
}

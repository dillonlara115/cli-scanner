package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dillonlara115/baracuda/internal/analyzer"
	"github.com/dillonlara115/baracuda/internal/exporter"
	"github.com/dillonlara115/baracuda/pkg/models"
	"github.com/spf13/cobra"
)

var (
	servePort    int
	serveResults string
	serveGraph   string
	serveSummary string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web server to view crawl results",
	Long: `Start a local web server to view crawl results in a web interface.
The server will serve the Svelte frontend and provide API endpoints for results data.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to run the server on")
	serveCmd.Flags().StringVar(&serveResults, "results", "results.json", "Path to JSON results file")
	serveCmd.Flags().StringVar(&serveGraph, "graph", "", "Path to link graph JSON file")
	serveCmd.Flags().StringVar(&serveSummary, "summary", "", "Path to summary JSON file (optional, will be generated from results if not provided)")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Determine file type and load results
	var results []*models.PageResult
	var err error

	if strings.HasSuffix(strings.ToLower(serveResults), ".csv") {
		// Load from CSV
		results, err = exporter.ImportCSV(serveResults)
		if err != nil {
			return fmt.Errorf("failed to import CSV: %w", err)
		}
	} else {
		// Load from JSON
		resultsData, err := os.ReadFile(serveResults)
		if err != nil {
			return fmt.Errorf("failed to read results file: %w", err)
		}

		if err := json.Unmarshal(resultsData, &results); err != nil {
			return fmt.Errorf("failed to parse results JSON: %w", err)
		}
	}

	// Generate or load summary
	var summary *analyzer.Summary
	if serveSummary != "" {
		summaryData, err := os.ReadFile(serveSummary)
		if err == nil {
			var s analyzer.Summary
			if err := json.Unmarshal(summaryData, &s); err == nil {
				summary = &s
			}
		}
	}

	if summary == nil {
		// Generate summary from results
		summary = analyzer.AnalyzeWithImages(results, 30*1000*1000*1000) // 30s timeout
	}

	// Load graph if provided
	var graphData map[string][]string
	if serveGraph != "" {
		graphBytes, err := os.ReadFile(serveGraph)
		if err == nil {
			json.Unmarshal(graphBytes, &graphData)
		}
	}

	// Setup API routes first (must be before catch-all handler)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(results)
	})

	apiMux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(summary)
	})

	apiMux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if graphData == nil {
			json.NewEncoder(w).Encode(map[string][]string{})
		} else {
			json.NewEncoder(w).Encode(graphData)
		}
	})

	// Serve static files from web/dist (after build)
	webDir := "web/dist"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "⚠️  Frontend not built. Run 'make frontend-build' or 'cd web && npm install && npm run build' first.\n")
		fmt.Fprintf(os.Stderr, "📁 Serving API only. Frontend files not found at %s\n", webDir)
		// Only serve API endpoints
		http.Handle("/", apiMux)
	} else {
		// Serve static files with SPA routing support
		fileServer := http.FileServer(http.Dir(webDir))
		
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Handle API routes first
			if strings.HasPrefix(r.URL.Path, "/api/") {
				apiMux.ServeHTTP(w, r)
				return
			}
			
			// Check if file exists
			path := filepath.Join(webDir, r.URL.Path)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			
			// For SPA routing, serve index.html for all non-API routes
			indexPath := filepath.Join(webDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
			} else {
				fileServer.ServeHTTP(w, r)
			}
		})
	}

	fmt.Fprintf(os.Stdout, "🚀 Starting Baracuda web server on http://localhost:%d\n", servePort)
	fmt.Fprintf(os.Stdout, "📊 Serving %d pages from %s\n", len(results), serveResults)
	fmt.Fprintf(os.Stdout, "🌐 Open http://localhost:%d in your browser\n", servePort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", servePort), nil); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}


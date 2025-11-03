package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dillonlara115/barracuda/internal/analyzer"
	"github.com/dillonlara115/barracuda/internal/exporter"
	"github.com/dillonlara115/barracuda/internal/gsc"
	"github.com/dillonlara115/barracuda/pkg/models"
	"github.com/spf13/cobra"
)

// FrontendFiles is set by main package via SetFrontendFiles
var frontendFiles fs.FS

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

	// Initialize GSC OAuth (non-blocking - will fail gracefully if credentials not set)
	gscRedirectURL := fmt.Sprintf("http://localhost:%d/api/gsc/callback", servePort)
	if err := gsc.InitializeOAuth(gscRedirectURL); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  GSC integration disabled: %v\n", err)
		fmt.Fprintf(os.Stderr, "💡 Set GSC_CLIENT_ID, GSC_CLIENT_SECRET, or GSC_CREDENTIALS_JSON to enable\n")
	}

	// GSC OAuth endpoints
	apiMux.HandleFunc("/api/gsc/connect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		authURL, state, err := gsc.GenerateAuthURL()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate auth URL: %v", err), http.StatusInternalServerError)
			return
		}
		// Return auth URL and state
		json.NewEncoder(w).Encode(map[string]string{
			"auth_url": authURL,
			"state":    state,
		})
	})

	apiMux.HandleFunc("/api/gsc/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if !gsc.ValidateState(state) {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		token, err := gsc.ExchangeCode(code)
		if err != nil {
			// Return error page that closes popup
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `
				<!DOCTYPE html>
				<html>
				<head><title>GSC Connection Error</title></head>
				<body>
					<h1>Connection Failed</h1>
					<p>%v</p>
					<script>
						window.opener && window.opener.postMessage({type: 'gsc_error', error: '%v'}, '*');
						setTimeout(() => window.close(), 2000);
					</script>
				</body>
				</html>
			`, err, err)
			return
		}

		// Store token (use session ID or IP as userID for now)
		userID := r.RemoteAddr
		gsc.StoreToken(userID, token)

		// Return success page that closes popup and signals parent window
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>GSC Connected</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
						background: #f5f5f5;
					}
					.container {
						text-align: center;
						background: white;
						padding: 2rem;
						border-radius: 8px;
						box-shadow: 0 2px 4px rgba(0,0,0,0.1);
					}
					.success { color: #10b981; font-size: 3rem; }
					h1 { color: #1f2937; }
					p { color: #6b7280; }
				</style>
			</head>
			<body>
				<div class="container">
					<div class="success">✓</div>
					<h1>Successfully Connected!</h1>
					<p>This window will close automatically...</p>
				</div>
				<script>
					// Signal parent window that connection succeeded
					if (window.opener) {
						window.opener.postMessage({type: 'gsc_connected', user_id: '%s'}, '*');
					}
					// Close popup after short delay
					setTimeout(() => {
						window.close();
					}, 1500);
				</script>
			</body>
			</html>
		`, userID)
	})

	apiMux.HandleFunc("/api/gsc/properties", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		// Get userID from query or use default
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			userID = r.RemoteAddr
		}

		properties, err := gsc.GetProperties(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Failed to get properties: %v", err),
			})
			return
		}
		json.NewEncoder(w).Encode(properties)
	})

	apiMux.HandleFunc("/api/gsc/performance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		var req struct {
			UserID   string `json:"user_id"`
			SiteURL  string `json:"site_url"`
			Days     int    `json:"days"` // Number of days to fetch (default 30)
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Invalid request: %v", err)})
			return
		}

		if req.UserID == "" {
			req.UserID = r.RemoteAddr
		}
		if req.Days == 0 {
			req.Days = 30
		}

		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -req.Days)

		performanceMap, err := gsc.FetchPerformanceData(req.UserID, req.SiteURL, startDate, endDate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Failed to fetch performance data: %v", err),
			})
			return
		}

		json.NewEncoder(w).Encode(performanceMap)
	})

	apiMux.HandleFunc("/api/gsc/enrich-issues", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		var req struct {
			UserID        string `json:"user_id"`
			SiteURL       string `json:"site_url"`
			Days          int    `json:"days"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Invalid request: %v", err)})
			return
		}

		if req.UserID == "" {
			req.UserID = r.RemoteAddr
		}
		if req.Days == 0 {
			req.Days = 30
		}

		// Fetch performance data
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -req.Days)
		performanceMap, err := gsc.FetchPerformanceData(req.UserID, req.SiteURL, startDate, endDate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Failed to fetch performance data: %v", err),
			})
			return
		}

		// Enrich issues
		enrichedIssues := gsc.EnrichIssues(summary.Issues, performanceMap)
		json.NewEncoder(w).Encode(enrichedIssues)
	})

	// Set cache headers based on file type
	// Assets with hashes can be cached long-term, HTML should not be cached
	setCacheHeaders := func(w http.ResponseWriter, path string) {
		// Assets with hashes (like /assets/index-abc123.js) can be cached long-term
		if strings.HasPrefix(path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasSuffix(path, ".html") {
			// HTML files should not be cached
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else {
			// Other static files (images, fonts, etc.) can be cached moderately
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
	}

	// Serve static files from embedded frontend or fallback to filesystem
	// Try embedded files first (production build)
	var fileServer http.Handler
	var useEmbedded bool
	
	// Check if embedded files exist (they should if frontend was built before Go build)
	if frontendFiles != nil {
		// Try to read from embedded files
		if entries, err := fs.ReadDir(frontendFiles, "web/dist"); err == nil && len(entries) > 0 {
			// Use embedded files
			fsys, err := fs.Sub(frontendFiles, "web/dist")
			if err == nil {
				fileServer = http.FileServer(http.FS(fsys))
				useEmbedded = true
			}
		}
	}
	
	// Fallback to filesystem (for development)
	if !useEmbedded {
		webDir := "web/dist"
		if _, err := os.Stat(webDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "⚠️  Frontend not built. Run 'make frontend-build' or 'cd web && npm install && npm run build' first.\n")
			fmt.Fprintf(os.Stderr, "📁 Serving API only. Frontend files not found at %s\n", webDir)
			// Only serve API endpoints
			http.Handle("/", apiMux)
		} else {
			// Use filesystem
			fileServer = http.FileServer(http.Dir(webDir))
			useEmbedded = false
		}
	}
	
	if fileServer != nil {
		// Serve static files with SPA routing support
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Handle API routes first
			if strings.HasPrefix(r.URL.Path, "/api/") {
				apiMux.ServeHTTP(w, r)
				return
			}
			
			if useEmbedded {
				// For embedded files, check if file exists
				fsys, _ := fs.Sub(frontendFiles, "web/dist")
				path := strings.TrimPrefix(r.URL.Path, "/")
				if path == "" {
					path = "index.html"
				}
				
				// Try to open the file
				if f, err := fsys.Open(path); err == nil {
					defer f.Close()
					// Get file info to check if it's a directory
					if info, err := f.Stat(); err == nil && !info.IsDir() {
						// Set cache headers based on file type
						setCacheHeaders(w, r.URL.Path)
						fileServer.ServeHTTP(w, r)
						return
					}
				}
				
				// For SPA routing, serve index.html for all non-API routes
				if _, err := fsys.Open("index.html"); err == nil {
					r.URL.Path = "/index.html"
					// Don't cache index.html - always get fresh version
					w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
					fileServer.ServeHTTP(w, r)
				} else {
					http.NotFound(w, r)
				}
			} else {
				// Filesystem fallback
				path := filepath.Join("web/dist", r.URL.Path)
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					// Set cache headers based on file type
					setCacheHeaders(w, r.URL.Path)
					fileServer.ServeHTTP(w, r)
					return
				}
				
				// For SPA routing, serve index.html for all non-API routes
				indexPath := filepath.Join("web/dist", "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					r.URL.Path = "/index.html"
					// Don't cache index.html - always get fresh version
					w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
					fileServer.ServeHTTP(w, r)
				} else {
					fileServer.ServeHTTP(w, r)
				}
			}
		})
	}

	fmt.Fprintf(os.Stdout, "🚀 Starting Barracuda web server on http://localhost:%d\n", servePort)
	fmt.Fprintf(os.Stdout, "📊 Serving %d pages from %s\n", len(results), serveResults)
	fmt.Fprintf(os.Stdout, "🌐 Open http://localhost:%d in your browser\n", servePort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", servePort), nil); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// SetFrontendFiles sets the embedded frontend filesystem
func SetFrontendFiles(fs fs.FS) {
	frontendFiles = fs
}


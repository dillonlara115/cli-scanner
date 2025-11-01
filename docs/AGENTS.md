# AGENTS.md - AI Agent Context Guide

This document provides essential context for AI agents working on the Baracuda project. It outlines the architecture, patterns, conventions, and workflows to help agents understand and modify the codebase effectively.

---

## Project Overview

**Baracuda** is a fast, lightweight SEO website crawler CLI tool inspired by Screaming Frog. It crawls websites recursively, extracts SEO data, detects issues, and provides a web dashboard for visualization.

**Key Features:**
- Recursive website crawling with configurable depth/limits
- SEO data extraction (titles, meta descriptions, headings, links, images)
- Automatic SEO issue detection and analysis
- Export to CSV/JSON formats
- Link graph visualization
- Interactive CLI prompts
- Svelte-based web dashboard

---

## Tech Stack

### Backend (Go)
- **Language:** Go 1.21+
- **CLI Framework:** Cobra (`github.com/spf13/cobra`)
- **HTML Parsing:** goquery (`github.com/PuerkitoBio/goquery`)
- **Robots.txt:** robotstxt (`github.com/temoto/robotstxt`)
- **Logging:** zap (`go.uber.org/zap`)
- **HTTP Client:** Standard library `net/http`

### Frontend (Svelte)
- **Framework:** Svelte + Vite
- **CSS Framework:** Tailwind CSS + DaisyUI (business theme)
- **Build Tool:** Vite
- **Package Manager:** npm

---

## Project Structure

```
baracuda/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command, banner display
│   ├── crawl.go           # Crawl command (main functionality)
│   ├── serve.go           # Serve command (web dashboard server)
│   ├── browser.go         # Browser opening utilities
│   └── banner.go          # ASCII art banner
├── internal/
│   ├── analyzer/          # SEO analysis engine
│   │   ├── analyzer.go    # Main analyzer logic
│   │   ├── image.go       # Image size analysis
│   │   └── printer.go     # Summary printing
│   ├── crawler/           # Crawling engine
│   │   ├── manager.go     # Orchestrates crawling (workers, queue)
│   │   ├── fetcher.go     # HTTP fetching with retry logic
│   │   ├── parser.go      # HTML parsing (goquery)
│   │   ├── robots.go      # Robots.txt checking
│   │   └── sitemap.go     # Sitemap.xml parsing
│   ├── exporter/          # Export formats
│   │   ├── csv.go         # CSV export
│   │   ├── json.go        # JSON export
│   │   └── csv_import.go  # CSV import (for serve command)
│   ├── graph/             # Link graph
│   │   └── graph.go       # Graph data structure
│   └── utils/             # Utilities
│       ├── config.go       # Configuration struct
│       ├── logger.go       # Logging setup
│       ├── prompt.go       # Interactive prompts
│       └── url.go          # URL utilities (normalize, resolve)
├── pkg/
│   └── models/
│       └── page.go         # PageResult and Image models
├── web/                    # Svelte frontend
│   ├── src/
│   │   ├── components/    # Svelte components
│   │   │   ├── Dashboard.svelte
│   │   │   ├── SummaryCard.svelte
│   │   │   ├── ResultsTable.svelte
│   │   │   ├── IssuesPanel.svelte
│   │   │   └── LinkGraph.svelte
│   │   ├── App.svelte      # Main app component
│   │   └── main.js         # Entry point
│   └── dist/               # Built frontend (served by Go)
├── docs/                   # Documentation
│   └── DASHBOARD_IMPROVEMENTS.md
├── main.go                 # Entry point
├── go.mod                  # Go dependencies
├── Makefile                # Build commands
└── README.md               # User documentation
```

---

## Architecture Patterns

### CLI Structure (Cobra)
- **Root command:** `baracuda` - Shows banner, defaults to interactive crawl
- **Subcommands:** 
  - `crawl [URL]` - Main crawling functionality
  - `serve` - Web dashboard server

**Pattern:**
```go
// Commands are defined in cmd/*.go files
var crawlCmd = &cobra.Command{
    Use:   "crawl [URL]",
    Short: "Description",
    RunE:  runCrawl,  // Handler function
}

func init() {
    rootCmd.AddCommand(crawlCmd)
    // Define flags here
}
```

### Crawling Architecture

**Flow:**
1. User runs `baracuda crawl <URL>` or `baracuda` (interactive)
2. `cmd/crawl.go` → `runCrawl()` validates config
3. Creates `crawler.Manager` with config
4. Manager spawns worker pool (goroutines)
5. Workers fetch URLs from queue, parse HTML, discover links
6. Results stored in memory, exported to CSV/JSON
7. Analyzer processes results, detects SEO issues
8. Optional: Open browser with dashboard

**Key Components:**
- **Manager** (`crawler/manager.go`): Orchestrates workers, manages queue, visited URLs
- **Fetcher** (`crawler/fetcher.go`): HTTP requests with retry, timeout, redirect handling
- **Parser** (`crawler/parser.go`): Extracts SEO data from HTML using goquery
- **RobotsChecker** (`crawler/robots.go`): Caches and checks robots.txt rules
- **SitemapParser** (`crawler/sitemap.go`): Parses sitemap.xml for seed URLs

**Concurrency:**
- Uses channels for task queue
- Atomic operations (`sync/atomic`) for counters
- `sync.Map` for visited URLs (concurrent-safe)
- Context cancellation for graceful shutdown

### Data Models

**PageResult** (`pkg/models/page.go`):
```go
type PageResult struct {
    URL           string
    StatusCode    int
    ResponseTime  int64  // milliseconds
    Title         string
    MetaDesc      string
    Canonical     string
    H1-H6         []string
    InternalLinks []string
    ExternalLinks []string
    Images        []Image
    RedirectChain []string
    Error         string
    CrawledAt     time.Time
}
```

**Summary** (`internal/analyzer/analyzer.go`):
- Contains aggregated statistics and detected issues
- Issues include: type, severity, URL, message, recommendation

### Export Flow

1. Results collected in `[]*models.PageResult`
2. Exporter chosen based on `--format` flag (csv/json)
3. CSV: Flattened with pipe-separated arrays
4. JSON: Full nested structure
5. Files written to disk or stdout

### Web Dashboard Flow

1. `baracuda serve --results results.json --graph graph.json`
2. `cmd/serve.go` loads JSON/CSV files
3. Generates summary via `analyzer.AnalyzeWithImages()`
4. Serves static files from `web/dist/`
5. API endpoints: `/api/results`, `/api/summary`, `/api/graph`
6. SPA routing: All non-API routes serve `index.html`

---

## Key Conventions

### Go Code Style
- **Package naming:** Lowercase, no underscores
- **Exported functions:** PascalCase
- **File organization:** One main type per file when possible
- **Error handling:** Always return errors, wrap with context
- **Logging:** Use `utils.Info()`, `utils.Debug()`, `utils.Error()`

### Import Organization
```go
import (
    // Standard library
    "fmt"
    "net/http"
    
    // Third-party
    "github.com/spf13/cobra"
    
    // Internal
    "github.com/dillonlara115/baracuda/internal/utils"
    "github.com/dillonlara115/baracuda/pkg/models"
)
```

### Frontend Style
- **Components:** PascalCase, one component per file
- **Props:** Use `export let propName`
- **Reactive statements:** Use `$:` for computed values
- **Styling:** Tailwind CSS classes, DaisyUI components

### Error Handling Pattern
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### Logging Pattern
```go
utils.Info("Message", utils.NewField("key", value))
utils.Debug("Debug message", utils.NewField("url", url))
utils.Error("Error message", utils.NewField("error", err.Error()))
```

---

## Common Tasks & Patterns

### Adding a New CLI Flag

**Location:** `cmd/crawl.go`

```go
var newFlag string

func init() {
    crawlCmd.Flags().StringVar(&newFlag, "new-flag", "default", "Description")
}

func runCrawl(cmd *cobra.Command, args []string) error {
    // Use newFlag here
}
```

### Adding a New SEO Issue Type

**Location:** `internal/analyzer/analyzer.go`

1. Add constant:
```go
const IssueNewType IssueType = "new_type"
```

2. Add detection logic in `Analyze()`:
```go
if condition {
    summary.Issues = append(summary.Issues, Issue{
        Type: IssueNewType,
        Severity: "warning",
        URL: result.URL,
        Message: "Description",
        Recommendation: "Fix suggestion",
    })
    summary.IssuesByType[IssueNewType]++
}
```

3. Update `formatIssueType()` in `printer.go`:
```go
case IssueNewType:
    return "Human Readable Name"
```

### Adding a New Frontend Component

**Location:** `web/src/components/`

1. Create `.svelte` file
2. Export props: `export let data`
3. Use DaisyUI components and Tailwind classes
4. Import in `Dashboard.svelte`:
```svelte
<script>
  import NewComponent from './NewComponent.svelte';
</script>
```

### Adding a New Export Format

**Location:** `internal/exporter/`

1. Create `newformat.go`:
```go
package exporter

func ExportNewFormat(results []*models.PageResult, path string) error {
    // Implementation
}
```

2. Update `cmd/crawl.go` to support new format:
```go
case "newformat":
    return exporter.ExportNewFormat(results, config.ExportPath)
```

### Adding a New API Endpoint

**Location:** `cmd/serve.go`

```go
apiMux.HandleFunc("/api/new-endpoint", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    json.NewEncoder(w).Encode(data)
})
```

---

## Important Notes & Gotchas

### 1. URL Normalization
- Always use `utils.NormalizeURL()` before storing/comparing URLs
- Handles trailing slashes, schemes, fragments
- Prevents duplicate crawling

### 2. Concurrent Access
- Use `sync.Map` for visited URLs (concurrent-safe)
- Use `sync.Mutex` for results slice (`resultsMu`)
- Use atomic operations for counters (`pending`)

### 3. Queue Management
- Queue is closed when `monitorQueue` detects no pending tasks
- Check `queueClosed` atomic flag before sending
- Workers handle closed queue gracefully

### 4. Image Size Analysis
- Uses HEAD requests (not GET) to check size
- Caches results to avoid duplicate requests
- Timeout controlled by config

### 5. Interactive Mode
- Triggered when: `--interactive` flag OR no URL + no flags
- Creates `crawls/{domain}_{timestamp}/` directory
- Always exports graph and opens browser

### 6. Banner Display
- Shown when running main commands
- NOT shown for `--help`, `--version`, `help`
- Displayed in `cmd/root.go` Run function

### 7. Frontend Build
- Frontend must be built (`make frontend-build`) before serving
- Built files go to `web/dist/`
- SPA routing: All routes serve `index.html` except `/api/*`

### 8. Export Paths
- Default: `results.csv` or `results.json` in current directory
- Interactive mode: `crawls/{domain}_{timestamp}/results.{format}`
- Graph export: `graph.json` (same directory as results)

### 9. Issue Detection
- Analyzer runs AFTER crawl completes
- Image analysis requires additional HTTP requests
- Summary printed to terminal, also available via API

### 10. Worker Pool
- Default: 10 workers
- Configurable via `--workers` flag
- Workers process queue concurrently
- Max pages limit checked before and after processing

---

## Testing Guidelines

### Running Tests
```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
make bench            # Run benchmarks
```

### Test Structure
- Tests should be in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Test concurrent behavior carefully

### Frontend Testing
- Currently no test setup, but consider:
  - Vitest for unit tests
  - Playwright for E2E tests

---

## Dependencies

### Go Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/PuerkitoBio/goquery` - HTML parsing
- `github.com/temoto/robotstxt` - Robots.txt parsing
- `go.uber.org/zap` - Structured logging

### Frontend Dependencies
- See `web/package.json` for full list
- Key: Svelte, Vite, Tailwind CSS, DaisyUI

---

## Development Workflow

### Local Development
```bash
# Build CLI
make build

# Run tests
make test

# Build frontend
make frontend-build

# Dev frontend (hot reload)
make frontend-dev

# Serve results
make serve
```

### Adding Features

1. **Backend Feature:**
   - Create/modify files in `internal/` or `cmd/`
   - Add tests if applicable
   - Update `cmd/crawl.go` if adding flags
   - Run `make test`

2. **Frontend Feature:**
   - Create/modify components in `web/src/components/`
   - Use DaisyUI components and Tailwind classes
   - Test in dev mode: `make frontend-dev`
   - Build: `make frontend-build`

3. **Export Format:**
   - Add file in `internal/exporter/`
   - Update `cmd/crawl.go` to handle new format
   - Update `cmd/serve.go` if needed for import

---

## File-Specific Notes

### `cmd/crawl.go`
- Main crawl command handler
- Handles interactive mode detection
- Coordinates crawler, analyzer, exporter
- Opens browser if `--open` flag is set

### `cmd/serve.go`
- Web server for dashboard
- Loads JSON/CSV results
- Generates summary if not provided
- Serves static files from `web/dist/`
- Handles SPA routing

### `internal/crawler/manager.go`
- Core crawling orchestration
- Manages worker pool and task queue
- Handles visited URLs and limits
- Thread-safe result collection

### `internal/analyzer/analyzer.go`
- SEO issue detection logic
- Processes PageResults
- Generates Summary with statistics
- Issue types defined as constants

### `pkg/models/page.go`
- Core data models
- PageResult: Complete page data
- Image: Image with URL and alt text
- Used throughout codebase

---

## Common Issues & Solutions

### Issue: Duplicate URLs being crawled
**Solution:** Ensure `utils.NormalizeURL()` is called before checking/storing URLs

### Issue: Queue closing prematurely
**Solution:** Check `queueClosed` atomic flag, wait for pending tasks to complete

### Issue: Frontend not loading
**Solution:** Ensure `make frontend-build` was run, check `web/dist/` exists

### Issue: Banner not showing
**Solution:** Check `cmd/root.go` Run function, ensure not triggered by help/version flags

### Issue: Image analysis slow
**Solution:** Uses HEAD requests and caching, but still requires HTTP requests per image

---

## Future Considerations

### Planned Features (see `docs/DASHBOARD_IMPROVEMENTS.md`)
- Export issues functionality
- Dashboard navigation improvements
- Issue prioritization
- Page-level issue views
- Progress tracking

### Technical Debt
- No database storage (all in-memory)
- No JavaScript rendering (static HTML only)
- Frontend testing not set up
- No persistent issue status tracking

### Performance Optimizations
- Consider caching robots.txt more aggressively
- Batch image size checks
- Implement request rate limiting per domain

---

## Quick Reference Commands

```bash
# Build
make build

# Test
make test

# Frontend
make frontend-build    # Build
make frontend-dev     # Dev server

# Install
make install          # Go install
make install-alias    # Add shell alias

# Clean
make clean

# Release
make release          # Cross-platform builds
```

---

## Contact & Resources

- **Project:** github.com/dillonlara115/baracuda
- **Documentation:** See `README.md` and `docs/` directory
- **Issues:** Check GitHub issues

---

**Last Updated:** 2025-01-01  
**Maintainer:** Maintain this file as the codebase evolves


# baracuda

A fast, lightweight SEO website crawler CLI tool inspired by Screaming Frog.

## Features

- **Recursive Crawling**: Crawl websites with configurable depth and page limits
- **SEO Data Extraction**: Extracts titles, meta descriptions, H1-H6 headings, canonical tags, and links
- **HTTP Status Tracking**: Tracks status codes, redirects, and response times
- **Concurrent Processing**: Async crawling with configurable worker pool
- **Robots.txt Support**: Respects robots.txt rules (optional)
- **Sitemap Parsing**: Optionally seed crawl queue from sitemap.xml
- **Export Formats**: CSV and JSON export options
- **Link Graph**: Build and export link graphs for visualization
- **Web Dashboard**: Beautiful Svelte-based web interface for viewing results
- **SEO Analysis**: Automatic issue detection with recommendations

## Installation

### From Source

```bash
git clone https://github.com/dillonlara115/baracuda.git
cd baracuda
go build -o baracuda .
sudo mv baracuda /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/dillonlara115/baracuda@latest
```

### Frontend Setup (Optional)

To use the web dashboard, you'll need to build the frontend:

```bash
cd web
npm install
npm run build
```

Or use the Makefile:

```bash
make frontend-build
```

## Usage

### Basic Usage

```bash
# Crawl a website
baracuda crawl https://example.com

# Crawl with custom depth and export format
baracuda crawl https://example.com --max-depth 2 --format json

# Export to specific file
baracuda crawl https://example.com --export results.csv
```

### Advanced Options

```bash
# Full example with all options
baracuda crawl https://example.com \
  --max-depth 3 \
  --max-pages 500 \
  --workers 20 \
  --delay 100ms \
  --timeout 60s \
  --format json \
  --export crawl-results.json \
  --graph-export link-graph.json \
  --parse-sitemap \
  --respect-robots
```

### Web Dashboard

After crawling, view your results in a beautiful web interface:

```bash
# First, crawl with JSON export
baracuda crawl https://example.com --format json --export results.json --graph-export graph.json

# Then serve the results
baracuda serve --results results.json --graph graph.json
```

Or use the Makefile shortcut:

```bash
make serve
```

The web dashboard includes:
- **Dashboard**: Overview statistics and metrics
- **Results Table**: Filterable, sortable table of all crawled pages
- **Issues Panel**: Detailed SEO issues with recommendations
- **Link Graph**: Visualization of internal/external link structure

Access the dashboard at `http://localhost:8080` (default port).

## Command-Line Flags

### Required Flags

- `--url, -u`: Starting URL to crawl (required)

### Crawl Options

- `--max-depth, -d`: Maximum crawl depth (default: 3)
- `--max-pages, -p`: Maximum number of pages to crawl (default: 1000)
- `--workers, -w`: Number of concurrent workers (default: 10)
- `--delay`: Delay between requests (e.g., 100ms) (default: 0ms)
- `--timeout`: HTTP request timeout (default: 30s)
- `--user-agent`: User agent string (default: baracuda/1.0.0)
- `--respect-robots`: Respect robots.txt rules (default: true)
- `--parse-sitemap`: Parse sitemap.xml for seed URLs (default: false)
- `--domain-filter`: Domain filter: 'same' or 'all' (default: same)

### Export Options

- `--format, -f`: Export format: 'csv' or 'json' (default: csv)
- `--export, -e`: Export file path (default: results.csv/json)
- `--graph-export`: Export link graph to JSON file (optional)

### Serve Command (Web Dashboard)

- `serve`: Start web server to view crawl results
  - `--port`: Port to run the server on (default: 8080)
  - `--results`: Path to JSON results file (default: results.json)
  - `--graph`: Path to link graph JSON file (optional)
  - `--summary`: Path to summary JSON file (optional, auto-generated if not provided)

### Global Flags

- `--debug`: Enable debug logging
- `--version`: Show version information

## Examples

### Example 1: Basic Crawl

```bash
baracuda crawl https://example.com
```

This will:
- Crawl up to 3 levels deep
- Export results to `results.csv`
- Respect robots.txt by default

### Example 2: JSON Export with Link Graph

```bash
baracuda crawl https://example.com \
  --format json \
  --export results.json \
  --graph-export graph.json
```

### Example 3: Fast Crawl (No Robots, Higher Concurrency)

```bash
baracuda crawl https://example.com \
  --workers 50 \
  --max-pages 5000 \
  --respect-robots=false
```

### Example 4: Sitemap-Based Crawl

```bash
baracuda crawl https://example.com \
  --parse-sitemap \
  --max-depth 1
```

### Example 5: View Results in Web Dashboard

```bash
# Step 1: Crawl and export to JSON
baracuda crawl https://example.com \
  --format json \
  --export results.json \
  --graph-export graph.json

# Step 2: Build frontend (first time only)
cd web && npm install && npm run build

# Step 3: Serve results
baracuda serve --results results.json --graph graph.json

# Open http://localhost:8080 in your browser
```

## Output Format

### CSV Export

The CSV export includes the following columns:
- URL
- Status Code
- Response Time (ms)
- Title
- Meta Description
- Canonical
- H1-H6 (pipe-separated values)
- Internal Links (pipe-separated)
- External Links (pipe-separated)
- Redirect Chain (arrow-separated)
- Error
- Crawled At

### JSON Export

The JSON export includes an array of page results with all SEO data fields.

### Link Graph Export

The link graph is exported as a JSON object mapping source URLs to arrays of target URLs:
```json
{
  "https://example.com/page1": [
    "https://example.com/page2",
    "https://example.com/page3"
  ],
  "https://example.com/page2": [
    "https://example.com/page4"
  ]
}
```

## Performance

- Typical crawl speed: 100-500 pages/minute (depends on server response times)
- Memory usage: ~50-100 MB for 1000 pages (varies by page size)
- Concurrent workers: Adjust `--workers` based on your system and target server capacity

## SEO Analysis

The crawler automatically detects SEO issues including:

- Missing or duplicate H1 tags
- Missing meta descriptions
- Missing or poor titles
- Large images (>100KB)
- Missing image alt text
- Slow response times
- Redirect chains
- Broken links

Issues are displayed in the terminal summary and can be viewed in detail in the web dashboard.

## Limitations

- No database storage (all data in-memory)
- No keyword/rank tracking (planned for future versions)
- No JavaScript rendering (static HTML only)
- Binary size: ~15-20 MB

## Development

### Building from Source

```bash
# Build CLI
make build

# Run tests
make test

# Install locally
make install

# Build frontend
make frontend-build

# Run frontend in dev mode (with hot reload)
make frontend-dev

# Serve crawl results
make serve
```

### Project Structure

```
baracuda/
├── cmd/              # CLI commands
│   ├── crawl.go     # Crawl command
│   └── serve.go     # Serve command (web dashboard)
├── internal/
│   ├── analyzer/    # SEO analysis and issue detection
│   ├── crawler/     # Crawling logic
│   ├── exporter/    # Export formats
│   ├── graph/       # Link graph
│   └── utils/       # Utilities
├── pkg/
│   └── models/      # Data models
├── web/             # Svelte frontend
│   ├── src/
│   │   ├── components/  # Svelte components
│   │   └── App.svelte   # Main app component
│   └── package.json
└── main.go          # Entry point
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Acknowledgments

Inspired by Screaming Frog SEO Spider.


# Baracuda Web Frontend

This is the Svelte frontend for viewing crawl results.

## Setup

```bash
cd web
npm install
```

## Development

```bash
npm run dev
```

This will start the Vite dev server on port 5173. The frontend will proxy API requests to `http://localhost:8080` (the Go server).

## Building for Production

```bash
npm run build
```

This creates a `dist/` folder with optimized static files that can be served by the Go server.

## Usage

1. Run a crawl with JSON export:
   ```bash
   baracuda crawl https://example.com --format json --export results.json --graph-export graph.json
   ```

2. Build the frontend:
   ```bash
   make frontend-build
   ```

3. Serve the results:
   ```bash
   baracuda serve --results results.json --graph graph.json
   ```

Or use the Makefile shortcut:
```bash
make serve
```

## Features

- **Dashboard**: Overview with statistics and metrics
- **Results Table**: Filterable, sortable table of all crawled pages
- **Issues Panel**: Detailed view of SEO issues with recommendations
- **Link Graph**: Visualization of internal/external link structure


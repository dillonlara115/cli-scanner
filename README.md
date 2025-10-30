# sfrog

sfrog is a lightweight, Screaming Frog–style site crawler designed for collecting SEO insights from websites. It provides a simple command line interface and outputs crawl data to CSV or JSON while producing a link graph for further analysis.

## Features

- Async HTML crawling with configurable concurrency
- Respects `robots.txt` rules and seeds with `sitemap.xml`
- Extracts titles, meta descriptions, headings, canonical URLs, and links
- Detects duplicate pages by hashing HTML content
- Builds a directed link graph (`.gexf`) for visualization tools
- Exports crawl data to CSV or JSON
- Displays live crawl progress using Rich

## Installation

1. Ensure Python 3.11 or later is installed.
2. Install dependencies:

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Usage

```bash
python -m sfrog.cli crawl https://example.com --max-depth 3 --threads 10 --export csv
```

### Options

- `--max-depth` – Maximum crawl depth (default: 2)
- `--threads` – Number of concurrent requests (default: 10)
- `--export` – Export format (`csv` or `json`)
- `--output` – Output file path (defaults to `results.csv` or `results.json`)
- `--config` – Optional JSON/YAML config file providing defaults

Upon completion the tool prints a summary containing totals, broken pages, duplicate groups, and the export locations.

## Configuration File

Supply defaults via JSON or YAML. Example:

```yaml
max_depth: 3
threads: 15
user_agent: "sfrog/1.0"
export: json
```

## Development

Run `ruff` or your preferred linters to keep code quality high. Contributions are welcome!

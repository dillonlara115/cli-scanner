package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dillonlara115/baracuda/internal/analyzer"
	"github.com/dillonlara115/baracuda/internal/crawler"
	"github.com/dillonlara115/baracuda/internal/exporter"
	"github.com/dillonlara115/baracuda/internal/graph"
	"github.com/dillonlara115/baracuda/internal/utils"
	"github.com/dillonlara115/baracuda/pkg/models"
	"github.com/spf13/cobra"
)

var (
	startURL      string
	maxDepth      int
	maxPages      int
	workers       int
	delay         time.Duration
	timeout       time.Duration
	userAgent     string
	respectRobots bool
	parseSitemap  bool
	exportFormat  string
	exportPath    string
	domainFilter  string
	graphExport   string
	interactive   bool
	openBrowser   bool
)

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl [URL]",
	Short: "Crawl a website and extract SEO data",
	Long: `Crawl a website recursively and extract SEO data including titles, meta descriptions,
headings (H1-H6), canonical tags, and internal/external links. Results are exported to CSV or JSON.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCrawl,
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	// URL flag (optional - can also be provided as positional argument)
	crawlCmd.Flags().StringVarP(&startURL, "url", "u", "", "Starting URL to crawl")

	// Crawl options
	crawlCmd.Flags().IntVarP(&maxDepth, "max-depth", "d", 3, "Maximum crawl depth")
	crawlCmd.Flags().IntVarP(&maxPages, "max-pages", "p", 1000, "Maximum number of pages to crawl")
	crawlCmd.Flags().IntVarP(&workers, "workers", "w", 10, "Number of concurrent workers")
	crawlCmd.Flags().DurationVar(&delay, "delay", 0, "Delay between requests (e.g., 100ms)")
	crawlCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	crawlCmd.Flags().StringVar(&userAgent, "user-agent", "baracuda/1.0.0", "User agent string")
	crawlCmd.Flags().BoolVar(&respectRobots, "respect-robots", true, "Respect robots.txt")
	crawlCmd.Flags().BoolVar(&parseSitemap, "parse-sitemap", false, "Parse sitemap.xml for seed URLs")
	crawlCmd.Flags().StringVar(&domainFilter, "domain-filter", "same", "Domain filter: 'same' or 'all'")

	// Export options
	crawlCmd.Flags().StringVarP(&exportFormat, "format", "f", "csv", "Export format: 'csv' or 'json'")
	crawlCmd.Flags().StringVarP(&exportPath, "export", "e", "", "Export file path (default: stdout or results.csv/json)")
	crawlCmd.Flags().StringVar(&graphExport, "graph-export", "", "Export link graph to JSON file")
	
	// Interactive mode
	crawlCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode with prompts")
	
	// Browser options
	crawlCmd.Flags().BoolVarP(&openBrowser, "open", "o", true, "Automatically open web dashboard in browser after crawl")
}

func runCrawl(cmd *cobra.Command, args []string) error {
	// Check if we should run in interactive mode
	// Interactive if: flag is set, OR no URL provided and no flags set
	shouldRunInteractive := interactive
	if !shouldRunInteractive && startURL == "" && len(args) == 0 {
		// Check if any flags were provided
		hasFlags := maxDepth != 3 || maxPages != 1000 || workers != 10 || exportFormat != "csv" || 
			exportPath != "" || graphExport != "" || respectRobots != true || parseSitemap != false
		if !hasFlags {
			shouldRunInteractive = true
		}
	}
	
	var crawlDir string
	
	if shouldRunInteractive {
		// Run interactive prompts
		config, graphExportPath, dir, shouldOpen, err := utils.PromptInteractive()
		if err != nil {
			return fmt.Errorf("interactive setup failed: %w", err)
		}
		
		// Use config from prompts
		startURL = config.StartURL
		maxDepth = config.MaxDepth
		maxPages = config.MaxPages
		workers = config.Workers
		exportFormat = config.ExportFormat
		exportPath = config.ExportPath
		respectRobots = config.RespectRobots
		parseSitemap = config.ParseSitemap
		graphExport = graphExportPath
		crawlDir = dir
		openBrowser = shouldOpen // Use interactive preference
	} else {
		// Get URL from positional argument or flag
		if len(args) > 0 {
			startURL = args[0]
		}
		
		// Validate that URL is provided
		if startURL == "" {
			return fmt.Errorf("starting URL is required. Provide it as an argument, use --url flag, or run with --interactive")
		}
	}

	// Initialize logger
	if err := utils.InitLogger(debug); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer utils.Sync()

	// Create config
	config := &utils.Config{
		StartURL:      startURL,
		MaxDepth:      maxDepth,
		MaxPages:      maxPages,
		Workers:       workers,
		Delay:         delay,
		Timeout:       timeout,
		UserAgent:     userAgent,
		RespectRobots: respectRobots,
		ParseSitemap:  parseSitemap,
		ExportFormat:  exportFormat,
		ExportPath:    exportPath,
		DomainFilter:  domainFilter,
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Set default export path if not provided
	if config.ExportPath == "" {
		ext := "csv"
		if config.ExportFormat == "json" {
			ext = "json"
		}
		config.ExportPath = fmt.Sprintf("results.%s", ext)
	}

	utils.Info("Starting crawl", utils.NewField("url", config.StartURL))

	// Create crawler manager
	manager := crawler.NewManager(config)

	// Start crawling
	results, err := manager.Crawl()
	if err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	utils.Info("Crawl completed", utils.NewField("pages_crawled", len(results)))

	// Analyze results and print summary (including image size checking)
	summary := analyzer.AnalyzeWithImages(results, config.Timeout)
	analyzer.PrintSummary(summary)

	// Export results
	if err := exportResults(results, config); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Export link graph if requested
	if graphExport != "" {
		if err := exportLinkGraph(manager.GetLinkGraph(), graphExport); err != nil {
			return fmt.Errorf("graph export failed: %w", err)
		}
		fmt.Fprintf(os.Stdout, "✓ Link graph exported to %s\n", graphExport)
	}

	fmt.Fprintf(os.Stdout, "\n✓ Crawled %d pages\n", len(results))
	fmt.Fprintf(os.Stdout, "✓ Results exported to %s\n", config.ExportPath)
	
	if crawlDir != "" {
		fmt.Fprintf(os.Stdout, "📁 All files saved to: %s\n", crawlDir)
	}

	// Optionally open browser with dashboard
	if openBrowser {
		fmt.Fprintf(os.Stdout, "\n")
		if err := startServerAndOpenBrowser(config.ExportPath, graphExport); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to start server: %v\n", err)
			fmt.Fprintf(os.Stderr, "   You can manually run: baracuda serve --results %s", config.ExportPath)
			if graphExport != "" {
				fmt.Fprintf(os.Stderr, " --graph %s", graphExport)
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	return nil
}

func exportLinkGraph(graph *graph.Graph, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create graph file: %w", err)
	}
	defer file.Close()

	edges := graph.GetAllEdges()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(edges); err != nil {
		return fmt.Errorf("failed to encode graph JSON: %w", err)
	}

	return nil
}

func exportResults(results []*models.PageResult, config *utils.Config) error {
	switch config.ExportFormat {
	case "csv":
		return exporter.ExportCSV(results, config.ExportPath)
	case "json":
		return exporter.ExportJSON(results, config.ExportPath, true)
	default:
		return fmt.Errorf("unsupported export format: %s", config.ExportFormat)
	}
}


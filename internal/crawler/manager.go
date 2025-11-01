package crawler

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dillonlara115/baracuda/internal/graph"
	"github.com/dillonlara115/baracuda/internal/utils"
	"github.com/dillonlara115/baracuda/pkg/models"
)

// Manager orchestrates the crawling process
type Manager struct {
	config        *utils.Config
	fetcher       *Fetcher
	robotsChecker *RobotsChecker
	sitemapParser *SitemapParser
	linkGraph     *graph.Graph
	visited       sync.Map // map[string]bool for visited URLs
	queue         chan crawlTask
	results       []*models.PageResult
	resultsMu     sync.Mutex
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	pending       int32 // Track pending tasks (atomic)
	queueClosed   int32 // Atomic flag to track if queue is closed
}

// crawlTask represents a URL to be crawled with its depth
type crawlTask struct {
	URL   string
	Depth int
}

// NewManager creates a new Manager instance
func NewManager(config *utils.Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:  config,
		fetcher: NewFetcher(config.Timeout, config.UserAgent),
		queue:   make(chan crawlTask, config.MaxPages*2), // Buffer for queue
		results: make([]*models.PageResult, 0, config.MaxPages),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize robots checker
	manager.robotsChecker = NewRobotsChecker(manager.fetcher, config.UserAgent, config.RespectRobots)

	// Initialize sitemap parser
	manager.sitemapParser = NewSitemapParser(manager.fetcher)

	// Initialize link graph
	manager.linkGraph = graph.NewGraph()

	// Setup graceful shutdown
	go manager.handleSignals()

	return manager
}

// Crawl starts the crawling process
func (m *Manager) Crawl() ([]*models.PageResult, error) {
	// Normalize start URL
	startURL, err := utils.NormalizeURL(m.config.StartURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	// Parse sitemap if enabled
	var seedURLs []string
	if m.config.ParseSitemap {
		sitemapURL := m.sitemapParser.DiscoverSitemapURL(startURL)
		utils.Info("Parsing sitemap", utils.NewField("url", sitemapURL))
		
		urls, err := m.sitemapParser.ParseSitemap(sitemapURL)
		if err != nil {
			utils.Debug("Failed to parse sitemap", utils.NewField("url", sitemapURL), utils.NewField("error", err.Error()))
		} else {
			seedURLs = urls
			utils.Info("Found URLs in sitemap", utils.NewField("count", len(seedURLs)))
		}
	}

	// If no sitemap URLs found, use start URL
	if len(seedURLs) == 0 {
		seedURLs = []string{startURL}
	}

	// Start worker pool
	for i := 0; i < m.config.Workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}

	// Enqueue initial tasks (don't mark as visited yet - workers will do that)
	enqueueDone := make(chan bool)
	go func() {
		defer close(enqueueDone)
		for _, url := range seedURLs {
			// Normalize URL
			normalized, err := utils.NormalizeURL(url)
			if err != nil {
				utils.Debug("Failed to normalize seed URL", utils.NewField("url", url), utils.NewField("error", err.Error()))
				continue
			}
			
			atomic.AddInt32(&m.pending, 1)
			m.queue <- crawlTask{
				URL:   normalized,
				Depth: 0,
			}
		}
	}()

	// Wait for initial enqueueing to complete
	<-enqueueDone
	utils.Debug("Initial tasks enqueued", utils.NewField("count", len(seedURLs)))

	// Monitor queue and close when done
	go m.monitorQueue()

	// Wait for all workers to finish
	m.wg.Wait()

	// Return results - don't treat cancellation as error if we got results
	// (cancellation might be due to reaching max-pages, which is success)
	if m.ctx.Err() != nil && len(m.results) == 0 {
		return m.results, fmt.Errorf("crawl cancelled: %w", m.ctx.Err())
	}

	return m.results, nil
}

// GetLinkGraph returns the link graph
func (m *Manager) GetLinkGraph() *graph.Graph {
	return m.linkGraph
}

// worker processes crawl tasks from the queue
func (m *Manager) worker(id int) {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			utils.Debug("Worker stopping", utils.NewField("worker_id", id))
			return
		case task, ok := <-m.queue:
			if !ok {
				utils.Debug("Worker queue closed", utils.NewField("worker_id", id))
				return
			}

			// Decrement pending counter
			atomic.AddInt32(&m.pending, -1)

			// Check if we've reached max pages BEFORE processing
			m.resultsMu.Lock()
			if len(m.results) >= m.config.MaxPages {
				m.resultsMu.Unlock()
				// Cancel to signal other workers to stop
				m.cancel()
				return
			}
			m.resultsMu.Unlock()

			// Check depth limit
			if task.Depth > m.config.MaxDepth {
				continue
			}

			// Check if already visited (before marking to avoid race condition)
			if _, visited := m.visited.LoadOrStore(task.URL, true); visited {
				continue
			}

			// Check robots.txt before fetching
			if allowed, err := m.robotsChecker.IsAllowed(task.URL); err != nil {
				utils.Debug("Robots check error", utils.NewField("url", task.URL), utils.NewField("error", err.Error()))
			} else if !allowed {
				utils.Debug("URL disallowed by robots.txt", utils.NewField("url", task.URL))
				continue
			}

			// Apply delay if configured
			if m.config.Delay > 0 {
				select {
				case <-m.ctx.Done():
					return
				case <-time.After(m.config.Delay):
				}
			}

			// Fetch the URL with retry logic
			result := m.fetcher.FetchWithRetry(task.URL, 3)

			// Store result (check limit again before storing)
			m.resultsMu.Lock()
			resultCount := len(m.results)
			if resultCount >= m.config.MaxPages {
				m.resultsMu.Unlock()
				m.cancel()
				return
			}
			m.results = append(m.results, result.PageResult)
			resultCount = len(m.results)
			m.resultsMu.Unlock()

			utils.Info("Crawled page",
				utils.NewField("url", task.URL),
				utils.NewField("status", result.PageResult.StatusCode),
				utils.NewField("depth", task.Depth),
				utils.NewField("total", resultCount),
			)

			// Check if we've reached max pages after storing
			if resultCount >= m.config.MaxPages {
				m.cancel()
				return
			}

			// If fetch failed or not HTML, don't discover links
			if result.Error != nil || result.PageResult.StatusCode != 200 {
				continue
			}

			// Parse HTML and discover links
			parser, err := NewParser(task.URL)
			if err != nil {
				utils.Error("Failed to create parser", utils.NewField("url", task.URL), utils.NewField("error", err.Error()))
				continue
			}

			// Merge parsed SEO data into result
			parsedData, err := parser.Parse(result.Body)
			if err != nil {
				utils.Error("Failed to parse HTML", utils.NewField("url", task.URL), utils.NewField("error", err.Error()))
				continue
			}

			// Merge parsed data into page result
			result.PageResult.Title = parsedData.Title
			result.PageResult.MetaDesc = parsedData.MetaDesc
			result.PageResult.Canonical = parsedData.Canonical
			result.PageResult.H1 = parsedData.H1
			result.PageResult.H2 = parsedData.H2
			result.PageResult.H3 = parsedData.H3
			result.PageResult.H4 = parsedData.H4
			result.PageResult.H5 = parsedData.H5
			result.PageResult.H6 = parsedData.H6
			result.PageResult.InternalLinks = parsedData.InternalLinks
			result.PageResult.ExternalLinks = parsedData.ExternalLinks

			// Add edges to link graph
			m.linkGraph.AddEdges(task.URL, parsedData.InternalLinks)
			m.linkGraph.AddEdges(task.URL, parsedData.ExternalLinks)

			// Enqueue discovered internal links for crawling
			if task.Depth < m.config.MaxDepth {
				for _, linkURL := range parsedData.InternalLinks {
					// Check domain filter
					if m.config.DomainFilter == "same" && !utils.IsSameDomain(linkURL, m.config.StartURL) {
						continue
					}

					// Check if already visited
					if _, visited := m.visited.Load(linkURL); visited {
						continue
					}

					// Enqueue new task (check if queue is still open)
					// Check if queue is closed before attempting to send
					if atomic.LoadInt32(&m.queueClosed) == 1 {
						return
					}
					
					select {
					case <-m.ctx.Done():
						return
					case m.queue <- crawlTask{URL: linkURL, Depth: task.Depth + 1}:
						// Successfully enqueued
						atomic.AddInt32(&m.pending, 1)
					default:
						// Queue full, skip (but don't panic)
						utils.Debug("Queue full, skipping link", utils.NewField("url", linkURL))
					}
				}
			}

			// Check if we've reached max pages
			if resultCount >= m.config.MaxPages {
				m.cancel()
				return
			}
		}
	}
}

// monitorQueue closes the queue when all tasks are processed
func (m *Manager) monitorQueue() {
	ticker := time.NewTicker(500 * time.Millisecond) // Check less frequently
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			atomic.StoreInt32(&m.queueClosed, 1)
			close(m.queue)
			return
		case <-ticker.C:
			// Check if queue is empty and no pending tasks
			// Wait a bit longer to ensure workers have finished processing
			pending := atomic.LoadInt32(&m.pending)
			queueLen := len(m.queue)
			
			if pending <= 0 && queueLen == 0 {
				// Give workers more time to finish processing and discover links
				time.Sleep(1 * time.Second)
				
				// Check again - if still empty, close the queue
				pending = atomic.LoadInt32(&m.pending)
				queueLen = len(m.queue)
				
				if pending <= 0 && queueLen == 0 {
					// Final check - wait a bit more to be safe
					time.Sleep(500 * time.Millisecond)
					pending = atomic.LoadInt32(&m.pending)
					queueLen = len(m.queue)
					
					if pending <= 0 && queueLen == 0 {
						utils.Debug("Closing queue - no pending tasks")
						atomic.StoreInt32(&m.queueClosed, 1)
						close(m.queue)
						return
					}
				}
			}
		}
	}
}

// handleSignals sets up graceful shutdown on interrupt signals
func (m *Manager) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	utils.Info("Received interrupt signal, shutting down gracefully...")
	m.cancel()
}


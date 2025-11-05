package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dillonlara115/barracuda/internal/analyzer"
	"github.com/dillonlara115/barracuda/internal/crawler"
	"github.com/dillonlara115/barracuda/internal/gsc"
	"github.com/dillonlara115/barracuda/internal/utils"
	"github.com/dillonlara115/barracuda/pkg/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleCrawls handles crawl-related endpoints
func (s *Server) handleCrawls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateCrawl(w, r)
	case http.MethodGet:
		s.handleListCrawls(w, r)
	default:
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCreateCrawl handles POST /api/v1/crawls - crawl ingestion
func (s *Server) handleCreateCrawl(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		s.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateCrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate required fields
	if req.ProjectID == "" {
		s.respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	if len(req.Pages) == 0 {
		s.respondError(w, http.StatusBadRequest, "pages array cannot be empty")
		return
	}

	// Verify user has access to project
	hasAccess, err := s.verifyProjectAccess(userID, req.ProjectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to verify project access")
		return
	}
	if !hasAccess {
		s.respondError(w, http.StatusForbidden, "You don't have access to this project")
		return
	}

	// Analyze pages to detect issues
	summary := analyzer.AnalyzeWithImages(req.Pages, 30*time.Second)

	// Create crawl record
	crawlID := uuid.New().String()
	crawl := map[string]interface{}{
		"id":            crawlID,
		"project_id":    req.ProjectID,
		"initiated_by":  userID,
		"source":        "cli",
		"status":        "succeeded",
		"started_at":    time.Now().UTC().Format(time.RFC3339),
		"completed_at":  time.Now().UTC().Format(time.RFC3339),
		"total_pages":  len(req.Pages),
		"total_issues": len(summary.Issues),
		"meta": map[string]interface{}{
			"user_agent": r.Header.Get("User-Agent"),
		},
	}

	// Insert crawl using service role (bypasses RLS)
	_, _, err = s.serviceRole.From("crawls").Insert(crawl, false, "", "", "").Execute()
	if err != nil {
		s.logger.Error("Failed to insert crawl", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create crawl")
		return
	}

	// Insert pages in batch
	pages := make([]map[string]interface{}, 0, len(req.Pages))
	for _, page := range req.Pages {
		pageData := map[string]interface{}{
			"crawl_id":        crawlID,
			"url":             page.URL,
			"status_code":     page.StatusCode,
			"response_time_ms": page.ResponseTime,
			"title":           page.Title,
			"meta_description": page.MetaDesc,
			"canonical_url":   page.Canonical,
			"h1":              strings.Join(page.H1, ", "),
			"word_count":      0, // TODO: calculate from content
			"data": map[string]interface{}{
				"h2":            page.H2,
				"h3":            page.H3,
				"h4":            page.H4,
				"h5":            page.H5,
				"h6":            page.H6,
				"internal_links": page.InternalLinks,
				"external_links": page.ExternalLinks,
				"images":         page.Images,
			},
		}
		pages = append(pages, pageData)
	}

	// Batch insert pages (Supabase supports up to 1000 rows per insert)
	batchSize := 1000
	for i := 0; i < len(pages); i += batchSize {
		end := i + batchSize
		if end > len(pages) {
			end = len(pages)
		}
		batch := pages[i:end]

		_, _, err = s.serviceRole.From("pages").Insert(batch, false, "", "", "").Execute()
		if err != nil {
			s.logger.Error("Failed to insert pages batch", zap.Int("batch_start", i), zap.Error(err))
			// Continue with other batches, but log error
		}
	}

	// Insert issues
	issues := make([]map[string]interface{}, 0, len(summary.Issues))
	for _, issue := range summary.Issues {
		// Find page ID for this issue
		var pageID *int64
		for i, page := range req.Pages {
			if page.URL == issue.URL {
				// We'd need to fetch the page ID from the database
				// For now, we'll insert without page_id and update later if needed
				_ = i
				break
			}
		}

		issueData := map[string]interface{}{
			"crawl_id":     crawlID,
			"project_id":   req.ProjectID,
			"type":         string(issue.Type),
			"severity":     issue.Severity,
			"message":      issue.Message,
			"recommendation": issue.Recommendation,
			"value":        issue.Value,
			"status":       "new",
		}
		if pageID != nil {
			issueData["page_id"] = *pageID
		}
		issues = append(issues, issueData)
	}

	if len(issues) > 0 {
		// Batch insert issues
		for i := 0; i < len(issues); i += batchSize {
			end := i + batchSize
			if end > len(issues) {
				end = len(issues)
			}
			batch := issues[i:end]

			_, _, err = s.serviceRole.From("issues").Insert(batch, false, "", "", "").Execute()
			if err != nil {
				s.logger.Error("Failed to insert issues batch", zap.Int("batch_start", i), zap.Error(err))
			}
		}
	}

	// Return crawl response
	response := CreateCrawlResponse{
		CrawlID:     crawlID,
		ProjectID:   req.ProjectID,
		TotalPages:  len(req.Pages),
		TotalIssues: len(summary.Issues),
		Status:      "succeeded",
	}

	s.respondJSON(w, http.StatusCreated, response)
}

// handleListCrawls handles GET /api/v1/crawls - list crawls
func (s *Server) handleListCrawls(w http.ResponseWriter, r *http.Request) {
	_, ok := userIDFromContext(r.Context())
	if !ok {
		s.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get project_id from query params (optional filter)
	projectID := r.URL.Query().Get("project_id")

	// Build query - user can only see crawls from projects they're a member of
	query := s.supabase.From("crawls").Select("*", "", false)

	if projectID != "" {
		query = query.Eq("project_id", projectID)
	}

	// The RLS policies will automatically filter to only projects the user has access to
	var crawls []map[string]interface{}
	data, _, err := query.Execute()
	if err != nil {
		s.logger.Error("Failed to list crawls", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to list crawls")
		return
	}

	// Parse data into crawls slice
	if err := json.Unmarshal(data, &crawls); err != nil {
		s.logger.Error("Failed to parse crawls data", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to parse crawls")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"crawls": crawls,
		"count":  len(crawls),
	})
}

// handleProjects handles project-related endpoints
func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateProject(w, r)
	case http.MethodGet:
		s.handleListProjects(w, r)
	default:
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCreateProject handles POST /api/v1/projects
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		s.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.Name == "" || req.Domain == "" {
		s.respondError(w, http.StatusBadRequest, "name and domain are required")
		return
	}

	project := map[string]interface{}{
		"name":     req.Name,
		"domain":   req.Domain,
		"owner_id": userID,
		"settings": req.Settings,
	}

	var result []map[string]interface{}
	data, _, err := s.supabase.From("projects").Insert(project, false, "", "", "").Execute()
	if err != nil {
		s.logger.Error("Failed to create project", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	// Parse data into result slice
	if err := json.Unmarshal(data, &result); err != nil {
		s.logger.Error("Failed to parse project data", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to parse project")
		return
	}

	if len(result) == 0 {
		s.respondError(w, http.StatusInternalServerError, "Project created but no data returned")
		return
	}

	// Also add the owner as a project member with 'owner' role
	member := map[string]interface{}{
		"project_id": result[0]["id"],
		"user_id":    userID,
		"role":       "owner",
	}
	_, _, err = s.serviceRole.From("project_members").Insert(member, false, "", "", "").Execute()
	if err != nil {
		s.logger.Warn("Failed to add owner as project member", zap.Error(err))
		// Continue anyway - the project was created
	}

	s.respondJSON(w, http.StatusCreated, result[0])
}

// handleListProjects handles GET /api/v1/projects
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	_, ok := userIDFromContext(r.Context())
	if !ok {
		s.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// RLS policies will automatically filter to projects user has access to
	var projects []map[string]interface{}
	data, _, err := s.supabase.From("projects").Select("*", "", false).Execute()
	if err != nil {
		s.logger.Error("Failed to list projects", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	// Parse data into projects slice
	if err := json.Unmarshal(data, &projects); err != nil {
		s.logger.Error("Failed to parse projects data", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to parse projects")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
}

// handleProjectByID handles project operations by ID
func (s *Server) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from path
	// After StripPrefix("/api/v1"), the path is like "/projects/:id/crawl"
	s.logger.Debug("handleProjectByID called", zap.String("path", r.URL.Path), zap.String("method", r.Method))
	
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	
	// Remove leading/trailing slashes and split
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	
	s.logger.Debug("Path parsing", zap.String("trimmed_path", path), zap.Strings("parts", parts))
	
	if len(parts) == 0 {
		s.respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}
	
	projectID := parts[0]

	if projectID == "" {
		s.respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok {
		s.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Handle sub-resources like /projects/:id/crawls or /projects/:id/crawl
	if len(parts) > 1 {
		resource := parts[1]
		switch resource {
		case "crawls":
			if r.Method == http.MethodGet {
				s.handleListProjectCrawls(w, r, projectID, userID)
			} else {
				s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		case "crawl":
			if r.Method == http.MethodPost {
				s.handleTriggerCrawl(w, r, projectID, userID)
			} else {
				s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		default:
			s.logger.Debug("Unknown resource", zap.String("resource", resource), zap.String("path", r.URL.Path), zap.Strings("parts", parts))
			s.respondError(w, http.StatusNotFound, fmt.Sprintf("Resource not found: %s", resource))
			return
		}
	}

	// Handle main project operations
	switch r.Method {
	case http.MethodGet:
		s.handleGetProject(w, r, projectID, userID)
	default:
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetProject handles GET /api/v1/projects/:id
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request, projectID, userID string) {
	// Verify access
	hasAccess, err := s.verifyProjectAccess(userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to verify project access")
		return
	}
	if !hasAccess {
		s.respondError(w, http.StatusForbidden, "You don't have access to this project")
		return
	}

	var projects []map[string]interface{}
	data, _, err := s.supabase.From("projects").Select("*", "", false).Eq("id", projectID).Execute()
	if err != nil {
		s.logger.Error("Failed to get project", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to get project")
		return
	}

	// Parse data into projects slice
	if err := json.Unmarshal(data, &projects); err != nil {
		s.logger.Error("Failed to parse project data", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to parse project")
		return
	}

	if len(projects) == 0 {
		s.respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	s.respondJSON(w, http.StatusOK, projects[0])
}

// handleListProjectCrawls handles GET /api/v1/projects/:id/crawls
func (s *Server) handleListProjectCrawls(w http.ResponseWriter, r *http.Request, projectID, userID string) {
	// Verify access
	hasAccess, err := s.verifyProjectAccess(userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to verify project access")
		return
	}
	if !hasAccess {
		s.respondError(w, http.StatusForbidden, "You don't have access to this project")
		return
	}

	var crawls []map[string]interface{}
	data, _, err := s.supabase.From("crawls").Select("*", "", false).Eq("project_id", projectID).Order("started_at", nil).Execute()
	if err != nil {
		s.logger.Error("Failed to list project crawls", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to list crawls")
		return
	}

	// Parse data into crawls slice
	if err := json.Unmarshal(data, &crawls); err != nil {
		s.logger.Error("Failed to parse crawls data", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to parse crawls")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"crawls": crawls,
		"count":  len(crawls),
	})
}

// handleExports handles export-related endpoints
func (s *Server) handleExports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: Implement export generation
	s.respondError(w, http.StatusNotImplemented, "Export functionality not yet implemented")
}

// handleTriggerCrawl handles POST /api/v1/projects/:id/crawl - trigger a new crawl
func (s *Server) handleTriggerCrawl(w http.ResponseWriter, r *http.Request, projectID, userID string) {
	// Verify access
	hasAccess, err := s.verifyProjectAccess(userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to verify project access")
		return
	}
	if !hasAccess {
		s.respondError(w, http.StatusForbidden, "You don't have access to this project")
		return
	}

	var req TriggerCrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate and set defaults
	if req.URL == "" {
		s.respondError(w, http.StatusBadRequest, "url is required")
		return
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}
	if req.MaxPages == 0 {
		req.MaxPages = 1000
	}
	if req.Workers == 0 {
		req.Workers = 10
	}

	// Create crawl record with status "running"
	crawlID := uuid.New().String()
	crawl := map[string]interface{}{
		"id":            crawlID,
		"project_id":    projectID,
		"initiated_by":  userID,
		"source":        "web",
		"status":        "running",
		"started_at":    time.Now().UTC().Format(time.RFC3339),
		"total_pages":   0,
		"total_issues":  0,
		"meta": map[string]interface{}{
			"url":            req.URL,
			"max_depth":      req.MaxDepth,
			"max_pages":      req.MaxPages,
			"workers":        req.Workers,
			"respect_robots": req.RespectRobots,
			"parse_sitemap":  req.ParseSitemap,
		},
	}

	// Insert crawl using service role (bypasses RLS)
	_, _, err = s.serviceRole.From("crawls").Insert(crawl, false, "", "", "").Execute()
	if err != nil {
		s.logger.Error("Failed to insert crawl", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create crawl")
		return
	}

	// Start crawl asynchronously
	go s.runCrawlAsync(crawlID, projectID, req)

	// Return immediately with crawl ID
	s.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"crawl_id": crawlID,
		"status":   "running",
		"message":  "Crawl started",
	})
}

// runCrawlAsync runs the crawler and stores results
func (s *Server) runCrawlAsync(crawlID, projectID string, req TriggerCrawlRequest) {
	// Initialize logger for crawler
	if err := utils.InitLogger(false); err != nil {
		s.logger.Error("Failed to initialize logger", zap.Error(err))
		s.updateCrawlStatus(crawlID, "failed", fmt.Sprintf("Failed to initialize logger: %v", err))
		return
	}
	defer utils.Sync()

	// Create crawler config
	config := &utils.Config{
		StartURL:      req.URL,
		MaxDepth:      req.MaxDepth,
		MaxPages:      req.MaxPages,
		Workers:       req.Workers,
		Delay:         0,
		Timeout:       30 * time.Second,
		UserAgent:     "barracuda/1.0.0",
		RespectRobots: req.RespectRobots,
		ParseSitemap:  req.ParseSitemap,
		DomainFilter:  "same",
		ExportFormat:  "csv", // Required for validation, but not used since we store in DB
		ExportPath:    "",     // Not used for web crawls
	}

	// Validate config
	if err := config.Validate(); err != nil {
		s.logger.Error("Invalid crawl config", zap.Error(err))
		s.updateCrawlStatus(crawlID, "failed", err.Error())
		return
	}

	// Create crawler manager
	manager := crawler.NewManager(config)

	// Track pages and page URL to ID mapping for real-time storage
	batchSize := 50 // Smaller batches for more frequent updates
	pages := make([]map[string]interface{}, 0, batchSize)
	pageURLToID := make(map[string]int64)
	var pagesMu sync.Mutex
	totalPagesProcessed := int32(0)

	// Set up progress callback to store pages in real-time
	manager.SetProgressCallback(func(page *models.PageResult, totalPages int) {
		pagesMu.Lock()
		defer pagesMu.Unlock()

		pageData := map[string]interface{}{
			"crawl_id":         crawlID,
			"url":              page.URL,
			"status_code":      page.StatusCode,
			"response_time_ms": page.ResponseTime,
			"title":            page.Title,
			"meta_description": page.MetaDesc,
			"canonical_url":    page.Canonical,
			"h1":               strings.Join(page.H1, ", "),
			"word_count":       0, // TODO: calculate from content
			"data": map[string]interface{}{
				"h2":            page.H2,
				"h3":            page.H3,
				"h4":            page.H4,
				"h5":            page.H5,
				"h6":            page.H6,
				"internal_links": page.InternalLinks,
				"external_links": page.ExternalLinks,
				"images":         page.Images,
			},
		}
		pages = append(pages, pageData)

		// Insert in batches and update progress
		if len(pages) >= batchSize {
			var pageResults []map[string]interface{}
			data, _, err := s.serviceRole.From("pages").Insert(pages, false, "", "", "").Execute()
			if err != nil {
				s.logger.Error("Failed to insert pages batch", zap.Error(err))
			} else {
				// Parse inserted pages to get IDs
				if err := json.Unmarshal(data, &pageResults); err == nil {
					for j, pageResult := range pageResults {
						if pageID, ok := pageResult["id"].(float64); ok {
							pageURLToID[pages[j]["url"].(string)] = int64(pageID)
						}
					}
				}

				// Update crawl total_pages in real-time
				atomic.AddInt32(&totalPagesProcessed, int32(len(pages)))
				currentTotal := int(atomic.LoadInt32(&totalPagesProcessed))
				update := map[string]interface{}{
					"total_pages": currentTotal,
				}
				_, _, err = s.serviceRole.From("crawls").Update(update, "", "").Eq("id", crawlID).Execute()
				if err != nil {
					s.logger.Warn("Failed to update crawl progress", zap.Error(err))
				} else {
					s.logger.Debug("Updated crawl progress", zap.Int("total_pages", currentTotal))
				}
			}
			pages = make([]map[string]interface{}, 0, batchSize)
		}
	})

	// Run crawl
	results, err := manager.Crawl()
	if err != nil {
		s.logger.Error("Crawl failed", zap.Error(err))
		s.updateCrawlStatus(crawlID, "failed", err.Error())
		return
	}

	// Store any remaining pages
	pagesMu.Lock()
	if len(pages) > 0 {
		var pageResults []map[string]interface{}
		data, _, err := s.serviceRole.From("pages").Insert(pages, false, "", "", "").Execute()
		if err != nil {
			s.logger.Error("Failed to insert final pages batch", zap.Error(err))
		} else {
			// Parse inserted pages to get IDs
			if err := json.Unmarshal(data, &pageResults); err == nil {
				for j, pageResult := range pageResults {
					if pageID, ok := pageResult["id"].(float64); ok {
						pageURLToID[pages[j]["url"].(string)] = int64(pageID)
					}
				}
			}
		}
	}
	finalTotal := int(atomic.LoadInt32(&totalPagesProcessed))
	pagesMu.Unlock()

	// Analyze results
	summary := analyzer.AnalyzeWithImages(results, config.Timeout)

	// Store issues
	issues := make([]map[string]interface{}, 0, len(summary.Issues))
	for _, issue := range summary.Issues {
		issueData := map[string]interface{}{
			"crawl_id":      crawlID,
			"project_id":    projectID,
			"type":          string(issue.Type),
			"severity":      issue.Severity,
			"message":       issue.Message,
			"recommendation": issue.Recommendation,
			"value":         issue.Value,
			"status":        "new",
		}
		// Try to find page ID
		if pageID, ok := pageURLToID[issue.URL]; ok {
			issueData["page_id"] = pageID
		}
		issues = append(issues, issueData)
	}

	if len(issues) > 0 {
		// Batch insert issues
		for i := 0; i < len(issues); i += batchSize {
			end := i + batchSize
			if end > len(issues) {
				end = len(issues)
			}
			batch := issues[i:end]

			_, _, err = s.serviceRole.From("issues").Insert(batch, false, "", "", "").Execute()
			if err != nil {
				s.logger.Error("Failed to insert issues batch", zap.Int("batch_start", i), zap.Error(err))
			}
		}
	}

	// Update crawl status to succeeded (total_pages already updated via callback)
	s.updateCrawlStatus(crawlID, "succeeded", "")
	update := map[string]interface{}{
		"total_pages":  finalTotal, // Use the final count from callback
		"total_issues": len(summary.Issues),
		"completed_at": time.Now().UTC().Format(time.RFC3339),
	}
	_, _, err = s.serviceRole.From("crawls").Update(update, "", "").Eq("id", crawlID).Execute()
	if err != nil {
		s.logger.Error("Failed to update crawl stats", zap.Error(err))
	}
}

// updateCrawlStatus updates the status of a crawl
func (s *Server) updateCrawlStatus(crawlID, status, errorMsg string) {
	update := map[string]interface{}{
		"status": status,
	}
	if status == "failed" && errorMsg != "" {
		update["meta"] = map[string]interface{}{
			"error": errorMsg,
		}
	}
	if status == "succeeded" || status == "failed" {
		update["completed_at"] = time.Now().UTC().Format(time.RFC3339)
	}

	_, _, err := s.serviceRole.From("crawls").Update(update, "", "").Eq("id", crawlID).Execute()
	if err != nil {
		s.logger.Error("Failed to update crawl status", zap.String("crawl_id", crawlID), zap.Error(err))
	}
}

// verifyProjectAccess checks if user has access to a project
// Uses service role client to bypass RLS since we've already validated the user's token
func (s *Server) verifyProjectAccess(userID, projectID string) (bool, error) {
	s.logger.Debug("Verifying project access", zap.String("user_id", userID), zap.String("project_id", projectID))
	
	// First check if user is a project member (using service role to bypass RLS)
	var members []map[string]interface{}
	data, _, err := s.serviceRole.From("project_members").
		Select("*", "", false).
		Eq("project_id", projectID).
		Eq("user_id", userID).
		Execute()

	if err != nil {
		s.logger.Error("Failed to query project_members", zap.Error(err))
		return false, err
	}

	// Parse data into members slice
	if err := json.Unmarshal(data, &members); err != nil {
		s.logger.Error("Failed to parse project_members data", zap.Error(err))
		return false, err
	}

	s.logger.Debug("Project members check", zap.Int("member_count", len(members)))

	if len(members) > 0 {
		return true, nil
	}

	// If not a member, check if user is the project owner (using service role to bypass RLS)
	var projects []map[string]interface{}
	projectData, _, err := s.serviceRole.From("projects").
		Select("owner_id", "", false).
		Eq("id", projectID).
		Execute()

	if err != nil {
		s.logger.Error("Failed to query projects", zap.Error(err))
		return false, err
	}

	// Parse data into projects slice
	if err := json.Unmarshal(projectData, &projects); err != nil {
		s.logger.Error("Failed to parse projects data", zap.Error(err))
		return false, err
	}

	s.logger.Debug("Projects check", zap.Int("project_count", len(projects)))

	if len(projects) > 0 {
		ownerID, ok := projects[0]["owner_id"].(string)
		s.logger.Debug("Owner check", zap.String("owner_id", ownerID), zap.Bool("type_ok", ok), zap.Bool("matches", ok && ownerID == userID))
		if ok && ownerID == userID {
			return true, nil
		}
	}

	s.logger.Debug("Access denied", zap.String("user_id", userID), zap.String("project_id", projectID))
	return false, nil
}

// handleGSCConnect handles GET /api/gsc/connect - Get OAuth authorization URL
func (s *Server) handleGSCConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authURL, state, err := gsc.GenerateAuthURL()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate auth URL: %v", err))
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// handleGSCCallback handles GET /api/gsc/callback - OAuth callback
func (s *Server) handleGSCCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !gsc.ValidateState(state) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head><title>GSC Connection Error</title></head>
			<body>
				<h1>Connection Failed</h1>
				<p>Invalid state</p>
				<script>
					window.opener && window.opener.postMessage({type: 'gsc_error', error: 'Invalid state'}, '*');
					setTimeout(() => window.close(), 2000);
				</script>
			</body>
			</html>
		`)
		return
	}

	token, err := gsc.ExchangeCode(code)
	if err != nil {
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

	// Store token (use RemoteAddr as userID for now)
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
}

// handleGSCProperties handles GET /api/gsc/properties - List GSC properties
func (s *Server) handleGSCProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get userID from query or use default
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.RemoteAddr
	}

	properties, err := gsc.GetProperties(userID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get properties: %v", err))
		return
	}

	s.respondJSON(w, http.StatusOK, properties)
}

// handleGSCPerformance handles POST /api/gsc/performance - Fetch performance data
func (s *Server) handleGSCPerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		UserID  string `json:"user_id"`
		SiteURL string `json:"site_url"`
		Days    int    `json:"days"` // Number of days to fetch (default 30)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
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
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch performance data: %v", err))
		return
	}

	s.respondJSON(w, http.StatusOK, performanceMap)
}

// handleGSCEnrichIssues handles POST /api/gsc/enrich-issues - Enrich issues with GSC data
func (s *Server) handleGSCEnrichIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		UserID  string            `json:"user_id"`
		SiteURL string            `json:"site_url"`
		Days    int               `json:"days"`
		Issues  []analyzer.Issue  `json:"issues"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
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
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch performance data: %v", err))
		return
	}

	// Enrich issues
	enrichedIssues := gsc.EnrichIssues(req.Issues, performanceMap)
	s.respondJSON(w, http.StatusOK, enrichedIssues)
}


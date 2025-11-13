package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dillonlara115/barracuda/internal/gsc"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// Config holds API server configuration
type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseAnonKey    string
	CronSyncSecret     string
	Logger             *zap.Logger
}

// Server represents the API server
type Server struct {
	config      Config
	supabase    *supabase.Client
	serviceRole *supabase.Client
	logger      *zap.Logger
	cronSecret  string
}

// NewServer creates a new API server instance
func NewServer(cfg Config) (*Server, error) {
	// Create Supabase client with anon key (for RLS-protected queries)
	supabaseClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	// Create service role client (bypasses RLS for admin operations)
	serviceRoleClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase service role client: %w", err)
	}

	return &Server{
		config:      cfg,
		supabase:    supabaseClient,
		serviceRole: serviceRoleClient,
		logger:      cfg.Logger,
		cronSecret:  cfg.CronSyncSecret,
	}, nil
}

// Router returns the HTTP router with all routes configured
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", s.handleHealth)

	// Initialize GSC OAuth (non-blocking - will fail gracefully if credentials not set)
	// Use the API port for redirect URL
	apiPort := os.Getenv("PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	gscRedirectURL := fmt.Sprintf("http://localhost:%s/api/gsc/callback", apiPort)
	if err := gsc.InitializeOAuth(gscRedirectURL); err != nil {
		s.logger.Warn("GSC integration disabled", zap.Error(err))
		s.logger.Info("Set GSC_CLIENT_ID, GSC_CLIENT_SECRET, or GSC_CREDENTIALS_JSON to enable")
	}

	// Initialize Stripe (non-blocking - will fail gracefully if credentials not set)
	stripeConfig := GetStripeConfig()
	if stripeConfig.SecretKey != "" {
		InitializeStripe(stripeConfig.SecretKey)
		s.logger.Info("Stripe initialized")
	} else {
		s.logger.Warn("Stripe integration disabled - set STRIPE_SECRET_KEY to enable")
	}

	// GSC OAuth callback (OAuth handles its own security)
	mux.HandleFunc("/api/gsc/callback", s.handleGSCCallback)
	// Internal cron endpoint for background sync (protected via shared secret)
	mux.HandleFunc("/api/internal/gsc/sync", s.handleGSCGlobalSync)

	// Stripe webhook (no auth required - verified by signature)
	mux.HandleFunc("/api/stripe/webhook", s.handleStripeWebhook)

	// API v1 routes
	v1 := http.NewServeMux()
	v1.HandleFunc("/crawls", s.handleCrawls)
	v1.HandleFunc("/crawls/", s.handleCrawlByID)
	v1.HandleFunc("/projects", s.handleProjects)
	v1.HandleFunc("/projects/", s.handleProjectByID)
	v1.HandleFunc("/exports", s.handleExports)
	v1.HandleFunc("/billing/checkout", s.handleCreateCheckoutSession)
	v1.HandleFunc("/billing/portal", s.handleCreateBillingPortalSession)

	// Wrap v1 routes with authentication middleware
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", s.authMiddleware(v1)))

	return s.corsMiddleware(s.loggingMiddleware(mux))
}

// authMiddleware validates Supabase JWT tokens
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.respondError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.respondError(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		token := parts[1]

		// Validate token with Supabase
		// Note: Supabase Go client doesn't have built-in JWT validation
		// We'll use the Supabase REST API to verify the token
		user, err := s.validateToken(token)
		if err != nil {
			s.logger.Debug("Token validation failed", zap.Error(err))
			s.respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user info to request context
		ctx := r.Context()
		ctx = contextWithUserID(ctx, user.ID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// validateToken validates a Supabase JWT token and returns user info
func (s *Server) validateToken(token string) (*User, error) {
	// Validate token via Supabase Auth API
	// In production, you might want to verify JWT signature locally for better performance
	return s.validateTokenViaAPI(token)
}

// validateTokenViaAPI validates token by making a request to Supabase Auth API
func (s *Server) validateTokenViaAPI(token string) (*User, error) {
	// Make request to Supabase Auth API to get user
	authURL := s.config.SupabaseURL + "/auth/v1/user"
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", s.config.SupabaseAnonKey)

	client := &http.Client{Timeout: 10 * 1000000000} // 10 seconds
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Debug("Token validation request failed", 
			zap.String("url", authURL),
			zap.Error(err))
		return nil, fmt.Errorf("token validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.logger.Debug("Token validation failed", 
			zap.String("url", authURL),
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(bodyBytes)))
		return nil, fmt.Errorf("token validation failed: status %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		s.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// respondError sends a JSON error response
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// User represents a Supabase user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dillonlara115/barracuda/internal/api"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	apiPort               int
	apiSupabaseURL        string
	apiSupabaseServiceKey string
	apiSupabaseAnonKey    string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the Cloud Run API server",
	Long: `Start the REST API server for Cloud Run deployment.
This server handles crawl ingestion, project management, and provides
authenticated endpoints for the dashboard.`,
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().IntVar(&apiPort, "port", 8080, "Port to run the API server on")
	apiCmd.Flags().StringVar(&apiSupabaseURL, "supabase-url", "", "Supabase project URL (or set PUBLIC_SUPABASE_URL env var)")
	apiCmd.Flags().StringVar(&apiSupabaseServiceKey, "supabase-service-key", "", "Supabase service role key (or set SUPABASE_SERVICE_ROLE_KEY env var)")
	apiCmd.Flags().StringVar(&apiSupabaseAnonKey, "supabase-anon-key", "", "Supabase anon key (or set PUBLIC_SUPABASE_ANON_KEY env var)")

	rootCmd.AddCommand(apiCmd)
}

func runAPI(cmd *cobra.Command, args []string) error {
	// Load .env file if it exists (for production/shared defaults)
	// Then load .env.local to override with local development values
	// Ignore errors - these files are optional
	_ = godotenv.Load()                 // Load .env first
	_ = godotenv.Overload(".env.local") // Override with .env.local if it exists

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Get configuration from flags or environment
	supabaseURL := apiSupabaseURL
	if supabaseURL == "" {
		supabaseURL = os.Getenv("PUBLIC_SUPABASE_URL")
	}
	if supabaseURL == "" {
		return fmt.Errorf("PUBLIC_SUPABASE_URL is required (flag or environment variable)")
	}

	supabaseServiceKey := apiSupabaseServiceKey
	if supabaseServiceKey == "" {
		supabaseServiceKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	}
	if supabaseServiceKey == "" {
		return fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is required (flag or environment variable)")
	}

	supabaseAnonKey := apiSupabaseAnonKey
	if supabaseAnonKey == "" {
		supabaseAnonKey = os.Getenv("PUBLIC_SUPABASE_ANON_KEY")
	}
	if supabaseAnonKey == "" {
		return fmt.Errorf("PUBLIC_SUPABASE_ANON_KEY is required (flag or environment variable)")
	}

	// Check if PORT is set (Cloud Run sets this)
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			apiPort = p
		}
	}

	// Log configuration (without sensitive data)
	logger.Info("Initializing API server",
		zap.String("supabase_url", supabaseURL),
		zap.Bool("has_service_key", supabaseServiceKey != ""),
		zap.Bool("has_anon_key", supabaseAnonKey != ""))

	// Initialize API server
	server, err := api.NewServer(api.Config{
		SupabaseURL:        supabaseURL,
		SupabaseServiceKey: supabaseServiceKey,
		SupabaseAnonKey:    supabaseAnonKey,
		CronSyncSecret:     os.Getenv("GSC_SYNC_SECRET"),
		Logger:             logger,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize API server: %w", err)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", apiPort),
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting API server",
			zap.Int("port", apiPort),
			zap.String("supabase_url", supabaseURL),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Info("Server exited")
	return nil
}

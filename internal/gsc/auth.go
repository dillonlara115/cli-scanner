package gsc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/searchconsole/v1"
	
	"github.com/dillonlara115/barracuda/pkg/models"
)

var (
	// OAuth2 config - will be initialized with credentials
	oauthConfig *oauth2.Config
	// In-memory token storage (in production, use database)
	tokenStore = make(map[string]*oauth2.Token)
	tokenMu    sync.RWMutex
	// State storage for OAuth flow
	stateStore = make(map[string]time.Time)
	stateMu    sync.RWMutex
)

// InitializeOAuth sets up OAuth2 configuration
// Credentials can be provided via environment variables
// Users authorize Barracuda to access their Search Console - no Google Cloud project needed!
func InitializeOAuth(redirectURL string) error {
	// Get credentials from environment variables (required)
	clientID := os.Getenv("GSC_CLIENT_ID")
	clientSecret := os.Getenv("GSC_CLIENT_SECRET")
	
	// If not set, try credentials JSON
	if clientID == "" || clientSecret == "" {
		credentialsJSON := os.Getenv("GSC_CREDENTIALS_JSON")
		if credentialsJSON != "" {
			config, err := google.ConfigFromJSON([]byte(credentialsJSON), searchconsole.WebmastersReadonlyScope)
			if err != nil {
				return fmt.Errorf("failed to parse credentials JSON: %w", err)
			}
			config.RedirectURL = redirectURL
			oauthConfig = config
			return nil
		}
	}
	
	// Final check - if still empty, return error with helpful message
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("GSC OAuth credentials not configured. Set environment variables:\n" +
			"\n" +
			"export GSC_CLIENT_ID='your-client-id'\n" +
			"export GSC_CLIENT_SECRET='your-client-secret'\n" +
			"\n" +
			"Or set GSC_CREDENTIALS_JSON with your full credentials JSON.\n" +
			"\n" +
			"For setup instructions, see: docs/GSC_SETUP_CHECKLIST.md")
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{searchconsole.WebmastersReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	return nil
}

// GenerateAuthURL creates an OAuth2 authorization URL
func GenerateAuthURL() (string, string, error) {
	if oauthConfig == nil {
		return "", "", fmt.Errorf("OAuth not initialized. Call InitializeOAuth first")
	}

	// Generate random state for security
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state with timestamp (expires in 10 minutes)
	stateMu.Lock()
	stateStore[state] = time.Now().Add(10 * time.Minute)
	stateMu.Unlock()

	// Clean up expired states
	go cleanupExpiredStates()

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return url, state, nil
}

// ValidateState checks if OAuth state is valid
func ValidateState(state string) bool {
	stateMu.RLock()
	expires, exists := stateStore[state]
	stateMu.RUnlock()
	
	if !exists {
		return false
	}
	if time.Now().After(expires) {
		stateMu.Lock()
		delete(stateStore, state)
		stateMu.Unlock()
		return false
	}
	stateMu.Lock()
	delete(stateStore, state)
	stateMu.Unlock()
	return true
}

// ExchangeCode exchanges authorization code for token
func ExchangeCode(code string) (*oauth2.Token, error) {
	if oauthConfig == nil {
		return nil, fmt.Errorf("OAuth not initialized")
	}

	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// StoreToken stores token for a user/session
func StoreToken(userID string, token *oauth2.Token) {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	tokenStore[userID] = token
}

// GetToken retrieves token for a user/session
func GetToken(userID string) (*oauth2.Token, bool) {
	tokenMu.RLock()
	token, exists := tokenStore[userID]
	tokenMu.RUnlock()
	
	if !exists {
		return nil, false
	}

	// Check if token needs refresh
	if !token.Valid() {
		// Attempt to refresh
		if token.RefreshToken != "" {
			ctx := context.Background()
			ts := oauthConfig.TokenSource(ctx, token)
			newToken, err := ts.Token()
			if err == nil {
				tokenMu.Lock()
				tokenStore[userID] = newToken
				tokenMu.Unlock()
				return newToken, true
			}
		}
		return nil, false
	}

	return token, true
}

// cleanupExpiredStates removes expired OAuth states
func cleanupExpiredStates() {
	now := time.Now()
	stateMu.Lock()
	defer stateMu.Unlock()
	for state, expires := range stateStore {
		if now.After(expires) {
			delete(stateStore, state)
		}
	}
}

// GetClient creates an authenticated HTTP client
func GetClient(userID string) (*http.Client, error) {
	token, exists := GetToken(userID)
	if !exists {
		return nil, fmt.Errorf("no valid token for user")
	}

	ctx := context.Background()
	client := oauthConfig.Client(ctx, token)
	return client, nil
}

// GetService creates a Search Console service client
func GetService(userID string) (*searchconsole.Service, error) {
	client, err := GetClient(userID)
	if err != nil {
		return nil, err
	}

	service, err := searchconsole.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create search console service: %w", err)
	}

	return service, nil
}

// GetProperties lists all Search Console properties for authenticated user
func GetProperties(userID string) ([]*models.GSCProperty, error) {
	service, err := GetService(userID)
	if err != nil {
		return nil, err
	}

	sites, err := service.Sites.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}

	properties := make([]*models.GSCProperty, 0, len(sites.SiteEntry))
	for _, site := range sites.SiteEntry {
		properties = append(properties, &models.GSCProperty{
			URL:      site.SiteUrl,
			Type:     site.PermissionLevel,
			Verified: site.PermissionLevel != "",
		})
	}

	return properties, nil
}


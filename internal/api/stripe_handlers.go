package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/webhook"
	billingportalsession "github.com/stripe/stripe-go/v78/billingportal/session"
	subscription "github.com/stripe/stripe-go/v78/subscription"
	"go.uber.org/zap"
)

// StripeConfig holds Stripe configuration
type StripeConfig struct {
	SecretKey         string
	WebhookSecret     string
	PriceIDPro        string // Monthly Pro plan
	PriceIDProAnnual  string // Annual Pro plan
	PriceIDTeamSeat   string
	SuccessURL        string
	CancelURL         string
}

// InitializeStripe initializes Stripe with API key
func InitializeStripe(secretKey string) {
	if secretKey == "" {
		return
	}
	stripe.Key = secretKey
}

// GetStripeConfig loads Stripe configuration from environment
func GetStripeConfig() StripeConfig {
	return StripeConfig{
		SecretKey:        os.Getenv("STRIPE_SECRET_KEY"),
		WebhookSecret:    os.Getenv("STRIPE_WEBHOOK_SECRET"),
		PriceIDPro:       os.Getenv("STRIPE_PRICE_ID_PRO"),        // Monthly Pro plan
		PriceIDProAnnual: os.Getenv("STRIPE_PRICE_ID_PRO_ANNUAL"), // Annual Pro plan
		PriceIDTeamSeat:  os.Getenv("STRIPE_PRICE_ID_TEAM_SEAT"),   // Team seat add-on
		SuccessURL:       os.Getenv("STRIPE_SUCCESS_URL"),
		CancelURL:        os.Getenv("STRIPE_CANCEL_URL"),
	}
}

// CreateCheckoutSessionRequest represents a request to create a checkout session
type CreateCheckoutSessionRequest struct {
	PriceID string `json:"price_id"` // Stripe price ID (e.g., "price_xxxxx")
	Quantity int   `json:"quantity,omitempty"` // For team seats, default 1
}

// CreateCheckoutSessionResponse represents the checkout session response
type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// handleCreateCheckoutSession creates a Stripe checkout session
func (s *Server) handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok || userID == "" {
		s.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PriceID == "" {
		s.respondError(w, http.StatusBadRequest, "price_id is required")
		return
	}

	if req.Quantity < 1 {
		req.Quantity = 1
	}

	// Get or create Stripe customer
	stripeConfig := GetStripeConfig()
	if stripeConfig.SecretKey == "" {
		s.respondError(w, http.StatusInternalServerError, "Stripe not configured")
		return
	}

	// Get user profile to check for existing Stripe customer ID
	// Create profile if it doesn't exist
	var profiles []map[string]interface{}
	data, _, err := s.serviceRole.From("profiles").
		Select("stripe_customer_id", "", false).
		Eq("id", userID).
		Execute()
	
	if err != nil {
		s.logger.Error("Failed to get user profile", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}
	
	if err := json.Unmarshal(data, &profiles); err != nil {
		s.logger.Error("Failed to parse user profile", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}
	
	// If profile doesn't exist, create it
	if len(profiles) == 0 {
		// Get user email from Auth API
		user, err := s.validateTokenViaAPI(r.Header.Get("Authorization")[7:]) // Remove "Bearer " prefix
		if err != nil {
			s.logger.Error("Failed to get user email for profile creation", zap.Error(err))
			s.respondError(w, http.StatusInternalServerError, "Failed to get user email")
			return
		}
		
		// Create profile using service role (bypasses RLS)
		_, _, err = s.serviceRole.From("profiles").
			Insert(map[string]interface{}{
				"id": userID,
				"display_name": user.Email,
			}, false, "", "", "").
			Execute()
		
		if err != nil {
			s.logger.Error("Failed to create user profile", zap.Error(err))
			s.respondError(w, http.StatusInternalServerError, "Failed to create user profile")
			return
		}
		
		// Re-fetch the newly created profile
		data, _, err = s.serviceRole.From("profiles").
			Select("stripe_customer_id", "", false).
			Eq("id", userID).
			Execute()
		
		if err != nil {
			s.logger.Error("Failed to fetch newly created profile", zap.Error(err))
			s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
			return
		}
		
		if err := json.Unmarshal(data, &profiles); err != nil || len(profiles) == 0 {
			s.logger.Error("Failed to parse newly created profile", zap.Error(err))
			s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
			return
		}
	}
	
	customerID := ""
	if len(profiles) > 0 {
		if val, ok := profiles[0]["stripe_customer_id"].(string); ok {
			customerID = val
		}
	}

	// Get user email from Auth API (needed for creating Stripe customer if needed)
	user, err := s.validateTokenViaAPI(r.Header.Get("Authorization")[7:]) // Remove "Bearer " prefix
	if err != nil {
		s.logger.Error("Failed to get user email", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to get user email")
		return
	}

	if customerID == "" {
		// Create Stripe customer
		params := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
			Metadata: map[string]string{
				"user_id": userID,
			},
		}
		cust, err := customer.New(params)
		if err != nil {
			s.logger.Error("Failed to create Stripe customer", zap.Error(err))
			s.respondError(w, http.StatusInternalServerError, "Failed to create customer")
			return
		}
		customerID = cust.ID

		// Save customer ID to profile
		_, _, err = s.serviceRole.From("profiles").
			Update(map[string]interface{}{
				"stripe_customer_id": customerID,
			}, "", "").
			Eq("id", userID).
			Execute()
		if err != nil {
			s.logger.Warn("Failed to save Stripe customer ID", zap.Error(err))
		}
	}

	// Create checkout session
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(int64(req.Quantity)),
			},
		},
		SuccessURL: stripe.String(stripeConfig.SuccessURL),
		CancelURL:  stripe.String(stripeConfig.CancelURL),
		Metadata: map[string]string{
			"user_id": userID,
		},
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		s.logger.Error("Failed to create checkout session", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	s.respondJSON(w, http.StatusOK, CreateCheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}

// handleStripeWebhook handles Stripe webhook events
func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Error reading request body", zap.Error(err))
		s.respondError(w, http.StatusServiceUnavailable, "Error reading request body")
		return
	}

	stripeConfig := GetStripeConfig()
	if stripeConfig.WebhookSecret == "" {
		s.respondError(w, http.StatusInternalServerError, "Stripe webhook secret not configured")
		return
	}

	// Verify webhook signature
	// Use ConstructEventWithOptions to handle API version mismatches from Stripe CLI
	event, err := webhook.ConstructEventWithOptions(
		payload,
		r.Header.Get("Stripe-Signature"),
		stripeConfig.WebhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true, // Allow newer API versions from Stripe CLI
		},
	)
	if err != nil {
		s.logger.Error("Webhook signature verification failed", zap.Error(err))
		s.respondError(w, http.StatusBadRequest, "Webhook signature verification failed")
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
			s.logger.Error("Error parsing checkout.session.completed", zap.Error(err))
			s.respondError(w, http.StatusBadRequest, "Error parsing webhook data")
			return
		}
		s.handleCheckoutSessionCompleted(&checkoutSession)

	case "customer.subscription.created", "customer.subscription.updated":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			s.logger.Error("Error parsing subscription event", zap.Error(err))
			s.respondError(w, http.StatusBadRequest, "Error parsing webhook data")
			return
		}
		s.handleSubscriptionUpdate(&subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			s.logger.Error("Error parsing subscription.deleted", zap.Error(err))
			s.respondError(w, http.StatusBadRequest, "Error parsing webhook data")
			return
		}
		s.handleSubscriptionDeleted(&subscription)

	default:
		s.logger.Info("Unhandled event type", zap.String("type", string(event.Type)))
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutSessionCompleted processes a completed checkout session
func (s *Server) handleCheckoutSessionCompleted(session *stripe.CheckoutSession) {
	userID := session.Metadata["user_id"]
	if userID == "" {
		s.logger.Error("Missing user_id in checkout session metadata", zap.String("session_id", session.ID))
		return
	}

	// Check if subscription is available in the session
	if session.Subscription == nil {
		s.logger.Info("Checkout session completed but subscription not yet available - waiting for customer.subscription.created event",
			zap.String("session_id", session.ID),
			zap.String("user_id", userID))
		// The subscription will be created via customer.subscription.created event
		// Just update the profile with the subscription ID when it becomes available
		return
	}

	// Retrieve the subscription
	sub, err := subscription.Get(session.Subscription.ID, nil)
	if err != nil {
		s.logger.Error("Failed to retrieve subscription",
			zap.Error(err),
			zap.String("subscription_id", session.Subscription.ID),
			zap.String("session_id", session.ID))
		return
	}

	s.logger.Info("Processing subscription from checkout session",
		zap.String("subscription_id", sub.ID),
		zap.String("user_id", userID))

	s.handleSubscriptionUpdate(sub)
}

// handleSubscriptionUpdate updates subscription in database
func (s *Server) handleSubscriptionUpdate(sub *stripe.Subscription) {
	// Get customer ID and find user
	customerID := sub.Customer.ID
	
	// Find user by Stripe customer ID
	userID, err := s.getUserIDByStripeCustomerID(customerID)
	if err != nil {
		s.logger.Error("Failed to find user by Stripe customer ID", zap.Error(err))
		return
	}

	// Determine tier based on price ID
	tier := "free"
	stripeConfig := GetStripeConfig()
	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID := sub.Items.Data[0].Price.ID
		if priceID == stripeConfig.PriceIDPro || priceID == stripeConfig.PriceIDProAnnual {
			tier = "pro"
		} else if priceID == stripeConfig.PriceIDTeamSeat {
			tier = "team"
		}
	}

	// Calculate quantity (team size)
	quantity := 1
	if len(sub.Items.Data) > 0 {
		quantity = int(sub.Items.Data[0].Quantity)
	}

	// Get price ID safely
	priceID := ""
	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID = sub.Items.Data[0].Price.ID
	}

	// Insert or update subscription record
	subscriptionData := map[string]interface{}{
		"user_id":                 userID,
		"stripe_subscription_id":  sub.ID,
		"stripe_customer_id":      customerID,
		"stripe_price_id":         priceID,
		"status":                  string(sub.Status),
		"tier":                    tier,
		"quantity":                quantity,
		"current_period_start":    time.Unix(sub.CurrentPeriodStart, 0).Format(time.RFC3339),
		"current_period_end":      time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		"cancel_at_period_end":    sub.CancelAtPeriodEnd,
	}

	if sub.CanceledAt > 0 {
		subscriptionData["canceled_at"] = time.Unix(sub.CanceledAt, 0).Format(time.RFC3339)
	}

	// Check if subscription exists
	var existing []map[string]interface{}
	selectData, _, selectErr := s.serviceRole.From("subscriptions").
		Select("id", "", false).
		Eq("stripe_subscription_id", sub.ID).
		Execute()
	
	if selectErr == nil && selectData != nil {
		if err := json.Unmarshal(selectData, &existing); err == nil && len(existing) > 0 {
			// Update existing subscription
			_, _, err = s.serviceRole.From("subscriptions").
				Update(subscriptionData, "", "").
				Eq("stripe_subscription_id", sub.ID).
				Execute()
		} else {
			// Insert new subscription
			_, _, err = s.serviceRole.From("subscriptions").
				Insert(subscriptionData, false, "", "", "").
				Execute()
		}
	} else {
		// Insert new subscription
		_, _, err = s.serviceRole.From("subscriptions").
			Insert(subscriptionData, false, "", "", "").
			Execute()
	}
	
	if err != nil {
		s.logger.Error("Failed to upsert subscription", zap.Error(err))
		return
	}

	// Update profile with subscription info
	profileUpdate := map[string]interface{}{
		"stripe_subscription_id":      sub.ID,
		"subscription_tier":            tier,
		"subscription_status":          string(sub.Status),
		"subscription_current_period_end": time.Unix(sub.CurrentPeriodEnd, 0).Format(time.RFC3339),
		"subscription_cancel_at_period_end": sub.CancelAtPeriodEnd,
	}

	_, _, err = s.serviceRole.From("profiles").
		Update(profileUpdate, "", "").
		Eq("id", userID).
		Execute()

	if err != nil {
		s.logger.Error("Failed to update profile with subscription info", zap.Error(err))
		// Don't return - subscription was created successfully
	}

	s.logger.Info("Subscription updated", 
		zap.String("user_id", userID),
		zap.String("subscription_id", sub.ID),
		zap.String("tier", tier),
		zap.String("status", string(sub.Status)),
	)
}

// handleSubscriptionDeleted handles subscription cancellation
func (s *Server) handleSubscriptionDeleted(sub *stripe.Subscription) {
	customerID := sub.Customer.ID
	userID, err := s.getUserIDByStripeCustomerID(customerID)
	if err != nil {
		s.logger.Error("Failed to find user by Stripe customer ID", zap.Error(err))
		return
	}

	// Update subscription status to canceled
	_, _, err = s.serviceRole.From("subscriptions").
		Update(map[string]interface{}{
			"status":      "canceled",
			"canceled_at": time.Now().Format(time.RFC3339),
		}, "", "").
		Eq("stripe_subscription_id", sub.ID).
		Execute()

	if err != nil {
		s.logger.Error("Failed to update subscription status", zap.Error(err))
		return
	}

	// Update profile to free tier
	_, _, err = s.serviceRole.From("profiles").
		Update(map[string]interface{}{
			"subscription_tier":   "free",
			"subscription_status": "canceled",
			"team_size":           1,
		}, "", "").
		Eq("id", userID).
		Execute()

	if err != nil {
		s.logger.Error("Failed to update profile", zap.Error(err))
		return
	}

	s.logger.Info("Subscription canceled", zap.String("user_id", userID))
}

// handleCreateBillingPortalSession creates a Stripe billing portal session
func (s *Server) handleCreateBillingPortalSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := userIDFromContext(r.Context())
	if !ok || userID == "" {
		s.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user's Stripe customer ID
	var profiles []map[string]interface{}
	data, _, err := s.serviceRole.From("profiles").
		Select("stripe_customer_id", "", false).
		Eq("id", userID).
		Execute()
	
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}
	
	if err := json.Unmarshal(data, &profiles); err != nil || len(profiles) == 0 {
		s.respondError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}
	
	customerID, ok := profiles[0]["stripe_customer_id"].(string)
	if !ok || customerID == "" {
		s.respondError(w, http.StatusBadRequest, "No active subscription found")
		return
	}

	stripeConfig := GetStripeConfig()
	if stripeConfig.SecretKey == "" {
		s.respondError(w, http.StatusInternalServerError, "Stripe not configured")
		return
	}

	// Create billing portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(stripeConfig.SuccessURL),
	}

	sess, err := billingportalsession.New(params)
	if err != nil {
		s.logger.Error("Failed to create billing portal session", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create billing portal session")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{
		"url": sess.URL,
	})
}

// Helper functions

func (s *Server) getUserIDByStripeCustomerID(customerID string) (string, error) {
	var profiles []map[string]interface{}
	data, _, err := s.serviceRole.From("profiles").
		Select("id", "", false).
		Eq("stripe_customer_id", customerID).
		Execute()
	
	if err != nil {
		return "", fmt.Errorf("failed to query profiles: %w", err)
	}
	
	if err := json.Unmarshal(data, &profiles); err != nil || len(profiles) == 0 {
		return "", fmt.Errorf("user not found for customer ID: %s", customerID)
	}
	
	userID, ok := profiles[0]["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID format")
	}
	return userID, nil
}

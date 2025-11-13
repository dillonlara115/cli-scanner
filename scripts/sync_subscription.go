package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dillonlara115/barracuda/internal/api"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/subscription"
	"go.uber.org/zap"
)

func main() {
	subscriptionID := flag.String("subscription", "", "Stripe subscription ID (e.g., sub_xxx)")
	flag.Parse()

	if *subscriptionID == "" {
		fmt.Println("Usage: go run scripts/sync_subscription.go -subscription sub_xxx")
		os.Exit(1)
	}

	// Initialize Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("STRIPE_SECRET_KEY environment variable not set")
	}
	stripe.Key = stripeKey

	// Retrieve subscription from Stripe
	sub, err := subscription.Get(*subscriptionID, nil)
	if err != nil {
		log.Fatalf("Failed to retrieve subscription: %v", err)
	}

	fmt.Printf("Retrieved subscription: %s\n", sub.ID)
	fmt.Printf("Customer: %s\n", sub.Customer.ID)
	fmt.Printf("Status: %s\n", sub.Status)

	// Initialize API server (minimal config for testing)
	logger, _ := zap.NewDevelopment()
	cfg := api.Config{
		SupabaseURL:        os.Getenv("PUBLIC_SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseAnonKey:    os.Getenv("PUBLIC_SUPABASE_ANON_KEY"),
		Logger:             logger,
	}

	server, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create API server: %v", err)
	}

	// Manually call handleSubscriptionUpdate
	// We need to access the private method, so we'll need to make it public or use reflection
	// For now, let's create a test HTTP request to the webhook endpoint
	fmt.Println("\nTo process this subscription, you can:")
	fmt.Println("1. Make sure your API server is running: go run . api --port 8080")
	fmt.Println("2. Make sure Stripe webhook forwarding is active: stripe listen --forward-to localhost:8080/api/stripe/webhook")
	fmt.Println("3. Update the subscription in Stripe dashboard (change description or metadata)")
	fmt.Println("   This will automatically trigger customer.subscription.updated webhook")
	fmt.Println("\nOr manually trigger via Stripe CLI:")
	fmt.Printf("   stripe trigger customer.subscription.updated\n")
	fmt.Println("\nThe subscription will be processed when the webhook is received.")
}


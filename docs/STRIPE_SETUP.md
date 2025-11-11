# Stripe Payment Integration Setup Guide

This document outlines the Stripe payment integration setup for Barracuda SEO.

## Overview

The Stripe integration supports:
- **Free Plan**: $0/month (100 pages, 1 user)
- **Pro Plan**: $29/month (10,000 pages, 1 user + $5/additional user)
- **Team Plan**: Custom pricing (25,000+ pages, 5+ users)

## Database Migration

Run the migration to add subscription tables:

```bash
# Apply the migration
supabase db push
# Or manually run:
psql $DATABASE_URL < supabase/migrations/20250104_add_subscriptions.sql
```

This creates:
- `subscriptions` table for subscription history
- Adds subscription fields to `profiles` table
- Sets up RLS policies and triggers

## Backend Setup

### 1. Install Stripe Go SDK

```bash
go get github.com/stripe/stripe-go/v78
go mod tidy
```

### 2. Environment Variables

Add these to your `.env` or Cloud Run environment:

```bash
# Stripe API Keys (get from https://dashboard.stripe.com/apikeys)
STRIPE_SECRET_KEY=sk_test_... # or sk_live_... for production
STRIPE_WEBHOOK_SECRET=whsec_... # Get from Stripe Dashboard > Webhooks

# Stripe Price IDs (create products/prices in Stripe Dashboard)
STRIPE_PRICE_ID_PRO=price_xxxxx # Pro plan price ID ($29/month)
STRIPE_PRICE_ID_TEAM_SEAT=price_xxxxx # Team seat add-on ($5/month)

# Redirect URLs after checkout
STRIPE_SUCCESS_URL=https://app.barracudaseo.com/settings?success=true
STRIPE_CANCEL_URL=https://app.barracudaseo.com/settings?canceled=true
```

### 3. Create Products in Stripe Dashboard

1. Go to [Stripe Dashboard > Products](https://dashboard.stripe.com/products)
2. Create two products:

   **Product 1: Pro Plan**
   - Name: "Barracuda Pro"
   - Pricing: Recurring, $29/month
   - Copy the Price ID (starts with `price_`)

   **Product 2: Team Seat Add-on**
   - Name: "Barracuda Team Seat"
   - Pricing: Recurring, $5/month
   - Copy the Price ID (starts with `price_`)

### 4. Set Up Webhook Endpoint

1. Go to [Stripe Dashboard > Webhooks](https://dashboard.stripe.com/webhooks)
2. Click "Add endpoint"
3. Endpoint URL: `https://your-api-domain.com/api/stripe/webhook`
4. Select events to listen to:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
5. Copy the webhook signing secret (starts with `whsec_`)

## API Endpoints

### Create Checkout Session
```
POST /api/v1/billing/checkout
Authorization: Bearer <supabase-jwt-token>
Content-Type: application/json

{
  "price_id": "price_xxxxx",
  "quantity": 1  // Optional, default 1
}
```

Response:
```json
{
  "session_id": "cs_test_...",
  "url": "https://checkout.stripe.com/..."
}
```

### Create Billing Portal Session
```
POST /api/v1/billing/portal
Authorization: Bearer <supabase-jwt-token>
```

Response:
```json
{
  "url": "https://billing.stripe.com/..."
}
```

### Webhook Endpoint
```
POST /api/stripe/webhook
Stripe-Signature: <signature>
```

(No authentication - verified by Stripe signature)

## Frontend Integration

See `web/src/components/Billing.svelte` for the subscription management UI component.

## Testing

### Test Mode

1. Use Stripe test mode API keys (`sk_test_...`)
2. Use test card numbers from [Stripe Testing](https://stripe.com/docs/testing)
   - Success: `4242 4242 4242 4242`
   - Decline: `4000 0000 0000 0002`
3. Use Stripe CLI for local webhook testing:
   ```bash
   stripe listen --forward-to localhost:8080/api/stripe/webhook
   ```

### Production

1. Switch to live API keys (`sk_live_...`)
2. Update webhook endpoint URL in Stripe Dashboard
3. Update redirect URLs to production domain

## Subscription Flow

1. User clicks "Upgrade to Pro" in UI
2. Frontend calls `/api/v1/billing/checkout` with Pro price ID
3. Backend creates Stripe checkout session
4. User redirected to Stripe checkout
5. After payment, Stripe sends webhook to `/api/stripe/webhook`
6. Backend updates `subscriptions` table and `profiles` table
7. User redirected back to app with success message

## Next Steps

1. ✅ Database migration created
2. ✅ Backend handlers created
3. ✅ API routes added
4. ⏳ Install Stripe SDK (`go get github.com/stripe/stripe-go/v78`)
5. ⏳ Create Stripe products/prices in dashboard
6. ⏳ Set environment variables
7. ⏳ Create frontend billing UI component
8. ⏳ Test checkout flow
9. ⏳ Deploy to production







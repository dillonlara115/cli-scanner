# Stripe Payment Integration - Implementation Summary

## ✅ Completed

### 1. Database Schema
- ✅ Created migration `20250104_add_subscriptions.sql`
- ✅ Added subscription fields to `profiles` table
- ✅ Created `subscriptions` table for subscription history
- ✅ Added RLS policies and triggers

### 2. Backend Implementation
- ✅ Created `internal/api/stripe_handlers.go` with:
  - Checkout session creation endpoint
  - Webhook handler for subscription events
  - Billing portal session creation
  - Subscription sync logic
- ✅ Added routes to `internal/api/server.go`:
  - `POST /api/v1/billing/checkout` - Create checkout session
  - `POST /api/v1/billing/portal` - Create billing portal session
  - `POST /api/stripe/webhook` - Handle Stripe webhooks
- ✅ Stripe initialization in server startup

### 3. Frontend Implementation
- ✅ Created `web/src/components/Billing.svelte` - Subscription management UI
- ✅ Created `web/src/routes/Billing.svelte` - Route wrapper
- ✅ Added billing route to `App.svelte`
- ✅ Added "Billing" link to user dropdown menu in `Auth.svelte`

### 4. Documentation
- ✅ Created `docs/STRIPE_SETUP.md` with setup instructions

## ⏳ Next Steps (Required)

### 1. Install Stripe Go SDK
```bash
cd /home/dillon/Sites/cli-scanner
go get github.com/stripe/stripe-go/v78
go mod tidy
```

### 2. Run Database Migration
```bash
# Apply the migration to your Supabase database
supabase db push
# Or manually:
psql $DATABASE_URL < supabase/migrations/20250104_add_subscriptions.sql
```

### 3. Set Up Stripe Account

1. **Create Stripe Account** (if not already done)
   - Go to https://dashboard.stripe.com/register
   - Complete account setup

2. **Create Products & Prices**
   - Go to Stripe Dashboard > Products
   - Create "Barracuda Pro" product:
     - Name: Barracuda Pro
     - Pricing: Recurring, $29/month
     - Copy the Price ID (starts with `price_`)
   - Create "Barracuda Team Seat" product:
     - Name: Barracuda Team Seat  
     - Pricing: Recurring, $5/month
     - Copy the Price ID

3. **Set Up Webhook**
   - Go to Stripe Dashboard > Webhooks
   - Add endpoint: `https://your-api-domain.com/api/stripe/webhook`
   - Select events:
     - `checkout.session.completed`
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
   - Copy webhook signing secret (starts with `whsec_`)

### 4. Configure Environment Variables

**Backend (.env or Cloud Run):**
```bash
STRIPE_SECRET_KEY=sk_test_... # or sk_live_... for production
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID_PRO=price_xxxxx
STRIPE_PRICE_ID_TEAM_SEAT=price_xxxxx
STRIPE_SUCCESS_URL=https://app.barracudaseo.com/billing?success=true
STRIPE_CANCEL_URL=https://app.barracudaseo.com/billing?canceled=true
```

**Frontend (.env or Vercel):**
```bash
VITE_API_URL=https://your-api-domain.com
VITE_STRIPE_PRICE_ID_PRO=price_xxxxx
```

### 5. Test the Integration

1. **Local Testing:**
   ```bash
   # Start API server
   go run . api --port 8080
   
   # In another terminal, forward Stripe webhooks locally
   stripe listen --forward-to localhost:8080/api/stripe/webhook
   ```

2. **Test Flow:**
   - Log into web app
   - Navigate to Billing page
   - Click "Upgrade to Pro"
   - Use test card: `4242 4242 4242 4242`
   - Complete checkout
   - Verify subscription appears in database

### 6. Deploy to Production

1. Update environment variables in Cloud Run
2. Update webhook endpoint URL in Stripe Dashboard
3. Switch to live Stripe keys (`sk_live_...`)
4. Test with real payment method

## API Endpoints

### Create Checkout Session
```
POST /api/v1/billing/checkout
Authorization: Bearer <supabase-jwt>
Content-Type: application/json

{
  "price_id": "price_xxxxx",
  "quantity": 1
}
```

### Create Billing Portal Session
```
POST /api/v1/billing/portal
Authorization: Bearer <supabase-jwt>
```

### Webhook Endpoint
```
POST /api/stripe/webhook
Stripe-Signature: <signature>
```

## Files Created/Modified

**New Files:**
- `supabase/migrations/20250104_add_subscriptions.sql`
- `internal/api/stripe_handlers.go`
- `web/src/components/Billing.svelte`
- `web/src/routes/Billing.svelte`
- `docs/STRIPE_SETUP.md`

**Modified Files:**
- `internal/api/server.go` - Added Stripe routes
- `web/src/App.svelte` - Added billing route
- `web/src/components/Auth.svelte` - Added billing link

## Notes

- The Stripe SDK needs to be installed before the code will compile
- Webhook endpoint doesn't require authentication (verified by Stripe signature)
- Subscription updates are handled automatically via webhooks
- Users can manage billing through Stripe's hosted billing portal
- Free tier is the default for all new users







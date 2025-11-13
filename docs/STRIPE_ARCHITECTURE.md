# Stripe Integration Architecture

## Overview

Our Stripe integration uses a **hybrid approach**:
- **Go Backend**: Handles all Stripe operations (webhooks, checkout, billing portal)
- **Stripe FDW Wrapper**: Allows querying Stripe data directly from Postgres when needed
- **Local Database**: Stores subscription data for fast queries

## Architecture Diagram

```
┌─────────────────┐
│   Frontend      │
│   (Svelte)      │
└────────┬────────┘
         │
         │ HTTP Requests
         ▼
┌─────────────────┐
│   Go Backend    │
│   (Cloud Run)   │
│                 │
│  - Webhooks     │
│  - Checkout     │
│  - Billing      │
└────────┬────────┘
         │
         ├─────────────────┐
         │                 │
         ▼                 ▼
┌─────────────────┐  ┌──────────────┐
│   Supabase DB   │  │  Stripe API  │
│                 │  │              │
│  - Local tables │  │  - Customers │
│  - FDW queries  │  │  - Subs      │
└─────────────────┘  └──────────────┘
```

## Components

### 1. Go Backend (`internal/api/stripe_handlers.go`)

**Responsibilities**:
- Handle Stripe webhooks (`POST /api/stripe/webhook`)
- Create checkout sessions (`POST /api/v1/billing/checkout`)
- Create billing portal sessions (`POST /api/v1/billing/portal`)
- Sync subscription data to database
- Customer management

**Why Go?**
- All backend logic in one place
- Easy to debug and maintain
- No need for Edge Function deployment
- Consistent with rest of backend

### 2. Stripe FDW Wrapper

**Purpose**: Query Stripe data directly from Postgres SQL

**Use Cases**:
- Verify subscription status directly from Stripe
- Get latest subscription details
- Admin/debugging queries
- Periodic sync operations

**Setup**: See `docs/STRIPE_FDW_SETUP.md`

### 3. Local Database Storage

**Tables**:
- `profiles`: User subscription info (tier, status, etc.)
- `subscriptions`: Subscription history and details

**Benefits**:
- Fast queries (no API calls)
- Historical data
- Works offline
- RLS policies for security

## Data Flow

### Webhook Flow

```
Stripe → Webhook → Go Handler → Update DB → Trigger → Update Profile
```

1. Stripe sends webhook to Go backend
2. Go handler verifies signature and processes event
3. Updates `subscriptions` table
4. Database trigger syncs to `profiles` table

### Checkout Flow

```
Frontend → Go API → Stripe API → Checkout Session → Stripe → Webhook → Go → DB
```

1. User clicks "Subscribe" in frontend
2. Frontend calls Go API `/api/v1/billing/checkout`
3. Go creates Stripe checkout session
4. User completes payment on Stripe
5. Stripe sends webhook to Go backend
6. Go updates database

### Query Flow (Using FDW)

```
Go Backend → SQL Query → FDW Wrapper → Stripe API → Results
```

1. Go backend needs latest Stripe data
2. Executes SQL query against `stripe.*` foreign tables
3. FDW wrapper translates to Stripe API call
4. Returns results to Go backend

## When to Use What

### Use Go Backend (Always):
- ✅ Webhook handling (required)
- ✅ Creating checkout sessions (required)
- ✅ Creating billing portal sessions (required)
- ✅ Customer creation/management
- ✅ Subscription updates from webhooks

### Use FDW Wrapper (Optional):
- ✅ Verify subscription status directly from Stripe
- ✅ Get latest subscription details
- ✅ Admin queries
- ✅ Debugging
- ✅ Periodic sync operations

### Use Local Database (Primary):
- ✅ Fast subscription checks
- ✅ User dashboard queries
- ✅ Billing page data
- ✅ Historical subscription data
- ✅ RLS-protected queries

## Security

1. **Webhook Verification**: Go backend verifies Stripe webhook signatures
2. **API Keys**: Stored securely (Cloud Run secrets, Supabase Vault)
3. **RLS Policies**: Local tables protected by Row Level Security
4. **Service Role**: FDW queries use service role (bypasses RLS)

## Benefits of This Architecture

1. **Single Backend**: All logic in Go, no split between Go and Edge Functions
2. **Flexibility**: Can query Stripe directly when needed via FDW
3. **Performance**: Fast local queries for common operations
4. **Reliability**: Local storage works even if Stripe API is down
5. **Maintainability**: One codebase, one language

## Migration Notes

- **No Edge Functions**: We intentionally avoid Edge Functions for Stripe
- **Go-First**: All Stripe operations go through Go backend
- **FDW as Enhancement**: FDW wrapper is optional enhancement, not replacement
- **Webhooks Required**: FDW doesn't replace webhooks - still need Go webhook handler


# Stripe Integration Plan: Go Backend + FDW Wrapper

## Overview

This document outlines our Stripe integration approach: **Keep all backend logic in Go** while using Supabase's Stripe FDW wrapper to enable querying Stripe data directly from Postgres when needed.

## Architecture Decision

**✅ Keep Stripe Logic in Go Backend**
- All Stripe operations handled by Go (`internal/api/stripe_handlers.go`)
- No Edge Functions for Stripe
- Single backend codebase

**✅ Use Stripe FDW Wrapper**
- Query Stripe data directly from Postgres SQL
- Optional enhancement for admin/debugging queries
- Does NOT replace webhooks or checkout logic

## Current State

### Backend (`internal/api/stripe_handlers.go`)
- ✅ Custom Stripe SDK integration (~735 lines)
- ✅ Webhook handling and verification
- ✅ Checkout session creation
- ✅ Billing portal session creation
- ✅ Subscription sync logic
- ✅ Customer creation and management

### Frontend (`web/src/components/Billing.svelte`)
- ✅ Calls Go API endpoints (`/api/v1/billing/checkout`, `/api/v1/billing/portal`)
- ✅ Manages Stripe price IDs via environment variables

### Database
- ✅ `profiles` table with subscription fields
- ✅ `subscriptions` table for subscription history
- ✅ Triggers to sync subscription data to profiles

## Target State

### Backend (No Changes)
- ✅ Keep existing Go Stripe handlers
- ✅ Continue using Stripe SDK in Go
- ✅ Webhooks, checkout, billing portal all in Go

### Database Enhancement
- ✅ Add Stripe FDW wrapper for optional direct queries
- ✅ Create foreign tables for Stripe objects (customers, subscriptions, prices, products)
- ✅ Enable querying Stripe directly from SQL when needed

### Benefits
1. **Single Backend**: All logic stays in Go
2. **Flexibility**: Can query Stripe directly via FDW when needed
3. **Performance**: Fast local queries + optional Stripe queries
4. **No Edge Functions**: Simpler architecture

## Implementation Steps

### Phase 1: Set Up Stripe FDW Wrapper ✅

1. **Enable Wrappers Extension**
   ```sql
   create extension if not exists wrappers with schema extensions;
   ```

2. **Enable Stripe Wrapper**
   ```sql
   create foreign data wrapper stripe_wrapper
     handler stripe_fdw_handler
     validator stripe_fdw_validator;
   ```

3. **Store Stripe API Key in Vault**
   ```sql
   select vault.create_secret(
     'sk_test_...' or 'sk_live_...',
     'stripe',
     'Stripe API key for FDW Wrapper'
   );
   -- Note the returned key_id
   ```

4. **Create Stripe Server Connection**
   ```sql
   create server stripe_server
     foreign data wrapper stripe_wrapper
     options (
       api_key_id '<key_ID>',
       api_url 'https://api.stripe.com/v1/',
       api_version '2024-06-20'
     );
   ```

5. **Create Foreign Tables**
   - `stripe.customers`
   - `stripe.subscriptions`
   - `stripe.prices`
   - `stripe.products`

See `supabase/migrations/20250113_setup_stripe_fdw.sql` for complete migration.

### Phase 2: Update Go Backend (Optional Enhancements)

**Option A: Add FDW Query Helpers**
- Create helper functions to query Stripe via FDW when needed
- Use for admin/debugging queries
- Keep existing webhook/checkout logic unchanged

**Option B: Keep As-Is**
- FDW wrapper available for ad-hoc queries
- No changes needed to Go code
- Use FDW for admin/debugging only

### Phase 3: Documentation

- ✅ Created `docs/STRIPE_FDW_SETUP.md` - Setup guide
- ✅ Created `docs/STRIPE_ARCHITECTURE.md` - Architecture overview
- ✅ Updated this migration plan

## What We're NOT Doing

❌ **NOT moving to Edge Functions**
- All Stripe logic stays in Go
- No TypeScript Edge Functions for Stripe

❌ **NOT replacing webhooks with FDW**
- FDW is read-only, webhooks still required
- Go webhook handler remains essential

❌ **NOT replacing checkout/portal with FDW**
- FDW can't create checkout sessions
- Go handlers remain for all write operations

## Usage Patterns

### Pattern 1: Normal Operations (Current)
```
Frontend → Go API → Stripe API → Webhook → Go → Database
```
- All operations go through Go backend
- Webhooks update local database
- Fast queries from local tables

### Pattern 2: FDW Queries (New, Optional)
```
Go Backend → SQL Query → FDW → Stripe API → Results
```
- Use for admin/debugging
- Verify subscription status directly
- Get latest data from Stripe

### Pattern 3: Hybrid (Recommended)
```
Normal: Local DB queries (fast)
Admin: FDW queries (latest from Stripe)
Webhooks: Keep updating local DB (real-time)
```

## Next Steps

1. ✅ Clean up Edge Functions (removed)
2. ✅ Create FDW setup migration
3. ✅ Update documentation
4. ⏳ Run migration to set up FDW wrapper
5. ⏳ Test FDW queries
6. ⏳ (Optional) Add Go helpers for FDW queries

## Migration Checklist

- [x] Remove Edge Functions
- [x] Create FDW migration
- [x] Update documentation
- [ ] Run migration locally
- [ ] Test FDW queries
- [ ] Run migration in production
- [ ] (Optional) Add Go FDW helpers

# Stripe FDW Wrapper Setup Guide

## Overview

This guide sets up Supabase's Stripe Foreign Data Wrapper (FDW) to enable querying Stripe data directly from Postgres, while keeping all backend logic in Go.

## ⚠️ Important: Dashboard Setup Required

**The Stripe FDW wrapper must be enabled through the Supabase Dashboard first!**

The Stripe wrapper is not available in local Supabase development by default. You need to:

1. **Enable in Supabase Dashboard** (Production/Hosted):
   - Go to: **Supabase Dashboard → Database → Extensions → Wrappers**
   - Find "Stripe" and click **Enable**
   - This creates the `stripe_fdw_handler` and `stripe_fdw_validator` functions

2. **Then run the migration** to create the wrapper and foreign tables

## Architecture

- **Go Backend**: Handles webhooks, checkout sessions, billing portal (existing code)
- **Stripe FDW Wrapper**: Allows querying Stripe data directly from SQL
- **Database**: Stores subscription data locally for fast queries + can query Stripe directly when needed

## Benefits

1. **Keep Go Backend**: All Stripe logic stays in Go (`internal/api/stripe_handlers.go`)
2. **Query Stripe from SQL**: Use FDW wrapper to query Stripe data directly from Postgres
3. **Best of Both Worlds**: Fast local queries + ability to fetch latest data from Stripe
4. **No Edge Functions**: Everything stays in Go backend

## Setup Steps

### Step 1: Enable Stripe Wrapper in Dashboard

**⚠️ REQUIRED FIRST STEP**

1. Go to your Supabase project dashboard
2. Navigate to **Database → Extensions → Wrappers**
3. Find "Stripe" in the list
4. Click **Enable** or **Install**
5. Wait for it to be enabled (creates handler/validator functions)

### Step 2: Enable Wrappers Extension

```sql
-- Run this migration
create extension if not exists wrappers with schema extensions;
```

### Step 3: Create Stripe Wrapper

After enabling in the dashboard, create the wrapper:

```sql
-- Check if handlers exist first
do $$
begin
  if not exists (
    select 1 from pg_foreign_data_wrapper where fdwname = 'stripe_wrapper'
  ) then
    -- Verify handlers exist (they should after enabling in dashboard)
    if exists (
      select 1 from pg_proc where proname = 'stripe_fdw_handler'
    ) then
      create foreign data wrapper stripe_wrapper
        handler stripe_fdw_handler
        validator stripe_fdw_validator;
    else
      raise exception 'Stripe FDW handlers not found. Please enable Stripe wrapper in Supabase Dashboard first.';
    end if;
  end if;
end $$;
```

### Step 4: Store Stripe API Key in Vault

```sql
-- Store your Stripe secret key securely in Vault
-- Replace <YOUR_STRIPE_SECRET_KEY> with your actual key
select vault.create_secret(
  '<YOUR_STRIPE_SECRET_KEY>',
  'stripe',
  'Stripe API key for FDW Wrapper'
);
```

**Note**: This returns a `key_id` that you'll use in the next step.

### Step 5: Create Stripe Server Connection

```sql
-- Replace <key_ID> with the key_id from Step 4
do $$
declare
  key_id text := '<key_ID>';  -- Replace with actual key_id from vault
begin
  if not exists (
    select 1 from pg_foreign_server where srvname = 'stripe_server'
  ) then
    execute format('
      create server stripe_server
        foreign data wrapper stripe_wrapper
        options (
          api_key_id %L,
          api_url ''https://api.stripe.com/v1/'',
          api_version ''2024-06-20''
        )
    ', key_id);
  end if;
end $$;
```

### Step 6: Create Schema for Stripe Tables

```sql
create schema if not exists stripe;
```

### Step 7: Create Foreign Tables

Create foreign tables for the Stripe objects you want to query. See `supabase/migrations/20250113_setup_stripe_fdw.sql` for complete table definitions.

## Local Development Limitation

**⚠️ Note**: The Stripe FDW wrapper is typically only available in hosted Supabase instances, not in local development (`supabase start`). 

For local development:
- Keep using your Go backend Stripe handlers (they work fine locally)
- Set up FDW wrapper in your production/hosted Supabase instance
- Use FDW queries in production for admin/debugging

## Usage Examples

### Query Stripe Customers

```sql
-- Get all customers
select id, email, name, created 
from stripe.customers 
limit 10;

-- Get specific customer
select * from stripe.customers where id = 'cus_xxx';
```

### Query Subscriptions

```sql
-- Get all active subscriptions
select 
  id,
  customer,
  status,
  current_period_start,
  current_period_end,
  attrs->>'status' as detailed_status
from stripe.subscriptions
where status = 'active';

-- Get subscription for specific customer
select * from stripe.subscriptions where customer = 'cus_xxx';
```

### Query Prices

```sql
-- Get all prices
select id, product, currency, unit_amount, active
from stripe.prices
where active = true;

-- Get specific price
select * from stripe.prices where id = 'price_xxx';
```

### Join Stripe Data with Local Data

```sql
-- Get subscription details from Stripe and join with local profiles
select 
  p.id as user_id,
  p.email,
  s.id as stripe_subscription_id,
  s.status,
  s.current_period_end,
  s.attrs->>'items' as subscription_items
from public.profiles p
join stripe.subscriptions s on s.customer = p.stripe_customer_id
where p.stripe_customer_id is not null;
```

## Using from Go Backend

You can query these foreign tables directly from your Go code:

```go
// Example: Query Stripe subscription directly
rows, err := db.Query(`
  SELECT id, customer, status, current_period_end, attrs
  FROM stripe.subscriptions
  WHERE id = $1
`, subscriptionID)

// Example: Get latest subscription status from Stripe
var status string
err := db.QueryRow(`
  SELECT status
  FROM stripe.subscriptions
  WHERE customer = $1
  ORDER BY created DESC
  LIMIT 1
`, customerID).Scan(&status)
```

## When to Use FDW vs Local Storage

### Use Local Storage (`subscriptions` table):
- Fast queries for subscription checks
- Historical data
- Frequent reads (user dashboard, billing checks)
- **Works in local development**

### Use FDW Wrapper:
- Verify subscription status directly from Stripe
- Get latest subscription details
- Ad-hoc queries for admin/debugging
- Sync operations (periodic refresh from Stripe)
- **Only available in hosted Supabase**

## Security Notes

1. **Vault Storage**: API keys are stored securely in Supabase Vault
2. **RLS Policies**: Apply RLS policies to foreign tables if needed
3. **Service Role**: FDW queries typically use service role key (bypasses RLS)

## Limitations

- **Read Operations**: FDW wrapper is primarily for reading Stripe data
- **Performance**: Each query makes an API call to Stripe (slower than local queries)
- **Rate Limits**: Be mindful of Stripe API rate limits
- **Webhooks Still Required**: FDW doesn't replace webhooks - you still need Go webhook handler
- **Local Development**: Not available in local Supabase (`supabase start`), only in hosted instances

## Troubleshooting

### Error: "function stripe_fdw_handler() does not exist"

**Solution**: Enable the Stripe wrapper in Supabase Dashboard first:
1. Go to Dashboard → Database → Extensions → Wrappers
2. Enable "Stripe"
3. Then run the migration

### Error: "server stripe_server does not exist"

**Solution**: Create the server connection first (Step 5) before creating foreign tables.

### Foreign tables not working locally

**Solution**: Stripe FDW wrapper is only available in hosted Supabase instances, not local development. Use your Go backend handlers for local development.

## Next Steps

1. ✅ Enable Stripe wrapper in Supabase Dashboard
2. ✅ Run migration to set up FDW wrapper
3. ✅ Create foreign tables for objects you need
4. ✅ Update Go code to optionally use FDW queries when needed
5. ✅ Keep existing webhook handler in Go (still needed for real-time updates)

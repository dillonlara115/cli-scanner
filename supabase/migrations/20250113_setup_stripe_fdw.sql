-- Setup Stripe Foreign Data Wrapper (FDW)
-- This allows querying Stripe data directly from Postgres
-- while keeping all backend logic in Go

-- Step 1: Enable wrappers extension
create extension if not exists wrappers with schema extensions;

-- Step 2: Enable Stripe wrapper
-- IMPORTANT: The Stripe FDW wrapper must be enabled through the Supabase Dashboard first!
-- Go to: Supabase Dashboard → Database → Extensions → Wrappers → Enable Stripe
-- 
-- After enabling in the dashboard, the handler/validator functions will be available.
-- Then you can create the wrapper:
--
-- Note: PostgreSQL doesn't support IF NOT EXISTS for CREATE FOREIGN DATA WRAPPER
-- Wrap in DO block to handle case where it already exists
-- 
-- do $$
-- begin
--   if not exists (
--     select 1 from pg_foreign_data_wrapper where fdwname = 'stripe_wrapper'
--   ) then
--     -- Check if handlers exist (they should after enabling in dashboard)
--     if exists (
--       select 1 from pg_proc where proname = 'stripe_fdw_handler'
--     ) then
--       create foreign data wrapper stripe_wrapper
--         handler stripe_fdw_handler
--         validator stripe_fdw_validator;
--     else
--       raise notice 'Stripe FDW handlers not found. Please enable Stripe wrapper in Supabase Dashboard first.';
--     end if;
--   end if;
-- end $$;

-- Step 3: Store Stripe API key in Vault
-- NOTE: You need to run this manually with your actual Stripe secret key:
-- select vault.create_secret(
--   'sk_test_...' or 'sk_live_...',
--   'stripe',
--   'Stripe API key for FDW Wrapper'
-- );
-- This will return a key_id that you'll use in the next step

-- Step 4: Create Stripe server connection
-- NOTE: Replace <key_ID> with the key_id from vault.create_secret above
-- You can find it by running: select * from vault.secrets where name = 'stripe';
-- 
-- Note: PostgreSQL doesn't support IF NOT EXISTS for CREATE SERVER
-- Wrap in DO block to handle case where it already exists:
-- 
-- do $$
-- declare
--   key_id text := '<key_ID>';  -- Replace with actual key_id from vault
-- begin
--   if not exists (
--     select 1 from pg_foreign_server where srvname = 'stripe_server'
--   ) then
--     execute format('
--       create server stripe_server
--         foreign data wrapper stripe_wrapper
--         options (
--           api_key_id %L,
--           api_url ''https://api.stripe.com/v1/'',
--           api_version ''2024-06-20''
--         )
--     ', key_id);
--   end if;
-- end $$;

-- Step 5: Create schema for Stripe foreign tables
create schema if not exists stripe;

-- Step 6: Create foreign tables for Stripe objects
-- NOTE: These will only be created if stripe_server exists
-- Run Step 4 first to create the server, then these tables can be created

-- Customers table
-- Note: Wrap in DO block since IF NOT EXISTS not supported for foreign tables
-- Also check that stripe_server exists before creating
do $$
begin
  if exists (
    select 1 from pg_foreign_server where srvname = 'stripe_server'
  ) and not exists (
    select 1 from information_schema.tables 
    where table_schema = 'stripe' and table_name = 'customers'
  ) then
    create foreign table stripe.customers (
  id text,
  email text,
  name text,
  description text,
  created timestamp,
  attrs jsonb
)
    server stripe_server
    options (
      object 'customers',
      rowid_column 'id'
    );
  end if;
end $$;

-- Subscriptions table
do $$
begin
  if exists (
    select 1 from pg_foreign_server where srvname = 'stripe_server'
  ) and not exists (
    select 1 from information_schema.tables 
    where table_schema = 'stripe' and table_name = 'subscriptions'
  ) then
    create foreign table stripe.subscriptions (
  id text,
  customer text,
  currency text,
  current_period_start timestamp,
  current_period_end timestamp,
  status text,
  attrs jsonb
)
    server stripe_server
    options (
      object 'subscriptions',
      rowid_column 'id'
    );
  end if;
end $$;

-- Prices table
do $$
begin
  if exists (
    select 1 from pg_foreign_server where srvname = 'stripe_server'
  ) and not exists (
    select 1 from information_schema.tables 
    where table_schema = 'stripe' and table_name = 'prices'
  ) then
    create foreign table stripe.prices (
  id text,
  product text,
  currency text,
  unit_amount bigint,
  recurring jsonb,
  active boolean,
  attrs jsonb
)
    server stripe_server
    options (
      object 'prices'
    );
  end if;
end $$;

-- Products table
do $$
begin
  if exists (
    select 1 from pg_foreign_server where srvname = 'stripe_server'
  ) and not exists (
    select 1 from information_schema.tables 
    where table_schema = 'stripe' and table_name = 'products'
  ) then
    create foreign table stripe.products (
  id text,
  name text,
  description text,
  active boolean,
  created timestamp,
  attrs jsonb
)
    server stripe_server
    options (
      object 'products',
      rowid_column 'id'
    );
  end if;
end $$;

-- IMPORTANT SETUP NOTES:
-- 
-- 1. FIRST: Enable Stripe wrapper in Supabase Dashboard
--    - Go to: Dashboard → Database → Extensions → Wrappers → Enable Stripe
--    - This creates the stripe_fdw_handler and stripe_fdw_validator functions
--
-- 2. Run vault.create_secret() manually with your Stripe secret key:
--    select vault.create_secret('sk_test_...', 'stripe', 'Stripe API key');
--    Note the returned key_id
--
-- 3. Uncomment and update Step 2 (create wrapper) and Step 4 (create server) above
--    Replace <key_ID> with the key_id from vault.create_secret
--
-- 4. Run this migration again to create the wrapper and server
--
-- 5. The foreign tables will then be able to query Stripe
--
-- NOTE: Stripe FDW wrapper is typically only available in hosted Supabase instances,
--       not in local development (supabase start). For local dev, use Go backend handlers.


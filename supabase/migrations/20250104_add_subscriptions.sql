-- Add subscription fields to profiles table
-- Reference: marketing/src/TEAM_ACCOUNTS.md - Pricing Model

-- Add subscription columns to profiles
alter table public.profiles
add column if not exists subscription_tier text check (subscription_tier in ('free', 'pro', 'team')) default 'free',
add column if not exists stripe_customer_id text,
add column if not exists stripe_subscription_id text,
add column if not exists subscription_status text check (subscription_status in ('active', 'canceled', 'past_due', 'trialing', 'incomplete', 'incomplete_expired')) default 'active',
add column if not exists team_size integer default 1,
add column if not exists subscription_current_period_end timestamptz,
add column if not exists subscription_cancel_at_period_end boolean default false;

-- Create subscriptions table for detailed subscription history
create table if not exists public.subscriptions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users (id) on delete cascade,
  stripe_subscription_id text not null unique,
  stripe_customer_id text not null,
  stripe_price_id text not null,
  status text check (status in ('active', 'canceled', 'past_due', 'trialing', 'incomplete', 'incomplete_expired')) not null,
  tier text check (tier in ('free', 'pro', 'team')) not null,
  quantity integer default 1, -- Number of seats for team plans
  current_period_start timestamptz not null,
  current_period_end timestamptz not null,
  cancel_at_period_end boolean default false,
  canceled_at timestamptz,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

-- Indexes for subscriptions
create index if not exists idx_subscriptions_user_id on public.subscriptions (user_id);
create index if not exists idx_subscriptions_stripe_subscription_id on public.subscriptions (stripe_subscription_id);
create index if not exists idx_subscriptions_status on public.subscriptions (status);

-- Index for profiles subscription lookup
create index if not exists idx_profiles_stripe_customer_id on public.profiles (stripe_customer_id);
create index if not exists idx_profiles_subscription_tier on public.profiles (subscription_tier);

-- RLS Policies for subscriptions
alter table public.subscriptions enable row level security;

create policy "Users can view their own subscriptions"
  on public.subscriptions
  for select
  using (auth.uid() = user_id);

-- Service role can manage all subscriptions (for webhook updates)
-- This is handled via service_role key, no policy needed

-- Function to sync subscription to profile
create or replace function public.sync_subscription_to_profile()
returns trigger
language plpgsql
security definer
as $$
begin
  -- Update profile with latest subscription info
  update public.profiles
  set
    subscription_tier = new.tier,
    stripe_customer_id = new.stripe_customer_id,
    stripe_subscription_id = new.stripe_subscription_id,
    subscription_status = new.status,
    team_size = new.quantity,
    subscription_current_period_end = new.current_period_end,
    subscription_cancel_at_period_end = new.cancel_at_period_end,
    updated_at = now()
  where id = new.user_id;
  
  return new;
end;
$$;

-- Trigger to sync subscription changes to profile
create trigger sync_subscription_to_profile_trigger
  after insert or update on public.subscriptions
  for each row
  execute function public.sync_subscription_to_profile();

-- Trigger for updated_at
create trigger set_updated_at_subscriptions
  before update on public.subscriptions
  for each row
  execute function public.handle_updated_at();











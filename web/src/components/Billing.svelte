<script>
  import { onMount } from 'svelte';
  import { user } from '../lib/auth.js';
  import { supabase } from '../lib/supabase.js';
  import { CreditCard, Check, X, Loader } from 'lucide-svelte';

  let loading = true;
  let profile = null;
  let subscription = null;
  let error = null;
  let creatingCheckout = false;
  let creatingPortal = false;
  
  const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  const STRIPE_PRICE_ID_PRO = import.meta.env.VITE_STRIPE_PRICE_ID_PRO || '';
  const STRIPE_PRICE_ID_PRO_ANNUAL = import.meta.env.VITE_STRIPE_PRICE_ID_PRO_ANNUAL || '';
  const STRIPE_PRICE_ID_TEAM_SEAT = import.meta.env.VITE_STRIPE_PRICE_ID_TEAM_SEAT || '';
  
  let selectedBillingPeriod = 'monthly'; // 'monthly' or 'annual'

  onMount(async () => {
    await loadSubscriptionData();
  });

  async function loadSubscriptionData() {
    if (!$user) return;
    
    loading = true;
    error = null;
    
    try {
      // Ensure we have a valid session before querying (RLS requires auth.uid())
      const { data: sessionData, error: sessionError } = await supabase.auth.getSession();
      if (sessionError || !sessionData.session) {
        console.warn('No active session, refreshing...');
        const { error: refreshError } = await supabase.auth.refreshSession();
        if (refreshError) {
          throw new Error('Session expired. Please sign in again.');
        }
      }
      
      // Load profile with subscription info
      // Note: Profile will be created automatically by the backend API if it doesn't exist
      // Use limit(1) so zero-row responses stay as arrays and avoid PGRST116 errors
      const { data: profileRows, error: profileError } = await supabase
        .from('profiles')
        .select('*')
        .eq('id', $user.id)
        .limit(1); // Avoid single-row coercion errors when no profile exists
      
      let profileData = profileRows?.[0] || null;
      
      // Handle profile error - PGRST116 means "no rows" which is OK
      if (profileError) {
        const profileErrCode = profileError.code;
        const profileErrMessage = profileError.message || JSON.stringify(profileError);
        const isProfilePGRST116 = profileErrCode === 'PGRST116' || 
                                  profileErrMessage?.includes('PGRST116') || 
                                  profileErrMessage?.includes('contains 0 rows');
        
        if (isProfilePGRST116) {
          // No profile found - use defaults (this is OK)
          profileData = null;
        } else {
          // Real error - throw it
          throw profileError;
        }
      }
      
      // If profile doesn't exist, use defaults (backend will create it when needed)
      profile = profileData || {
        id: $user.id,
        subscription_tier: 'free',
        subscription_status: 'active',
        team_size: 1
      };

      // Load subscription - try multiple approaches
      // First, try by user_id (most reliable since RLS policy checks user_id)
      if (profile.id) {
        // Query by user_id first - this matches the RLS policy
        const { data: userSubRows, error: userSubError } = await supabase
          .from('subscriptions')
          .select('*')
          .eq('user_id', profile.id)
          .limit(1);
        
        const userSubData = userSubRows?.[0] || null;
        
        if (userSubError) {
          if (userSubError.code === 'PGRST116') {
            // No subscription found - this is OK
            console.log('No subscription found for user_id:', profile.id);
          } else {
            console.warn('Subscription fetch by user_id error:', userSubError);
          }
        } else if (userSubData) {
          subscription = userSubData;
          // Update profile with subscription ID if it's missing
          if (!profile.stripe_subscription_id && userSubData.stripe_subscription_id) {
            profile.stripe_subscription_id = userSubData.stripe_subscription_id;
          }
        }
      }
      
      // Fallback: try by stripe_subscription_id if we have it and didn't find by user_id
      if (!subscription && profile.stripe_subscription_id) {
        const { data: subRows, error: subError } = await supabase
          .from('subscriptions')
          .select('*')
          .eq('stripe_subscription_id', profile.stripe_subscription_id)
          .limit(1);
        
        const subData = subRows?.[0] || null;
        
        if (subError) {
          if (subError.code !== 'PGRST116') {
            console.warn('Subscription fetch by stripe_subscription_id error:', subError);
          }
        } else if (subData) {
          subscription = subData;
        }
      }
    } catch (err) {
      // Only show error if it's not a "no rows" error (PGRST116)
      // Check multiple possible error structures from Supabase
      const errMessage = err.message || err.error?.message || JSON.stringify(err);
      const errCode = err.code || err.error?.code;
      const isPGRST116 = errCode === 'PGRST116' || errMessage?.includes('PGRST116') || errMessage?.includes('contains 0 rows');
      
      if (!isPGRST116) {
        // Real error - show it
        error = errMessage;
        console.error('Failed to load subscription data:', err);
      } else {
        // PGRST116 is expected when no subscription exists yet - not an error
        // Clear any previous error and just log
        error = null;
        console.log('No subscription data found yet (this is OK - webhook may still be processing)');
      }
    } finally {
      loading = false;
    }
  }

  async function createCheckoutSession(priceId) {
    if (!$user) return;
    
    creatingCheckout = true;
    error = null;
    
    try {
      // Refresh session to ensure we have a valid token
      const { data: sessionData, error: sessionError } = await supabase.auth.getSession();
      if (sessionError || !sessionData.session) {
        throw new Error('Not authenticated. Please sign in again.');
      }
      
      // Check if token is expired or about to expire (within 60 seconds)
      const expiresAt = sessionData.session.expires_at;
      if (expiresAt && expiresAt * 1000 < Date.now() + 60000) {
        // Refresh the session
        const { error: refreshError } = await supabase.auth.refreshSession();
        if (refreshError) {
          throw new Error('Session expired. Please sign in again.');
        }
      }
      
      const token = (await supabase.auth.getSession()).data.session?.access_token;
      if (!token) {
        throw new Error('Not authenticated');
      }

      const response = await fetch(`${API_URL}/api/v1/billing/checkout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          price_id: priceId,
          quantity: 1,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create checkout session');
      }

      const data = await response.json();
      
      // Redirect to Stripe checkout
      window.location.href = data.url;
    } catch (err) {
      error = err.message;
      console.error('Failed to create checkout session:', err);
    } finally {
      creatingCheckout = false;
    }
  }

  async function openBillingPortal() {
    if (!$user) return;
    
    creatingPortal = true;
    error = null;
    
    try {
      const token = (await supabase.auth.getSession()).data.session?.access_token;
      if (!token) {
        throw new Error('Not authenticated');
      }

      const response = await fetch(`${API_URL}/api/v1/billing/portal`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create billing portal session');
      }

      const data = await response.json();
      
      // Open billing portal in new window
      window.location.href = data.url;
    } catch (err) {
      error = err.message;
      console.error('Failed to open billing portal:', err);
    } finally {
      creatingPortal = false;
    }
  }

  function getPlanFeatures(tier) {
    switch (tier) {
      case 'pro':
        return {
          pages: '10,000+',
          users: profile?.team_size || 1,
          integrations: true,
          recommendations: true,
        };
      case 'team':
        return {
          pages: '25,000+',
          users: profile?.team_size || 5,
          integrations: true,
          recommendations: true,
        };
      default:
        return {
          pages: '100',
          users: 1,
          integrations: false,
          recommendations: false,
        };
    }
  }

  function formatDate(dateString) {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  $: planFeatures = getPlanFeatures(profile?.subscription_tier || 'free');
  $: isProOrTeam = profile?.subscription_tier === 'pro' || profile?.subscription_tier === 'team';
</script>

<div class="container mx-auto p-6 max-w-4xl">
  <div class="mb-6">
    <h1 class="text-3xl font-bold mb-2">Billing & Subscription</h1>
    <p class="text-base-content/70">
      Manage your subscription and billing information.
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center min-h-[400px]">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else if error}
    <div class="alert alert-error mb-6">
      <X class="w-5 h-5" />
      <span>{error}</span>
    </div>
  {:else if profile}
    <div class="space-y-6">
      <!-- Current Plan Card -->
      <div class="card bg-base-100 shadow-lg">
        <div class="card-body">
          <h2 class="card-title text-xl mb-4">Current Plan</h2>
          
          <div class="flex items-center justify-between mb-4">
            <div>
              <div class="badge badge-lg badge-primary badge-outline uppercase">
                {profile.subscription_tier || 'free'}
              </div>
              {#if subscription}
                <p class="text-sm text-base-content/70 mt-2">
                  Status: <span class="badge badge-sm badge-success">{subscription.status}</span>
                </p>
              {/if}
            </div>
            
            {#if isProOrTeam}
              <button 
                class="btn btn-primary"
                on:click={openBillingPortal}
                disabled={creatingPortal}
              >
                {#if creatingPortal}
                  <Loader class="w-4 h-4 animate-spin" />
                {:else}
                  <CreditCard class="w-4 h-4" />
                {/if}
                Manage Billing
              </button>
            {/if}
          </div>

          <div class="grid grid-cols-2 gap-4 mt-4">
            <div>
              <p class="text-sm text-base-content/70">Crawl Limit</p>
              <p class="text-lg font-semibold">{planFeatures.pages} pages</p>
            </div>
            <div>
              <p class="text-sm text-base-content/70">Team Members</p>
              <p class="text-lg font-semibold">{planFeatures.users}</p>
            </div>
          </div>

          {#if subscription}
            <div class="divider my-4"></div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-sm text-base-content/70">Current Period</p>
                <p class="text-sm">
                  {formatDate(subscription.current_period_start)} - {formatDate(subscription.current_period_end)}
                </p>
              </div>
              {#if subscription.cancel_at_period_end}
                <div>
                  <p class="text-sm text-warning">Cancels on</p>
                  <p class="text-sm">{formatDate(subscription.current_period_end)}</p>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </div>

      <!-- Plan Features -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title text-xl mb-4">Plan Features</h2>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              {#if planFeatures.integrations}
                <Check class="w-5 h-5 text-success" />
              {:else}
                <X class="w-5 h-5 text-base-content/30" />
              {/if}
              <span>Google Search Console & Analytics integrations</span>
            </div>
            <div class="flex items-center gap-2">
              {#if planFeatures.recommendations}
                <Check class="w-5 h-5 text-success" />
              {:else}
                <X class="w-5 h-5 text-base-content/30" />
              {/if}
              <span>AI-powered recommendations</span>
            </div>
            <div class="flex items-center gap-2">
              {#if isProOrTeam}
                <Check class="w-5 h-5 text-success" />
              {:else}
                <X class="w-5 h-5 text-base-content/30" />
              {/if}
              <span>Team collaboration</span>
            </div>
            <div class="flex items-center gap-2">
              {#if isProOrTeam}
                <Check class="w-5 h-5 text-success" />
              {:else}
                <X class="w-5 h-5 text-base-content/30" />
              {/if}
              <span>Priority support</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Upgrade Options -->
      {#if !isProOrTeam}
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title text-xl mb-4">Upgrade Plan</h2>
            <p class="text-base-content/70 mb-4">
              Unlock more features with a Pro subscription.
            </p>
            
            <!-- Billing Period Toggle -->
            <div class="flex justify-center mb-6">
              <div class="btn-group">
                <button 
                  class="btn btn-sm {selectedBillingPeriod === 'monthly' ? 'btn-primary' : 'btn-outline'}"
                  on:click={() => selectedBillingPeriod = 'monthly'}
                >
                  Monthly
                </button>
                <button 
                  class="btn btn-sm {selectedBillingPeriod === 'annual' ? 'btn-primary' : 'btn-outline'}"
                  on:click={() => selectedBillingPeriod = 'annual'}
                >
                  Annual
                  <span class="badge badge-success badge-sm ml-2">Save 20%</span>
                </button>
              </div>
            </div>
            
            <div class="bg-primary/10 rounded-lg p-4 mb-4">
              {#if selectedBillingPeriod === 'monthly'}
                <h3 class="font-semibold mb-2">Pro Plan - $29/month</h3>
              {:else}
                <h3 class="font-semibold mb-2">Pro Plan - Annual</h3>
                <p class="text-sm text-base-content/70 mb-2">Billed annually, save 20%</p>
              {/if}
              <ul class="text-sm space-y-1 mb-4">
                <li>✓ Crawl up to 10,000 pages</li>
                <li>✓ Team collaboration (1 user included, +$5/user)</li>
                <li>✓ All integrations</li>
                <li>✓ AI recommendations</li>
                <li>✓ Priority support</li>
              </ul>
            </div>

            <button 
              class="btn btn-primary w-full"
              on:click={() => {
                const priceId = selectedBillingPeriod === 'monthly' 
                  ? STRIPE_PRICE_ID_PRO 
                  : STRIPE_PRICE_ID_PRO_ANNUAL;
                createCheckoutSession(priceId);
              }}
              disabled={creatingCheckout || (!STRIPE_PRICE_ID_PRO && !STRIPE_PRICE_ID_PRO_ANNUAL)}
            >
              {#if creatingCheckout}
                <Loader class="w-4 h-4 animate-spin" />
                Processing...
              {:else}
                Upgrade to Pro {selectedBillingPeriod === 'annual' ? '(Annual)' : ''}
              {/if}
            </button>

            {#if !STRIPE_PRICE_ID_PRO && !STRIPE_PRICE_ID_PRO_ANNUAL}
              <p class="text-sm text-warning mt-2">
                Stripe is not configured. Please set VITE_STRIPE_PRICE_ID_PRO and VITE_STRIPE_PRICE_ID_PRO_ANNUAL environment variables.
              </p>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  :global(.badge-success) {
    background-color: #8ec07c;
    color: white;
  }
</style>





<script>
  import { onMount } from 'svelte';
  import Router, { link, push } from 'svelte-spa-router';
  import { initAuth, user } from './lib/auth.js';
  import { supabase } from './lib/supabase.js';
  import Auth from './components/Auth.svelte';
  import ConfigError from './components/ConfigError.svelte';
  import ProjectsList from './routes/ProjectsList.svelte';
  import ProjectView from './routes/ProjectView.svelte';
  import CrawlView from './routes/CrawlView.svelte';
  import Integrations from './routes/Integrations.svelte';
  import Settings from './routes/Settings.svelte';

  let loading = true;
  let configError = null;

  // Route definitions
  const routes = {
    '/': ProjectsList,
    '/project/:id': ProjectView,
    '/project/:projectId/crawl/:crawlId': CrawlView,
    '/project/:projectId/settings': Settings,
    '/integrations': Integrations,
  };

  // Check Supabase configuration
  $: {
    const supabaseUrl = import.meta.env.PUBLIC_SUPABASE_URL || import.meta.env.VITE_PUBLIC_SUPABASE_URL;
    const supabaseAnonKey = import.meta.env.PUBLIC_SUPABASE_ANON_KEY || import.meta.env.VITE_PUBLIC_SUPABASE_ANON_KEY;
    
    if (!supabaseUrl || !supabaseAnonKey) {
      configError = 'Missing Supabase configuration. Please set PUBLIC_SUPABASE_URL and PUBLIC_SUPABASE_ANON_KEY environment variables.';
    } else {
      configError = null;
    }
  }

  onMount(async () => {
    // Check for auth callback (email confirmation, password reset, etc.)
    const hashParams = new URLSearchParams(window.location.hash.substring(1));
    const accessToken = hashParams.get('access_token');
    
    if (accessToken) {
      // Handle auth callback from email confirmation
      const { data, error } = await supabase.auth.setSession({
        access_token: accessToken,
        refresh_token: hashParams.get('refresh_token') || ''
      });
      
      if (error) {
        console.error('Auth callback error:', error);
      } else {
        // Clear hash from URL
        window.history.replaceState(null, '', window.location.pathname);
      }
    }

    // Initialize auth
    await initAuth();

    // React to auth state changes
    user.subscribe(async (currentUser) => {
      if (!currentUser) {
        // Redirect to home if not authenticated
        push('/');
      }
      loading = false;
    });
  });
</script>

<div class="min-h-screen bg-base-100">
  {#if configError}
    <!-- Show configuration error -->
    <div class="flex items-center justify-center min-h-screen p-4">
      <ConfigError error={configError} />
    </div>
  {:else if !$user}
    <!-- Show auth UI when not logged in -->
    <Auth />
  {:else if loading}
    <!-- Loading state -->
    <div class="flex items-center justify-center min-h-screen">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else}
    <!-- Router handles all authenticated routes -->
    <Router {routes} />
  {/if}
</div>

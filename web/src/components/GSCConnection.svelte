<script>
  import { createEventDispatcher, onMount } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  export let summary = null;
  export let navigateToTab = null;
  export let project = null; // Project object with settings

  let isConnecting = false;
  let isConnected = false;
  let selectedProperty = null;
  let properties = [];
  let isLoadingProperties = false;
  let error = null;

  // Get API base URL (same pattern as other API calls)
  const getApiUrl = () => {
    return import.meta.env.VITE_CLOUD_RUN_API_URL || 'http://localhost:8080';
  };

  // Check if already connected on mount
  onMount(() => {
    checkConnection();
  });

  // Also check when summary changes (for backward compatibility)
  $: if (summary) {
    checkConnection();
  }

  async function checkConnection() {
    try {
      // Skip GSC API calls in production (Vercel) - these endpoints don't exist there
      // GSC integration requires the local API server
      if (import.meta.env.PROD || window.location.hostname !== 'localhost') {
        isConnected = false;
        return;
      }

      // Try to get user ID from session storage
      const userID = sessionStorage.getItem('gsc_user_id');
      const apiUrl = getApiUrl();
      const url = userID 
        ? `${apiUrl}/api/gsc/properties?user_id=${encodeURIComponent(userID)}` 
        : `${apiUrl}/api/gsc/properties`;
      
      const response = await fetch(url);
      if (response.ok) {
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
          const props = await response.json();
          if (props && props.length > 0) {
            isConnected = true;
            properties = props;
            if (!selectedProperty && props.length > 0) {
              selectedProperty = props[0].url;
            }
          }
        }
      } else {
        // Handle error response
        try {
          const contentType = response.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const errorData = await response.json();
            console.error('GSC connection check failed:', errorData.error);
          }
        } catch (err) {
          // Not JSON, ignore
        }
      }
    } catch (err) {
      // Not connected or error
      isConnected = false;
    }
  }

  async function connectGSC() {
    isConnecting = true;
    error = null;
    
    try {
      // Skip GSC API calls in production (Vercel) - these endpoints don't exist there
      if (import.meta.env.PROD || window.location.hostname !== 'localhost') {
        error = 'GSC integration is only available when running the local API server (barracuda serve)';
        isConnecting = false;
        return;
      }

      const apiUrl = getApiUrl();
      const response = await fetch(`${apiUrl}/api/gsc/connect`);
      if (!response.ok) {
        // Try to get error message from response
        let errorMessage = 'Failed to get auth URL';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorData.message || errorMessage;
        } catch (parseErr) {
          // If response isn't JSON, try text
          try {
            const text = await response.text();
            if (text) errorMessage = text;
          } catch (textErr) {
            // Use default message
          }
        }
        throw new Error(errorMessage);
      }
      
      const data = await response.json();
      // Open OAuth URL in new window
      const popup = window.open(data.auth_url, 'gsc-auth', 'width=600,height=700');
      
      // Listen for message from popup
      const messageHandler = (event) => {
        if (event.data.type === 'gsc_connected') {
          // Connection successful
          window.removeEventListener('message', messageHandler);
          if (popup) popup.close();
          // Store user ID if provided
          if (event.data.user_id) {
            sessionStorage.setItem('gsc_user_id', event.data.user_id);
          }
          // Refresh connection status
          checkConnection();
          isConnecting = false;
        } else if (event.data.type === 'gsc_error') {
          // Connection failed
          window.removeEventListener('message', messageHandler);
          if (popup) popup.close();
          error = event.data.error || 'Failed to connect to Google Search Console';
          isConnecting = false;
        }
      };
      
      window.addEventListener('message', messageHandler);
      
      // Fallback: check if popup was closed manually
      const checkPopup = setInterval(() => {
        if (popup && popup.closed) {
          clearInterval(checkPopup);
          window.removeEventListener('message', messageHandler);
          if (isConnecting) {
            isConnecting = false;
            // Don't set error if user manually closed - they might try again
          }
        }
      }, 500);
      
      // Timeout after 2 minutes
      setTimeout(() => {
        clearInterval(checkPopup);
        window.removeEventListener('message', messageHandler);
        if (popup && !popup.closed) {
          popup.close();
        }
        if (isConnecting) {
          isConnecting = false;
          error = 'Connection timeout. Please try again.';
        }
      }, 120000); // 2 minutes
    } catch (err) {
      error = err.message;
      isConnecting = false;
    }
  }

  async function pollForConnection() {
    // This function is no longer needed - we use postMessage instead
    // But keeping it for backward compatibility
  }

  async function loadProperties() {
    isLoadingProperties = true;
    error = null;
    
    try {
      // Skip GSC API calls in production (Vercel)
      if (import.meta.env.PROD || window.location.hostname !== 'localhost') {
        error = 'GSC integration is only available when running the local API server';
        return;
      }

      const userID = sessionStorage.getItem('gsc_user_id');
      const apiUrl = getApiUrl();
      const url = userID 
        ? `${apiUrl}/api/gsc/properties?user_id=${encodeURIComponent(userID)}` 
        : `${apiUrl}/api/gsc/properties`;
      
      const response = await fetch(url);
      if (!response.ok) {
        // Handle error response
        let errorMessage = `Failed to load properties (HTTP ${response.status})`;
        try {
          const contentType = response.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const errorData = await response.json();
            errorMessage = errorData.error || errorMessage;
          }
        } catch (parseErr) {
          // Not JSON, use default message
        }
        throw new Error(errorMessage);
      }
      
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const props = await response.json();
        properties = props;
        if (props.length > 0 && !selectedProperty) {
          selectedProperty = props[0].url;
        }
      }
    } catch (err) {
      error = err.message;
} finally {
      isLoadingProperties = false;
    }
  }

  async function enrichIssues() {
    // Use project's saved property if available, otherwise use selectedProperty
    const propertyToUse = project?.settings?.gsc_property_url || selectedProperty;
    
    if (!propertyToUse) {
      error = 'Please select a property first, or save one in Project Settings';
      return;
    }

    error = null;
    
    try {
      const userID = sessionStorage.getItem('gsc_user_id');
      const apiUrl = getApiUrl();
      
      // Get issues from summary
      const issues = summary?.issues || [];
      
      const response = await fetch(`${apiUrl}/api/gsc/enrich-issues`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userID || undefined,
          site_url: propertyToUse,
          days: 30,
          issues: issues,
        }),
      });

      if (!response.ok) {
        // Clone response to read body without consuming it
        const clonedResponse = response.clone();
        let errorMessage = `Failed to enrich issues (HTTP ${response.status})`;
        
        try {
          const errorData = await clonedResponse.json();
          errorMessage = errorData.error || errorMessage;
        } catch (parseErr) {
          // If not JSON, try to get text from original response
          try {
            const text = await response.text();
            errorMessage = text || errorMessage;
          } catch (textErr) {
            // Use default message
          }
        }
        throw new Error(errorMessage);
      }

      // Parse successful response
      const enrichedData = await response.json();
      
      // Store enriched data or trigger refresh
      console.log('Enriched issues:', enrichedData);
      
      // Dispatch event to parent with enriched issues
      dispatch('enriched', enrichedData);

      // Refresh page or navigate to recommendations
      if (navigateToTab) {
        navigateToTab('issues'); // Navigate to issues to see enriched data
      } else {
        window.location.reload();
      }
    } catch (err) {
      error = err.message;
    }
  }
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <h2 class="card-title">Google Search Console Integration</h2>
    
    {#if error}
      <div class="alert alert-error mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{error}</span>
      </div>
    {/if}

    {#if !isConnected}
      <div class="mb-4">
        <p class="text-base-content/70 mb-4">
          Connect your Google Search Console account to enhance recommendations with real search performance data.
          This helps prioritize fixes based on actual traffic and identify optimization opportunities.
        </p>
        
        <button 
          class="btn btn-primary"
          on:click={connectGSC}
          disabled={isConnecting}
        >
          {#if isConnecting}
            <span class="loading loading-spinner loading-sm"></span>
            Connecting...
          {:else}
            🔗 Connect Google Search Console
          {/if}
        </button>
      </div>

      <div class="alert alert-info">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <div>
          <h3 class="font-bold">How It Works</h3>
          <div class="text-sm">
            <p>You don't need to create a Google Cloud project! Just click "Connect" above and authorize Barracuda with your Google account.</p>
            <p class="mt-2">If you see a setup error, advanced users can provide their own OAuth credentials via environment variables:</p>
            <ul class="list-disc list-inside mt-2">
              <li><code>GSC_CLIENT_ID</code> - Your Google OAuth Client ID</li>
              <li><code>GSC_CLIENT_SECRET</code> - Your Google OAuth Client Secret</li>
            </ul>
          </div>
        </div>
      </div>
    {:else}
      <div class="mb-4">
        <div class="alert alert-success mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>Connected to Google Search Console</span>
        </div>

        {#if properties.length > 0}
          <div class="form-control mb-4">
            <label class="label" for="property-select">
              <span class="label-text">Select Property</span>
            </label>
            <select 
              id="property-select"
              class="select select-bordered w-full"
              bind:value={selectedProperty}
            >
              {#each properties as prop}
                <option value={prop.url}>{prop.url}</option>
              {/each}
            </select>
          </div>

          <button 
            class="btn btn-primary"
            on:click={enrichIssues}
          >
            📊 Enrich Issues with GSC Data
          </button>
        {:else}
          <button 
            class="btn btn-ghost"
            on:click={loadProperties}
            disabled={isLoadingProperties}
          >
            {#if isLoadingProperties}
              <span class="loading loading-spinner loading-sm"></span>
              Loading...
            {:else}
              Refresh Properties
            {/if}
          </button>
        {/if}
      </div>
    {/if}
  </div>
</div>


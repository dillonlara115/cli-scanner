<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { updateProjectSettings } from '../lib/data.js';
  
  const dispatch = createEventDispatcher();
  
  export let project = null;
  export let projectId = null;
  export let summary = null; // For enriching issues

  let isConnected = false;
  let properties = [];
  let selectedProperty = null;
  let isLoadingProperties = false;
  let isSaving = false;
  let isEnriching = false;
  let error = null;

  // Get API base URL
  const getApiUrl = () => {
    return import.meta.env.VITE_CLOUD_RUN_API_URL || 'http://localhost:8080';
  };

  onMount(() => {
    // Load saved property from project settings
    if (project?.settings?.gsc_property_url) {
      selectedProperty = project.settings.gsc_property_url;
    }
    
    // Check connection and load properties
    checkConnection();
  });

  async function checkConnection() {
    try {
      if (import.meta.env.PROD || window.location.hostname !== 'localhost') {
        isConnected = false;
        return;
      }

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
            
            // If no property selected, try to match project domain
            if (!selectedProperty && project?.domain) {
              const domain = project.domain.toLowerCase().replace(/^https?:\/\//, '').replace(/\/$/, '');
              const matchingProperty = props.find(p => {
                const propUrl = p.url.toLowerCase().replace(/^https?:\/\//, '').replace(/\/$/, '');
                return propUrl === domain || propUrl === `sc-domain:${domain}`;
              });
              if (matchingProperty) {
                selectedProperty = matchingProperty.url;
              } else if (props.length > 0) {
                selectedProperty = props[0].url;
              }
            } else if (!selectedProperty && props.length > 0) {
              selectedProperty = props[0].url;
            }
          }
        }
      }
    } catch (err) {
      isConnected = false;
    }
  }

  async function loadProperties() {
    isLoadingProperties = true;
    error = null;
    
    try {
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
        throw new Error(`Failed to load properties (HTTP ${response.status})`);
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

  async function saveProperty() {
    if (!selectedProperty || !projectId) return;

    isSaving = true;
    error = null;

    try {
      const { data, error: updateError } = await updateProjectSettings(projectId, {
        gsc_property_url: selectedProperty
      });

      if (updateError) throw updateError;

      // Update local project object
      if (project) {
        project.settings = {
          ...(project.settings || {}),
          gsc_property_url: selectedProperty
        };
      }

      // Show success message briefly
      const successMsg = document.createElement('div');
      successMsg.className = 'alert alert-success mb-4';
      successMsg.textContent = 'Property saved successfully';
      // You could use a toast library here instead
      
    } catch (err) {
      error = err.message || 'Failed to save property';
    } finally {
      isSaving = false;
    }
  }

  async function enrichIssues() {
    const propertyToUse = project?.settings?.gsc_property_url || selectedProperty;
    
    if (!propertyToUse) {
      error = 'Please select and save a property first';
      return;
    }

    if (!summary?.issues || summary.issues.length === 0) {
      error = 'No issues found to enrich';
      return;
    }

    isEnriching = true;
    error = null;
    
    try {
      const userID = sessionStorage.getItem('gsc_user_id');
      const apiUrl = getApiUrl();
      
      const response = await fetch(`${apiUrl}/api/gsc/enrich-issues`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userID || undefined,
          site_url: propertyToUse,
          days: 30,
          issues: summary.issues,
        }),
      });

      if (!response.ok) {
        let errorMessage = `Failed to enrich issues (HTTP ${response.status})`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorMessage;
        } catch (parseErr) {
          try {
            const text = await response.text();
            if (text) errorMessage = text;
          } catch (textErr) {
            // Use default message
          }
        }
        throw new Error(errorMessage);
      }

      const enrichedData = await response.json();
      dispatch('enriched', enrichedData);
      
    } catch (err) {
      error = err.message;
    } finally {
      isEnriching = false;
    }
  }

  // Listen for GSC connection events
  if (typeof window !== 'undefined') {
    window.addEventListener('message', (event) => {
      if (event.data.type === 'gsc_connected') {
        // Refresh properties when GSC is connected
        setTimeout(() => {
          checkConnection();
          loadProperties();
        }, 1000);
      }
    });
  }
</script>

{#if !isConnected}
  <div class="alert alert-info">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
    </svg>
    <span>Connect your Google Search Console account in <a href="/integrations" class="link link-primary">Integrations</a> to select a property for this project.</span>
  </div>
{:else if properties.length === 0}
  <div class="alert alert-warning">
    <span>No Google Search Console properties found. Make sure you have verified access to at least one property.</span>
  </div>
{:else}
  <div class="form-control w-full">
    <label class="label">
      <span class="label-text">Google Search Console Property</span>
    </label>
    <div class="flex gap-2">
      <select 
        class="select select-bordered flex-1"
        bind:value={selectedProperty}
        disabled={isLoadingProperties || isSaving}
      >
        {#each properties as prop}
          <option value={prop.url}>{prop.url}</option>
        {/each}
      </select>
      <button 
        class="btn btn-primary"
        on:click={saveProperty}
        disabled={isSaving || isLoadingProperties || !selectedProperty}
      >
        {#if isSaving}
          <span class="loading loading-spinner loading-sm"></span>
          Saving...
        {:else}
          Save
        {/if}
      </button>
    </div>
    {#if project?.settings?.gsc_property_url}
      <label class="label">
        <span class="label-text-alt text-success">Property saved for this project</span>
      </label>
    {/if}
    {#if summary && summary.issues && summary.issues.length > 0}
      <div class="mt-4">
        <button 
          class="btn btn-error w-full"
          on:click={enrichIssues}
          disabled={isEnriching || !selectedProperty || isLoadingProperties}
        >
          {#if isEnriching}
            <span class="loading loading-spinner loading-sm"></span>
            Enriching...
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
            </svg>
            Enrich Issues with GSC Data
          {/if}
        </button>
      </div>
    {/if}
    {#if error}
      <div class="alert alert-error mt-2">
        <span>{error}</span>
      </div>
    {/if}
  </div>
{/if}


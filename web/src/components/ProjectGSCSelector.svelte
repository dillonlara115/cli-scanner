<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { link } from 'svelte-spa-router';
  import {
    updateProjectSettings,
    fetchProjectGSCStatus,
    fetchProjectGSCProperties,
    updateProjectGSCProperty,
    fetchProjectGSCDimensions,
    triggerProjectGSCSync
  } from '../lib/data.js';
  import { buildEnrichedIssues } from '../lib/gsc.js';
  
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

  let gscStatus = null;
  let gscLoading = false;
  let gscRefreshing = false;
  let gscError = null;
  let lastProjectId = null;
  let propertySelectId = 'gsc-property-select';

  const formatDateTime = (value) => {
    if (!value) return null;
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return null;
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
  };

  $: isConnected = Boolean(gscStatus?.integration?.property_url);
  $: lastSyncedDisplay = gscStatus?.sync_state?.last_synced_at ? formatDateTime(gscStatus.sync_state.last_synced_at) : null;
  $: if (!selectedProperty && project?.settings?.gsc_property_url) {
    selectedProperty = project.settings.gsc_property_url;
  }
  $: if (projectId && projectId !== lastProjectId) {
    lastProjectId = projectId;
    initialize();
  }
  $: propertySelectId = projectId ? `gsc-property-${projectId}` : 'gsc-property-select';

  onMount(() => {
    if (project?.settings?.gsc_property_url) {
      selectedProperty = project.settings.gsc_property_url;
    }

    if (projectId) {
      lastProjectId = projectId;
    }

    initialize();

    if (typeof window !== 'undefined') {
      const handleMessage = async (event) => {
        if (event.data?.type === 'gsc_connected' && event.data?.project_id === projectId) {
          await initialize();
        } else if (event.data?.type === 'gsc_error') {
          error = event.data.error || 'Failed to connect to Google Search Console.';
        }
      };

      window.addEventListener('message', handleMessage);
      return () => window.removeEventListener('message', handleMessage);
    }
  });

  async function initialize() {
    if (!projectId) return;
    await loadStatus();
    if (isConnected) {
      await loadProperties();
    } else {
      properties = [];
    }
  }

  async function loadStatus() {
    if (!projectId) return;

    gscLoading = true;
    gscError = null;
    const statusResult = await fetchProjectGSCStatus(projectId);

    if (statusResult.error) {
      gscStatus = null;
      gscError = statusResult.error.message || 'Unable to load Google Search Console status.';
      gscLoading = false;
      return;
    }

    gscStatus = statusResult.data;
    if (!selectedProperty && gscStatus?.integration?.property_url) {
      selectedProperty = gscStatus.integration.property_url;
    }
    gscLoading = false;
  }

  async function loadProperties() {
    if (!projectId) return;

    isLoadingProperties = true;
    error = null;

    const result = await fetchProjectGSCProperties(projectId);
    if (result.error) {
      error = result.error.message || 'Failed to load properties';
      properties = [];
      isLoadingProperties = false;
      return;
    }

    const payload = result.data || {};
    properties = payload.properties || [];

    if (!selectedProperty) {
      if (payload.selectedProperty) {
        selectedProperty = payload.selectedProperty;
      } else if (properties.length > 0) {
        const domain = project?.domain
          ? project.domain.toLowerCase().replace(/^https?:\/\//, '').replace(/\/$/, '')
          : null;

        if (domain) {
          const matchingProperty = properties.find((prop) => {
            const normalized = typeof prop.url === 'string'
              ? prop.url.toLowerCase().replace(/^https?:\/\//, '').replace(/\/$/, '')
              : '';
            return normalized === domain || normalized === `sc-domain:${domain}`;
          });

          selectedProperty = matchingProperty?.url || properties[0].url;
        } else {
          selectedProperty = properties[0].url;
        }
      } else {
        selectedProperty = null;
      }
    }

    isLoadingProperties = false;
  }

  async function saveProperty() {
    if (!selectedProperty || !projectId) return;

    isSaving = true;
    error = null;

    const saveResult = await updateProjectGSCProperty(projectId, selectedProperty);
    if (saveResult.error) {
      error = saveResult.error.message || 'Failed to save property';
      isSaving = false;
      return;
    }

    const settingsResult = await updateProjectSettings(projectId, {
      gsc_property_url: selectedProperty
    });
    if (settingsResult.error) {
      error = settingsResult.error.message || 'Failed to persist property selection';
    }

    if (project) {
      project.settings = {
        ...(project.settings || {}),
        gsc_property_url: selectedProperty
      };
    }

    await loadStatus();
    isSaving = false;
  }

  async function refreshGSCData() {
    if (!projectId) return;
    gscRefreshing = true;
    error = null;

    const result = await triggerProjectGSCSync(projectId, { lookback_days: 30 });
    if (result.error) {
      error = result.error.message || 'Failed to refresh Google Search Console data';
      gscRefreshing = false;
      return;
    }

    await initialize();
    gscRefreshing = false;
  }

  async function enrichIssues() {
    if (!projectId) {
      error = 'Project context is missing';
      return;
    }

    if (!summary?.issues || summary.issues.length === 0) {
      error = 'No issues found to enrich';
      return;
    }

    isEnriching = true;
    error = null;

    const result = await fetchProjectGSCDimensions(projectId, 'page', { limit: 1000 });
    if (result.error) {
      error = result.error.message || 'Failed to load cached Search Console metrics';
      isEnriching = false;
      return;
    }

    const rows = result.data?.rows || [];
    if (rows.length === 0) {
      error = 'No cached Search Console data found. Refresh the integration first.';
      isEnriching = false;
      return;
    }

    const enrichedData = buildEnrichedIssues(summary.issues, rows);
    if (!enrichedData.length) {
      error = 'No Search Console data matched the current issues yet.';
    } else {
      dispatch('enriched', enrichedData);
    }

    isEnriching = false;
  }
</script>

{#if gscLoading}
  <div class="alert alert-info">
    <span>Loading Google Search Console status...</span>
  </div>
{:else if gscError}
  <div class="alert alert-warning">
    <span>{gscError}</span>
  </div>
{:else if !isConnected}
  <div class="alert alert-info">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
    </svg>
    <span>Connect your Google Search Console account in <a href="/integrations" use:link class="link link-primary">Integrations</a> to select a property for this project.</span>
  </div>
{:else}
  <div class="space-y-4">
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-3 rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
      <div>
        <div class="text-sm font-semibold text-base-content/80">Google Search Console</div>
        <div class="text-sm">
          Connected to <span class="font-semibold">{gscStatus?.integration?.property_url}</span>.
        </div>
        {#if lastSyncedDisplay}
          <div class="text-xs text-base-content/60">Last synced {lastSyncedDisplay}</div>
        {:else}
          <div class="text-xs text-base-content/60">No cached metrics yet. Refresh to pull the latest data.</div>
        {/if}
      </div>
      <div class="flex gap-2">
        <button
          class="btn btn-sm btn-outline"
          on:click={refreshGSCData}
          disabled={gscRefreshing || gscLoading}
        >
          {#if gscRefreshing}
            <span class="loading loading-spinner loading-xs"></span>
            Refreshing...
          {:else}
            Refresh Data
          {/if}
        </button>
      </div>
    </div>

    {#if properties.length === 0}
      <div class="alert alert-warning">
        <span>No Google Search Console properties found. Confirm access in Search Console.</span>
      </div>
    {:else}
      <div class="form-control w-full">
        <label class="label" for={propertySelectId}>
          <span class="label-text">Google Search Console Property</span>
        </label>
        <div class="flex gap-2">
          <select
            id={propertySelectId}
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
          <div class="label">
            <span class="label-text-alt text-success">Property saved for this project</span>
          </div>
        {/if}
      </div>
    {/if}

    {#if summary && summary.issues && summary.issues.length > 0}
      <div>
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
      <div class="alert alert-error">
        <span>{error}</span>
      </div>
    {/if}
  </div>
{/if}

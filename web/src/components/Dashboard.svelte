<script>
  import { onMount } from 'svelte';
  import { push, querystring, link } from 'svelte-spa-router';
  import SummaryCard from './SummaryCard.svelte';
  import ResultsTable from './ResultsTable.svelte';
  import IssuesPanel from './IssuesPanel.svelte';
  import LinkGraph from './LinkGraph.svelte';
  import RecommendationsPanel from './RecommendationsPanel.svelte';
  import Logo from './Logo.svelte';
  import { fetchProjects, fetchProjectGSCStatus, fetchProjectGSCDimensions, triggerProjectGSCSync } from '../lib/data.js';
  import { buildEnrichedIssues } from '../lib/gsc.js';

  export let summary = null;
  export let results = [];
  export let initialTab = 'dashboard';
  export let projectId = null;
  export let crawlId = null;
  export let project = null; // Accept project as prop from parent

  // Load project if not provided
  onMount(async () => {
    if (!project && projectId) {
      const { data: projects } = await fetchProjects();
      if (projects) {
        project = projects.find(p => p.id === projectId);
      }
    }
  });

  $: activeTab = $querystring 
    ? new URLSearchParams($querystring).get('tab') || initialTab 
    : initialTab;
  let issuesFilter = { severity: 'all', type: 'all', url: null };
  let resultsFilter = { status: 'all', performance: false };
  let cachedEnrichedIssues = [];
  let activeEnrichedIssues = [];
  let enrichedIssuesMap = {};
  let gscStatus = null;
  let gscLoading = false;
  let gscRefreshing = false;
  let gscError = null;
  let gscPageRows = [];
  let gscInitializedProjectId = null;

  const navigateToTab = (tab, nextFilters = {}) => {
    // Update URL with tab query param
    if (projectId && crawlId) {
      const params = new URLSearchParams();
      params.set('tab', tab);
      push(`/project/${projectId}/crawl/${crawlId}?${params.toString()}`);
    }

    const { severity, type, url, status, performance } = nextFilters;

    if (severity !== undefined || type !== undefined || url !== undefined) {
      issuesFilter = {
        ...issuesFilter,
        ...(severity !== undefined ? { severity } : {}),
        ...(type !== undefined ? { type } : {}),
        ...(url !== undefined ? { url } : {})
      };
    }

    if (status !== undefined || performance !== undefined) {
      resultsFilter = {
        ...resultsFilter,
        ...(status !== undefined ? { status } : {}),
        ...(performance !== undefined ? { performance } : {})
      };
    }
  };

  // Callback for GSC to update enriched issues
  const formatDateTime = (value) => {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
  };

  async function loadGSCData(targetProjectId) {
    if (!targetProjectId) return;

    gscLoading = true;
    gscError = null;
    gscStatus = null;
    gscPageRows = [];

    const statusResult = await fetchProjectGSCStatus(targetProjectId);
    if (statusResult.error) {
      gscError = statusResult.error.message || 'Unable to load Google Search Console status.';
      gscLoading = false;
      return;
    }

    gscStatus = statusResult.data;

    if (gscStatus?.integration?.property_url) {
      const pageResult = await fetchProjectGSCDimensions(targetProjectId, 'page', { limit: 1000 });
      if (pageResult.error) {
        gscError = pageResult.error.message || 'Unable to load Search Console metrics.';
      } else {
        gscPageRows = pageResult.data?.rows || [];
      }
    }

    gscLoading = false;
  }

  async function refreshGSCData() {
    if (!projectId || gscRefreshing) return;
    gscRefreshing = true;
    gscError = null;
    gscLoading = true;
    const syncResult = await triggerProjectGSCSync(projectId, { lookback_days: 30 });
    if (syncResult.error) {
      gscError = syncResult.error.message || 'Failed to refresh Google Search Console data.';
      gscRefreshing = false;
      gscLoading = false;
      return;
    }
    await loadGSCData(projectId);
    gscRefreshing = false;
  }

  $: if (projectId && projectId !== gscInitializedProjectId) {
    gscInitializedProjectId = projectId;
    loadGSCData(projectId);
  }

  $: cachedEnrichedIssues = buildEnrichedIssues(summary?.issues || [], gscPageRows);

  $: activeEnrichedIssues = cachedEnrichedIssues;

  $: displayIssues = activeEnrichedIssues.length > 0
    ? activeEnrichedIssues.map((ei) => ei.issue)
    : (summary?.issues || []);
  
  $: enrichedIssuesMap = activeEnrichedIssues.reduce((acc, ei) => {
    if (ei?.issue?.url && ei?.issue?.type) {
      acc[`${ei.issue.url}|${ei.issue.type}`] = ei;
    }
    return acc;
  }, {});

  $: gscProperty = gscStatus?.integration?.property_url || null;
  $: gscLastSynced = gscStatus?.sync_state?.last_synced_at ? formatDateTime(gscStatus.sync_state.last_synced_at) : null;
</script>

<div class="navbar bg-base-100 shadow-lg border-b border-base-300">
 
  <div class="flex-none">
    <ul class="menu menu-horizontal px-1 flex gap-2">
      <li>
        <button 
          type="button" 
          class="btn btn-ghost {activeTab === 'dashboard' ? 'bg-primary text-primary-content' : ''}"
          on:click={() => navigateToTab('dashboard')}
        >
          Dashboard
        </button>
      </li>
      <li>
        <button 
          type="button" 
          class="btn btn-ghost {activeTab === 'results' ? 'bg-primary text-primary-content' : ''}"
          on:click={() => navigateToTab('results')}
        >
          Results
        </button>
      </li>
      <li>
        <button 
          type="button" 
          class="btn btn-ghost {activeTab === 'issues' ? 'bg-primary text-primary-content' : ''}"
          on:click={() => navigateToTab('issues')}
        >
          Issues
        </button>
      </li>
      <li>
        <button 
          type="button" 
          class="btn btn-ghost {activeTab === 'recommendations' ? 'bg-primary text-primary-content' : ''}"
          on:click={() => navigateToTab('recommendations')}
        >
          Recommendations
        </button>
      </li>
      <li>
        <button 
          type="button" 
          class="btn btn-ghost {activeTab === 'graph' ? 'bg-primary text-primary-content' : ''}"
          on:click={() => navigateToTab('graph')}
        >
          Link Graph
        </button>
      </li>
      {#if projectId}
        <li>
          <a 
            href="/project/{projectId}/settings" 
            use:link
            class="btn btn-ghost"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Settings
          </a>
        </li>
      {/if}
    </ul>
  </div>
</div>

<div class="container mx-auto p-4">
  {#if activeTab === 'dashboard'}
    <div class="space-y-4">
      {#if gscError}
        <div class="alert alert-warning flex flex-col md:flex-row md:items-center md:justify-between gap-3">
          <span>{gscError}</span>
          <div class="flex gap-2">
            <button class="btn btn-sm btn-outline" on:click={() => loadGSCData(projectId)} disabled={gscLoading}>
              {#if gscLoading}
                <span class="loading loading-spinner loading-xs"></span>
                Retrying...
              {:else}
                Retry
              {/if}
            </button>
            <a class="btn btn-sm btn-ghost" href="/integrations">Manage</a>
          </div>
        </div>
      {:else if gscProperty}
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-3 rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
          <div>
            <div class="text-sm font-semibold text-base-content/80">Google Search Console</div>
            <div class="text-sm">
              Connected to <span class="font-semibold">{gscProperty}</span>.
              {#if gscLoading}
                <span class="ml-1 text-xs text-base-content/60">(Refreshing...)</span>
              {/if}
            </div>
            {#if gscLastSynced}
              <div class="text-xs text-base-content/60">Last synced {gscLastSynced}</div>
            {:else}
              <div class="text-xs text-base-content/60">No cached metrics yet. Run a refresh to pull the latest data.</div>
            {/if}
          </div>
          <div class="flex gap-2">
            <button
              class="btn btn-sm btn-outline"
              on:click={refreshGSCData}
              disabled={gscRefreshing || gscLoading}
            >
              {#if gscRefreshing || gscLoading}
                <span class="loading loading-spinner loading-xs"></span>
                Refreshing...
              {:else}
                Refresh Data
              {/if}
            </button>
            <a class="btn btn-sm btn-ghost" href="/integrations">Manage</a>
          </div>
        </div>
      {:else if gscLoading}
        <div class="alert alert-info">
          <span>Loading Google Search Console status...</span>
        </div>
      {:else}
        <div class="alert alert-info">
          <span>
            Connect Google Search Console in
            <a class="link link-primary" href="/integrations">Integrations</a>
            to surface search performance metrics alongside crawl data.
          </span>
        </div>
      {/if}

      <SummaryCard
        {summary}
        {navigateToTab}
        gscTotals={gscStatus?.summary?.totals}
        gscSyncState={gscStatus?.sync_state}
        gscIntegration={gscStatus?.integration}
        gscLoading={gscLoading}
        gscError={gscError}
      />
    </div>
  {:else if activeTab === 'results'}
    <ResultsTable 
      {results} 
      issues={displayIssues}
      filter={resultsFilter}
      {navigateToTab}
    />
  {:else if activeTab === 'issues'}
    <IssuesPanel
      issues={displayIssues}
      filter={issuesFilter}
      enrichedIssues={enrichedIssuesMap}
      gscStatus={gscStatus}
      gscLoading={gscLoading}
      gscError={gscError}
    />
  {:else if activeTab === 'recommendations'}
    <div class="space-y-4">
      <RecommendationsPanel issues={displayIssues} {navigateToTab} enrichedIssues={enrichedIssuesMap} />
    </div>
  {:else if activeTab === 'graph'}
    <LinkGraph crawlId={crawlId} />
  {/if}
</div>

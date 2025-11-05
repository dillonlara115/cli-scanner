<script>
  import { push, querystring, link } from 'svelte-spa-router';
  import SummaryCard from './SummaryCard.svelte';
  import ResultsTable from './ResultsTable.svelte';
  import IssuesPanel from './IssuesPanel.svelte';
  import LinkGraph from './LinkGraph.svelte';
  import RecommendationsPanel from './RecommendationsPanel.svelte';
  import GSCConnection from './GSCConnection.svelte';
  import Logo from './Logo.svelte';

  export let summary = null;
  export let results = [];
  export let initialTab = 'dashboard';
  export let projectId = null;
  export let crawlId = null;

  $: activeTab = $querystring 
    ? new URLSearchParams($querystring).get('tab') || initialTab 
    : initialTab;
  let issuesFilter = { severity: 'all', type: 'all', url: null };
  let resultsFilter = { status: 'all', performance: false };
  
  // Store enriched issues from GSC
  let enrichedIssues = [];
  let useEnrichedIssues = false;

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
  const handleEnrichedIssues = (enriched) => {
    enrichedIssues = enriched;
    useEnrichedIssues = true;
    // Navigate to issues tab to see enriched data
    navigateToTab('issues');
  };

  // Get issues to display - use enriched if available, otherwise regular
  $: displayIssues = useEnrichedIssues && enrichedIssues.length > 0 
    ? enrichedIssues.map(ei => ei.issue) 
    : (summary?.issues || []);
  
  // Get enriched issues data for components that need it
  $: enrichedIssuesMap = enrichedIssues.reduce((acc, ei) => {
    acc[ei.issue.url + '|' + ei.issue.type] = ei;
    return acc;
  }, {});
</script>

<div class="navbar bg-base-100 shadow-lg border-b border-base-300">
 
  <div class="flex-none">
    <ul class="menu menu-horizontal px-1">
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
    </ul>
  </div>
</div>

<div class="container mx-auto p-4">
  {#if activeTab === 'dashboard'}
    <SummaryCard {summary} {navigateToTab} />
  {:else if activeTab === 'results'}
    <ResultsTable 
      {results} 
      issues={displayIssues}
      filter={resultsFilter}
      {navigateToTab}
    />
  {:else if activeTab === 'issues'}
    <IssuesPanel issues={displayIssues} filter={issuesFilter} enrichedIssues={enrichedIssuesMap} />
  {:else if activeTab === 'recommendations'}
    <div class="space-y-4">
      <GSCConnection {summary} {navigateToTab} on:enriched={e => handleEnrichedIssues(e.detail)} />
      <RecommendationsPanel issues={displayIssues} {navigateToTab} enrichedIssues={enrichedIssuesMap} />
    </div>
  {:else if activeTab === 'graph'}
    <LinkGraph />
  {/if}
</div>

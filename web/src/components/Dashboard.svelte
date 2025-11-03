<script>
  import SummaryCard from './SummaryCard.svelte';
  import ResultsTable from './ResultsTable.svelte';
  import IssuesPanel from './IssuesPanel.svelte';
  import LinkGraph from './LinkGraph.svelte';
  import RecommendationsPanel from './RecommendationsPanel.svelte';
  import GSCConnection from './GSCConnection.svelte';

  export let summary = null;
  export let results = [];

  let activeTab = 'dashboard';
  let issuesFilter = { severity: 'all', type: 'all', url: null };
  let resultsFilter = { status: 'all', performance: false };
  
  // Store enriched issues from GSC
  let enrichedIssues = [];
  let useEnrichedIssues = false;

  const navigateToTab = (tab, nextFilters = {}) => {
    activeTab = tab;

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
    activeTab = 'issues';
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

<div class="navbar bg-base-300 shadow-lg">
  <div class="flex-1">
    <a href="/" class="btn btn-ghost text-xl">Barracuda</a>
  </div>
  <div class="flex-none">
    <ul class="menu menu-horizontal px-1">
      <li><button type="button" class:active={activeTab === 'dashboard'} on:click={() => activeTab = 'dashboard'}>Dashboard</button></li>
      <li><button type="button" class:active={activeTab === 'results'} on:click={() => activeTab = 'results'}>Results</button></li>
      <li><button type="button" class:active={activeTab === 'issues'} on:click={() => activeTab = 'issues'}>Issues</button></li>
      <li><button type="button" class:active={activeTab === 'recommendations'} on:click={() => activeTab = 'recommendations'}>Recommendations</button></li>
      <li><button type="button" class:active={activeTab === 'graph'} on:click={() => activeTab = 'graph'}>Link Graph</button></li>
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

<style>
  .active {
    @apply bg-base-100;
  }
</style>

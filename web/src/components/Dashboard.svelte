<script>
  import SummaryCard from './SummaryCard.svelte';
  import ResultsTable from './ResultsTable.svelte';
  import IssuesPanel from './IssuesPanel.svelte';
  import LinkGraph from './LinkGraph.svelte';

  export let summary = null;
  export let results = [];

  let activeTab = 'dashboard';
  let issuesFilter = { severity: 'all', type: 'all', url: null };
  let resultsFilter = { status: 'all', performance: false };

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
</script>

<div class="navbar bg-base-300 shadow-lg">
  <div class="flex-1">
    <a href="/" class="btn btn-ghost text-xl">🐊 Barracuda</a>
  </div>
  <div class="flex-none">
    <ul class="menu menu-horizontal px-1">
      <li><button type="button" class:active={activeTab === 'dashboard'} on:click={() => activeTab = 'dashboard'}>Dashboard</button></li>
      <li><button type="button" class:active={activeTab === 'results'} on:click={() => activeTab = 'results'}>Results</button></li>
      <li><button type="button" class:active={activeTab === 'issues'} on:click={() => activeTab = 'issues'}>Issues</button></li>
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
      issues={summary?.issues || []}
      filter={resultsFilter}
      {navigateToTab}
    />
  {:else if activeTab === 'issues'}
    <IssuesPanel issues={summary?.issues || []} filter={issuesFilter} />
  {:else if activeTab === 'graph'}
    <LinkGraph />
  {/if}
</div>

<style>
  .active {
    @apply bg-base-100;
  }
</style>

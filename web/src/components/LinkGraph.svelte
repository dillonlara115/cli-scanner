<script>
  import { onMount } from 'svelte';
  import { fetchCrawlGraph } from '../lib/data.js';

  export let crawlId = null;

  let graphData = null;
  let loading = true;
  let error = null;

  onMount(async () => {
    if (!crawlId) {
      error = 'No crawl ID provided';
      loading = false;
      return;
    }

    await loadGraph();
  });

  async function loadGraph() {
    if (!crawlId) return;

    loading = true;
    error = null;

    try {
      const { data, error: fetchError } = await fetchCrawlGraph(crawlId);
      if (fetchError) {
        throw fetchError;
      }
      graphData = data || {};
    } catch (err) {
      console.error('Error loading link graph:', err);
      error = err.message || 'Failed to load link graph';
    } finally {
      loading = false;
    }
  }

  // Calculate stats
  $: totalNodes = graphData ? Object.keys(graphData).length : 0;
  $: totalEdges = graphData ? Object.values(graphData).reduce((sum, links) => sum + (links?.length || 0), 0) : 0;
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <div class="flex justify-between items-center mb-4">
      <h2 class="card-title">Link Graph Visualization</h2>
      {#if graphData && totalNodes > 0}
        <div class="badge badge-info badge-lg">
          {totalNodes} pages, {totalEdges} links
        </div>
      {/if}
    </div>
    
    {#if loading}
      <div class="flex justify-center py-8">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    {:else if error}
      <div class="alert alert-error">
        <span>Error: {error}</span>
      </div>
    {:else if !graphData || totalNodes === 0}
      <div class="alert alert-info">
        <span>No link graph data available for this crawl.</span>
      </div>
    {:else}
      <div class="space-y-4">
        <!-- Stats Summary -->
        <div class="stats stats-vertical lg:stats-horizontal shadow w-full">
          <div class="stat">
            <div class="stat-title">Total Pages</div>
            <div class="stat-value text-primary">{totalNodes}</div>
          </div>
          <div class="stat">
            <div class="stat-title">Total Links</div>
            <div class="stat-value text-secondary">{totalEdges}</div>
          </div>
          <div class="stat">
            <div class="stat-title">Avg Links/Page</div>
            <div class="stat-value text-accent">{totalNodes > 0 ? (totalEdges / totalNodes).toFixed(1) : 0}</div>
          </div>
        </div>

        <!-- Graph Visualization -->
        <div class="overflow-x-auto max-h-[600px] overflow-y-auto border border-base-300 rounded-lg p-4">
          <div class="space-y-3">
            {#each Object.entries(graphData) as [source, targets]}
              <div class="border-l-4 border-primary pl-4 py-2 hover:bg-base-200 transition-colors">
                <div class="font-semibold text-sm mb-2 break-all">
                  <a href={source} target="_blank" rel="noopener noreferrer" class="link link-primary">
                    {source}
                  </a>
                </div>
                <div class="space-y-1 ml-4">
                  {#each targets as target}
                    <div class="text-sm text-base-content/70 flex items-center gap-2">
                      <span class="text-primary">→</span>
                      <a href={target} target="_blank" rel="noopener noreferrer" class="link link-secondary break-all">
                        {target}
                      </a>
                    </div>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>


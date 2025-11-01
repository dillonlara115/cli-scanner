<script>
  import { onMount } from 'svelte';

  let graphData = null;
  let loading = true;

  onMount(async () => {
    try {
      const res = await fetch('/api/graph');
      if (res.ok) {
        graphData = await res.json();
      }
      loading = false;
    } catch (err) {
      loading = false;
    }
  });
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <h2 class="card-title">Link Graph Visualization</h2>
    
    {#if loading}
      <div class="flex justify-center py-8">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    {:else if !graphData || Object.keys(graphData).length === 0}
      <div class="alert alert-info">
        <span>No link graph data available. Run a crawl with --graph-export flag.</span>
      </div>
    {:else}
      <div class="overflow-x-auto">
        <div class="space-y-4">
          {#each Object.entries(graphData) as [source, targets]}
            <div class="border-l-4 border-primary pl-4">
              <div class="font-semibold text-sm mb-2 break-all">{source}</div>
              <div class="space-y-1">
                {#each targets as target}
                  <div class="text-sm text-base-content/70 ml-4">
                    → <a href={target} target="_blank" class="link link-primary break-all">{target}</a>
                  </div>
                {/each}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>


<script>
  import Dashboard from './components/Dashboard.svelte';
  import { onMount } from 'svelte';

  let loading = true;
  let error = null;
  let summary = null;
  let results = [];

  onMount(async () => {
    try {
      // Try to load existing results
      const summaryRes = await fetch('/api/summary');
      const resultsRes = await fetch('/api/results');

      if (summaryRes.ok) {
        summary = await summaryRes.json();
      }

      if (resultsRes.ok) {
        results = await resultsRes.json();
      }

      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
    }
  });
</script>

<div class="min-h-screen bg-base-200">
  {#if loading}
    <div class="flex items-center justify-center min-h-screen">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else if error}
    <div class="flex items-center justify-center min-h-screen">
      <div class="alert alert-error">
        <span>Error loading data: {error}</span>
      </div>
    </div>
  {:else}
    <Dashboard {summary} {results} />
  {/if}
</div>


<script>
  export let summary = null;
  export let navigateToTab = () => {};

  if (!summary) {
    summary = {
      total_pages: 0,
      total_issues: 0,
      average_response_time_ms: 0,
      pages_with_errors: 0,
      pages_with_redirects: 0,
      total_internal_links: 0,
      total_external_links: 0,
      issues_by_type: {}
    };
  }

  const getSeverityCount = (severity) => {
    if (!summary.issues) return 0;
    return summary.issues.filter(i => i.severity === severity).length;
  };

  const handleTotalIssuesClick = () => {
    if (summary.total_issues > 0) {
      navigateToTab('issues');
    }
  };

  const handleFixCriticalIssues = () => {
    navigateToTab('issues', { severity: 'error' });
  };

  const handleViewSlowPages = () => {
    navigateToTab('results', { performance: true });
  };
</script>

<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
  <div class="stat bg-base-100 rounded-box shadow">
    <div class="stat-figure text-primary">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
    </div>
    <div class="stat-title">Total Pages</div>
    <div class="stat-value text-primary">{summary.total_pages}</div>
  </div>

  <div 
    class="stat bg-base-100 rounded-box shadow cursor-pointer hover:shadow-lg transition-shadow"
    role="button"
    tabindex="0"
    on:click={handleTotalIssuesClick}
    on:keydown={(e) => e.key === 'Enter' && handleTotalIssuesClick()}
  >
    <div class="stat-figure text-error">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
      </svg>
    </div>
    <div class="stat-title">Total Issues</div>
    <div class="stat-value text-error">{summary.total_issues}</div>
    {#if summary.total_issues > 0}
      <div class="stat-desc text-xs mt-1">Click to view all issues</div>
    {/if}
  </div>

  <div class="stat bg-base-100 rounded-box shadow">
    <div class="stat-figure text-info">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
      </svg>
    </div>
    <div class="stat-title">Avg Response Time</div>
    <div class="stat-value text-info">{summary.average_response_time_ms}ms</div>
  </div>

  <div class="stat bg-base-100 rounded-box shadow">
    <div class="stat-figure text-warning">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
    </div>
    <div class="stat-title">Pages with Errors</div>
    <div class="stat-value text-warning">{summary.pages_with_errors}</div>
  </div>
</div>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
  <div class="card bg-base-100 shadow">
    <div class="card-body">
      <h2 class="card-title">Severity Breakdown</h2>
      <div class="space-y-2">
        <div class="flex justify-between">
          <span class="text-error">Errors:</span>
          <span class="font-bold">{getSeverityCount('error')}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-warning">Warnings:</span>
          <span class="font-bold">{getSeverityCount('warning')}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-info">Info:</span>
          <span class="font-bold">{getSeverityCount('info')}</span>
        </div>
        {#if getSeverityCount('error') > 0}
          <div class="card-actions justify-end mt-4">
            <button class="btn btn-error btn-sm" on:click={handleFixCriticalIssues}>
              Fix Critical Issues
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <div class="card bg-base-100 shadow">
    <div class="card-body">
      <h2 class="card-title">Link Statistics</h2>
      <div class="space-y-2">
        <div class="flex justify-between">
          <span>Internal Links:</span>
          <span class="font-bold">{summary.total_internal_links}</span>
        </div>
        <div class="flex justify-between">
          <span>External Links:</span>
          <span class="font-bold">{summary.total_external_links}</span>
        </div>
        <div class="flex justify-between">
          <span>Pages with Redirects:</span>
          <span class="font-bold">{summary.pages_with_redirects}</span>
        </div>
      </div>
    </div>
  </div>

  <div class="card bg-base-100 shadow">
    <div class="card-body">
      <h2 class="card-title">Issue Types</h2>
      <div class="space-y-1">
        {#each Object.entries(summary.issues_by_type || {}) as [type, count]}
          <div class="flex justify-between text-sm">
            <span class="truncate">{type.replace(/_/g, ' ')}</span>
            <span class="badge badge-primary">{count}</span>
          </div>
        {/each}
      </div>
      {#if summary.total_issues > 0}
        <div class="card-actions justify-end mt-4">
          <button class="btn btn-primary btn-sm" on:click={() => navigateToTab('issues')}>
            View All Issues
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if summary.slowest_pages && summary.slowest_pages.length > 0}
  <div class="card bg-base-100 shadow">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">Slowest Pages</h2>
        <button class="btn btn-outline btn-sm" on:click={handleViewSlowPages}>
          View All Slow Pages
        </button>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>URL</th>
              <th>Response Time</th>
            </tr>
          </thead>
          <tbody>
            {#each summary.slowest_pages.slice(0, 10) as page}
              <tr>
                <td class="max-w-md truncate">{page.url}</td>
                <td>{page.response_time_ms}ms</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  </div>
{/if}


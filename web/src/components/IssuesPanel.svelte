<script>
  export let issues = [];

  let severityFilter = 'all';
  let typeFilter = 'all';

  $: filteredIssues = issues.filter(i => {
    const matchesSeverity = severityFilter === 'all' || i.severity === severityFilter;
    const matchesType = typeFilter === 'all' || i.type === typeFilter;
    return matchesSeverity && matchesType;
  });

  $: uniqueTypes = [...new Set(issues.map(i => i.type))];
  
  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'error': return 'text-error';
      case 'warning': return 'text-warning';
      case 'info': return 'text-info';
      default: return '';
    }
  };

  const getSeverityBadge = (severity) => {
    switch (severity) {
      case 'error': return 'badge-error';
      case 'warning': return 'badge-warning';
      case 'info': return 'badge-info';
      default: return 'badge-ghost';
    }
  };
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <h2 class="card-title mb-4">SEO Issues</h2>

    <div class="flex flex-col md:flex-row gap-4 mb-4">
      <select class="select select-bordered" bind:value={severityFilter}>
        <option value="all">All Severities</option>
        <option value="error">Errors</option>
        <option value="warning">Warnings</option>
        <option value="info">Info</option>
      </select>
      <select class="select select-bordered" bind:value={typeFilter}>
        <option value="all">All Types</option>
        {#each uniqueTypes as type}
          <option value={type}>{type.replace(/_/g, ' ')}</option>
        {/each}
      </select>
    </div>

    <div class="text-sm text-base-content/70 mb-4">
      Showing {filteredIssues.length} of {issues.length} issues
    </div>

    <div class="space-y-4">
      {#each filteredIssues as issue}
        <div class="alert {getSeverityBadge(issue.severity)} shadow-lg">
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-2">
              <span class="badge {getSeverityBadge(issue.severity)}">
                {issue.severity}
              </span>
              <span class="badge badge-outline">
                {issue.type.replace(/_/g, ' ')}
              </span>
            </div>
            <h3 class="font-bold">{issue.message}</h3>
            <div class="text-sm mt-2">
              <div class="font-semibold">URL:</div>
              <a href={issue.url} target="_blank" class="link link-primary break-all">
                {issue.url}
              </a>
            </div>
            {#if issue.value}
              <div class="text-sm mt-2">
                <div class="font-semibold">Value:</div>
                <div class="break-all">{issue.value}</div>
              </div>
            {/if}
            {#if issue.recommendation}
              <div class="text-sm mt-2">
                <div class="font-semibold">Recommendation:</div>
                <div>{issue.recommendation}</div>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>

    {#if filteredIssues.length === 0}
      <div class="alert alert-success">
        <span>🎉 No issues found!</span>
      </div>
    {/if}
  </div>
</div>


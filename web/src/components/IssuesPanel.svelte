<script>
  export let issues = [];
  export let filter = { severity: 'all', type: 'all' };

  let severityFilter = 'all';
  let typeFilter = 'all';

  // Initialize filters from prop
  $: {
    if (filter.severity) severityFilter = filter.severity;
    if (filter.type) typeFilter = filter.type;
  }

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

  const timestamp = () => {
    const now = new Date();
    const pad = (value) => value.toString().padStart(2, '0');
    return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  };

  const downloadFile = (content, fileName, mimeType) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const exportAsJson = () => {
    const fileName = `issues-${timestamp()}.json`;
    const content = JSON.stringify(filteredIssues, null, 2);
    downloadFile(content, fileName, 'application/json');
  };

  const toCsv = (rows) => {
    const escapeValue = (value) => {
      if (value === null || value === undefined) return '';
      const stringValue = String(value);
      return /[",\n]/.test(stringValue) ? `"${stringValue.replace(/"/g, '""')}"` : stringValue;
    };

    const headers = ['url', 'type', 'severity', 'message', 'recommendation', 'value'];
    const dataRows = rows.map(issue => [
      issue.url || '',
      issue.type || '',
      issue.severity || '',
      issue.message || '',
      issue.recommendation || '',
      issue.value || ''
    ]);

    return [headers, ...dataRows]
      .map(row => row.map(escapeValue).join(','))
      .join('\n');
  };

  const exportAsCsv = () => {
    const fileName = `issues-${timestamp()}.csv`;
    const content = toCsv(filteredIssues);
    downloadFile(content, fileName, 'text/csv');
  };

  const closeDropdown = (event) => {
    const details = event?.currentTarget?.closest('details');
    if (details) {
      details.removeAttribute('open');
    }
  };

  const handleExportCsv = (event) => {
    exportAsCsv();
    closeDropdown(event);
  };

  const handleExportJson = (event) => {
    exportAsJson();
    closeDropdown(event);
  };
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <h2 class="card-title mb-4">SEO Issues</h2>

    <div class="flex flex-col md:flex-row md:items-start md:justify-between gap-4 mb-4">
      <div class="flex flex-col md:flex-row gap-4">
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
      <details class="dropdown dropdown-end">
        <summary class="btn btn-primary select-none">Export Issues</summary>
        <ul class="dropdown-content menu bg-base-200 rounded-box w-52 shadow mt-2">
          <li>
            <button type="button" on:click={handleExportCsv}>
              Export as CSV
            </button>
          </li>
          <li>
            <button type="button" on:click={handleExportJson}>
              Export as JSON
            </button>
          </li>
        </ul>
      </details>
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

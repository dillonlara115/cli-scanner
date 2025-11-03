<script>
  export let issues = [];
  export let filter = { severity: 'all', type: 'all', url: null };

  const normalizeSeverity = (value) => value ?? 'all';
  const normalizeType = (value) => value ?? 'all';
  const normalizeUrl = (value) => value ?? '';

  const computeFilterSignature = (incoming = {}) => [
    normalizeSeverity(incoming.severity),
    normalizeType(incoming.type),
    normalizeUrl(incoming.url)
  ].join('|');

  const applyFiltersFromProps = (incoming = {}) => {
    severityFilter = normalizeSeverity(incoming.severity);
    typeFilter = normalizeType(incoming.type);
    searchTerm = normalizeUrl(incoming.url);
  };

  // Initialize filters - user can change these locally
  let severityFilter = normalizeSeverity(filter?.severity);
  let typeFilter = normalizeType(filter?.type);
  let searchTerm = normalizeUrl(filter?.url);
  let groupBy = 'none'; // 'none', 'url', 'type', 'severity'
  let sortBy = 'none'; // 'none', 'priority'

  let lastAppliedFilterSignature = computeFilterSignature(filter);

  // Sync local filters when parent provides a new filter object (e.g., via navigation)
  $: {
    const nextSignature = computeFilterSignature(filter);
    if (nextSignature !== lastAppliedFilterSignature) {
      applyFiltersFromProps(filter);
      lastAppliedFilterSignature = nextSignature;
    }
  }

  // Calculate affected pages count for each issue type
  $: affectedPagesByType = issues.reduce((acc, issue) => {
    if (!acc[issue.type]) {
      acc[issue.type] = new Set();
    }
    acc[issue.type].add(issue.url);
    return acc;
  }, {});

  // Convert Sets to counts
  $: affectedPagesCounts = Object.entries(affectedPagesByType).reduce((acc, [type, urlSet]) => {
    acc[type] = urlSet.size;
    return acc;
  }, {});

  // Calculate priority score for each issue: severity_weight * pages_affected
  const getSeverityWeight = (severity) => {
    switch (severity) {
      case 'error': return 10;
      case 'warning': return 5;
      case 'info': return 1;
      default: return 1;
    }
  };

  const calculatePriorityScore = (issue) => {
    const severityWeight = getSeverityWeight(issue.severity);
    const pagesAffected = affectedPagesCounts[issue.type] || 0;
    return severityWeight * pagesAffected;
  };

  // Calculate priority scores for all issues (for top 10 highlighting)
  $: issuesWithPriority = issues.map(issue => ({
    ...issue,
    priorityScore: calculatePriorityScore(issue)
  }));

  // Get top 10 priority issues (for highlighting) - create a Set of issue identifiers
  $: top10PriorityIssues = new Set(
    issuesWithPriority
      .sort((a, b) => b.priorityScore - a.priorityScore)
      .slice(0, 10)
      .map(issue => `${issue.url}|${issue.type}`) // Use URL + type as unique identifier
  );

  // Filter issues based on search, severity, type, affected pages count, and URL
  // Direct reactive statement - Svelte tracks all referenced variables automatically
  $: filteredIssues = issues.filter(i => 
    (severityFilter === 'all' || i.severity === severityFilter) &&
    (typeFilter === 'all' || i.type === typeFilter) &&
    (!filter.url || i.url === filter.url) &&
      (!searchTerm || 
        i.url.toLowerCase().includes(searchTerm.toLowerCase()) ||
        i.message?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        i.recommendation?.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  // Sort filtered issues by priority if selected
  $: sortedFilteredIssues = (() => {
    if (sortBy === 'priority') {
      return [...filteredIssues].sort((a, b) => {
        const scoreA = calculatePriorityScore(a);
        const scoreB = calculatePriorityScore(b);
        return scoreB - scoreA; // Descending order (highest priority first)
      });
    }
    return filteredIssues;
  })();

  // Group filtered issues
  $: groupedIssues = (() => {
    if (groupBy === 'none') {
      return { 'All Issues': sortedFilteredIssues };
    } else if (groupBy === 'url') {
      const grouped = {};
      sortedFilteredIssues.forEach(issue => {
        if (!grouped[issue.url]) {
          grouped[issue.url] = [];
        }
        grouped[issue.url].push(issue);
      });
      return grouped;
    } else if (groupBy === 'type') {
      const grouped = {};
      sortedFilteredIssues.forEach(issue => {
        if (!grouped[issue.type]) {
          grouped[issue.type] = [];
        }
        grouped[issue.type].push(issue);
      });
      return grouped;
    } else if (groupBy === 'severity') {
      const grouped = {};
      sortedFilteredIssues.forEach(issue => {
        if (!grouped[issue.severity]) {
          grouped[issue.severity] = [];
        }
        grouped[issue.severity].push(issue);
      });
      return grouped;
    }
    return {};
  })();

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
    // Include priority scores in export
    const issuesWithPriority = filteredIssues.map(issue => ({
      ...issue,
      priorityScore: calculatePriorityScore(issue)
    }));
    const content = JSON.stringify(issuesWithPriority, null, 2);
    downloadFile(content, fileName, 'application/json');
  };

  const toCsv = (rows) => {
    const escapeValue = (value) => {
      if (value === null || value === undefined) return '';
      const stringValue = String(value);
      return /[",\n]/.test(stringValue) ? `"${stringValue.replace(/"/g, '""')}"` : stringValue;
    };

    const headers = ['url', 'type', 'severity', 'priority_score', 'message', 'recommendation', 'value'];
    const dataRows = rows.map(issue => [
      issue.url || '',
      issue.type || '',
      issue.severity || '',
      calculatePriorityScore(issue).toString(),
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

    <div class="flex flex-col gap-4 mb-4">
      <!-- Search and grouping controls -->
      <div class="flex flex-col md:flex-row gap-4">
        <input
          type="text"
          placeholder="Search by URL, message, or recommendation..."
          class="input input-bordered flex-1"
          bind:value={searchTerm}
        />
        <div class="flex flex-wrap gap-2">
          <select class="select select-bordered" bind:value={groupBy}>
            <option value="none">No Grouping</option>
            <option value="url">Group by URL</option>
            <option value="type">Group by Type</option>
            <option value="severity">Group by Severity</option>
          </select>
          <select 
            class="select select-bordered" 
            bind:value={severityFilter}
          >
            <option value="all">All Severities</option>
            <option value="error">Errors</option>
            <option value="warning">Warnings</option>
            <option value="info">Info</option>
          </select>
          <select 
            class="select select-bordered" 
            bind:value={typeFilter}
          >
            <option value="all">All Types</option>
            {#each uniqueTypes as type}
              <option value={type}>{type.replace(/_/g, ' ')}</option>
            {/each}
          </select>
          <select class="select select-bordered" bind:value={sortBy}>
            <option value="none">Sort by...</option>
            <option value="priority">Sort by Priority</option>
          </select>
        </div>
      </div>
      
      <!-- Export button -->
      <div class="flex justify-end">
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
    </div>

    <div class="text-sm text-base-content/70 mb-4">
      Showing {filteredIssues.length} of {issues.length} issues
      {#if filter.url}
        <span class="badge badge-info ml-2">Filtered by URL: {filter.url}</span>
      {/if}
      {#if groupBy !== 'none'}
        | Grouped by {groupBy === 'url' ? 'URL' : groupBy === 'type' ? 'Type' : 'Severity'}
      {/if}
      {#if sortBy === 'priority'}
        | Sorted by Priority (highest first)
      {/if}
    </div>

    <div class="space-y-4">
      {#each Object.entries(groupedIssues) as [groupKey, groupIssues]}
        {#if groupBy !== 'none'}
          <div class="border-l-4 border-primary pl-4 mb-4">
            <h3 class="font-bold text-lg mb-2">
              {#if groupBy === 'url'}
                {groupKey}
              {:else if groupBy === 'severity'}
                {groupKey.charAt(0).toUpperCase() + groupKey.slice(1)} Issues
              {:else}
                {groupKey.replace(/_/g, ' ')}
                <span class="badge badge-primary ml-2">
                  {affectedPagesCounts[groupKey] || 0} page{affectedPagesCounts[groupKey] !== 1 ? 's' : ''}
                </span>
              {/if}
              <span class="badge badge-ghost ml-2">{groupIssues.length} issue{groupIssues.length !== 1 ? 's' : ''}</span>
            </h3>
          </div>
        {/if}
        
        {#each groupIssues as issue}
          {@const priorityScore = calculatePriorityScore(issue)}
          {@const issueId = `${issue.url}|${issue.type}`}
          {@const isTopPriority = top10PriorityIssues.has(issueId)}
          <div class="alert {getSeverityBadge(issue.severity)} shadow-lg {isTopPriority ? 'ring-2 ring-warning ring-offset-2' : ''}">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2 flex-wrap">
                <span class="badge {getSeverityBadge(issue.severity)}">
                  {issue.severity}
                </span>
                <span class="badge badge-outline">
                  {issue.type.replace(/_/g, ' ')}
                </span>
                <span class="badge badge-primary">
                  Priority: {priorityScore}
                </span>
                {#if isTopPriority}
                  <span class="badge badge-warning">
                    🔥 Top Priority
                  </span>
                {/if}
                {#if groupBy === 'url' && affectedPagesCounts[issue.type]}
                  <span class="badge badge-ghost">
                    {affectedPagesCounts[issue.type]} page{affectedPagesCounts[issue.type] !== 1 ? 's' : ''} affected
                  </span>
                {/if}
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
      {/each}
    </div>

    {#if filteredIssues.length === 0}
      <div class="alert alert-success">
        <span>🎉 No issues found!</span>
      </div>
    {/if}
  </div>
</div>

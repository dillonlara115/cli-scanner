<script>
  import PageDetailModal from './PageDetailModal.svelte';

  export let results = [];
  export let issues = [];
  export let filter = { status: 'all', performance: false };
  export let navigateToTab = null; // Function to navigate to tabs

  let searchTerm = '';
  let statusFilter = filter.status || 'all';
  let sortBy = 'url';
  let sortOrder = 'asc';
  let showSlowOnly = filter.performance || false;
  let showIssuesOnly = false;
  let selectedPage = null;
  let showModal = false;

  // Sync with prop changes
  $: if (filter.status) statusFilter = filter.status;
  $: if (filter.performance !== undefined) showSlowOnly = filter.performance;

  // Calculate issue counts per URL
  $: issuesByUrl = issues.reduce((acc, issue) => {
    if (!acc[issue.url]) {
      acc[issue.url] = [];
    }
    acc[issue.url].push(issue);
    return acc;
  }, {});

  $: issueCountsByUrl = Object.entries(issuesByUrl).reduce((acc, [url, urlIssues]) => {
    acc[url] = urlIssues.length;
    return acc;
  }, {});

  const openPageDetail = (page) => {
    selectedPage = page;
    showModal = true;
  };

  const closeModal = () => {
    showModal = false;
    selectedPage = null;
  };

  const viewIssuesForPage = (url) => {
    if (navigateToTab) {
      navigateToTab('issues', { url: url });
    }
    closeModal();
  };

  $: filteredResults = results.filter(r => {
    const matchesSearch = !searchTerm || 
      r.url.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.title?.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' ||
      (statusFilter === 'success' && r.status_code >= 200 && r.status_code < 300) ||
      (statusFilter === 'redirect' && r.status_code >= 300 && r.status_code < 400) ||
      (statusFilter === 'error' && r.status_code >= 400) ||
      (statusFilter === 'failed' && r.error);

    const matchesPerformance = !showSlowOnly || r.response_time_ms > 2000;

    const hasIssues = (issueCountsByUrl[r.url] || 0) > 0;
    const matchesIssuesFilter = !showIssuesOnly || hasIssues;

    return matchesSearch && matchesStatus && matchesPerformance && matchesIssuesFilter;
  }).sort((a, b) => {
    let aVal, bVal;
    
    switch (sortBy) {
      case 'url':
        aVal = a.url;
        bVal = b.url;
        break;
      case 'status':
        aVal = a.status_code;
        bVal = b.status_code;
        break;
      case 'time':
        aVal = a.response_time_ms;
        bVal = b.response_time_ms;
        break;
      case 'title':
        aVal = a.title || '';
        bVal = b.title || '';
        break;
      default:
        return 0;
    }

    if (typeof aVal === 'string') {
      return sortOrder === 'asc' 
        ? aVal.localeCompare(bVal)
        : bVal.localeCompare(aVal);
    }
    
    return sortOrder === 'asc' ? aVal - bVal : bVal - aVal;
  });

  const getStatusBadge = (status) => {
    if (status >= 200 && status < 300) return 'badge-success';
    if (status >= 300 && status < 400) return 'badge-warning';
    if (status >= 400) return 'badge-error';
    return 'badge-ghost';
  };
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <div class="flex flex-col md:flex-row gap-4 mb-4">
      <input
        type="text"
        placeholder="Search by URL or title..."
        class="input input-bordered flex-1"
        bind:value={searchTerm}
      />
      <select class="select select-bordered" bind:value={statusFilter}>
        <option value="all">All Status</option>
        <option value="success">Success (2xx)</option>
        <option value="redirect">Redirect (3xx)</option>
        <option value="error">Error (4xx/5xx)</option>
        <option value="failed">Failed</option>
      </select>
      <label class="label cursor-pointer gap-2">
        <span class="label-text">Slow only (>2s)</span>
        <input type="checkbox" class="checkbox checkbox-primary" bind:checked={showSlowOnly} />
      </label>
      <label class="label cursor-pointer gap-2">
        <span class="label-text">With issues only</span>
        <input type="checkbox" class="checkbox checkbox-primary" bind:checked={showIssuesOnly} />
      </label>
      <select class="select select-bordered" bind:value={sortBy}>
        <option value="url">Sort by URL</option>
        <option value="status">Sort by Status</option>
        <option value="time">Sort by Response Time</option>
        <option value="title">Sort by Title</option>
      </select>
      <button
        class="btn btn-outline"
        on:click={() => sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'}
      >
        {sortOrder === 'asc' ? '↑' : '↓'}
      </button>
    </div>

    <div class="text-sm text-base-content/70 mb-4">
      Showing {filteredResults.length} of {results.length} pages
    </div>

    <div class="overflow-x-auto">
      <table class="table table-zebra">
        <thead>
          <tr>
            <th>URL</th>
            <th>Status</th>
            <th>Response Time</th>
            <th>Title</th>
            <th>Issues</th>
            <th>Links</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredResults as result}
            {@const issueCount = issueCountsByUrl[result.url] || 0}
            <tr class="cursor-pointer hover:bg-base-200" on:click={() => openPageDetail(result)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && openPageDetail(result)}>
              <td>
                <a href={result.url} target="_blank" class="link link-primary" onclick={(e) => e.stopPropagation()}>
                  {result.url}
                </a>
              </td>
              <td>
                <span class="badge {getStatusBadge(result.status_code)}">
                  {result.status_code}
                </span>
              </td>
              <td>{result.response_time_ms}ms</td>
              <td class="max-w-xs truncate">{result.title || '-'}</td>
              <td>
                {#if issueCount > 0}
                  <span class="badge badge-error">
                    {issueCount} issue{issueCount !== 1 ? 's' : ''}
                  </span>
                {:else}
                  <span class="badge badge-success">No issues</span>
                {/if}
              </td>
              <td onclick={(e) => e.stopPropagation()}>
                <span class="badge badge-ghost">
                  {result.internal_links?.length || 0} internal
                </span>
                <span class="badge badge-ghost">
                  {result.external_links?.length || 0} external
                </span>
              </td>
              <td onclick={(e) => e.stopPropagation()}>
                <button 
                  class="btn btn-sm btn-primary"
                  on:click={(e) => { e.stopPropagation(); viewIssuesForPage(result.url); }}
                >
                  View Issues
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>

  {#if showModal && selectedPage}
    <PageDetailModal 
      page={selectedPage} 
      issues={issuesByUrl[selectedPage.url] || []}
      on:close={closeModal}
      {navigateToTab}
    />
  {/if}
</div>


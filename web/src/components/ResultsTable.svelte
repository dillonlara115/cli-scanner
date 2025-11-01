<script>
  import { onMount } from 'svelte';

  export let results = [];

  let searchTerm = '';
  let statusFilter = 'all';
  let sortBy = 'url';
  let sortOrder = 'asc';

  $: filteredResults = results.filter(r => {
    const matchesSearch = !searchTerm || 
      r.url.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.title?.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' ||
      (statusFilter === 'success' && r.status_code >= 200 && r.status_code < 300) ||
      (statusFilter === 'redirect' && r.status_code >= 300 && r.status_code < 400) ||
      (statusFilter === 'error' && r.status_code >= 400) ||
      (statusFilter === 'failed' && r.error);

    return matchesSearch && matchesStatus;
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
            <th>Links</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredResults as result}
            <tr>
              <td>
                <a href={result.url} target="_blank" class="link link-primary">
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
                <span class="badge badge-ghost">
                  {result.internal_links?.length || 0} internal
                </span>
                <span class="badge badge-ghost">
                  {result.external_links?.length || 0} external
                </span>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>


<script>
  import { createEventDispatcher } from 'svelte';

  export let page = null;
  export let issues = [];
  export let navigateToTab = null;

  const dispatch = createEventDispatcher();

  const close = () => {
    dispatch('close');
  };

  const handleViewIssues = () => {
    if (navigateToTab) {
      navigateToTab('issues', { url: page.url });
    }
    close();
  };

  const getSeverityBadge = (severity) => {
    switch (severity) {
      case 'error': return 'badge-error';
      case 'warning': return 'badge-warning';
      case 'info': return 'badge-info';
      default: return 'badge-ghost';
    }
  };

  const formatHeadingList = (headings) => {
    if (!headings || headings.length === 0) return 'None';
    return headings.join(', ');
  };
</script>

{#if page}
  <!-- Modal backdrop -->
  <div class="modal modal-open">
    <div class="modal-box max-w-4xl max-h-[90vh] overflow-y-auto">
      <h3 class="font-bold text-2xl mb-4">Page Details</h3>
      
      <!-- Page URL -->
      <div class="mb-4">
        <div class="font-semibold mb-2">URL:</div>
        <a href={page.url} target="_blank" class="link link-primary break-all">
          {page.url}
        </a>
      </div>

      <!-- Page Status & Performance -->
      <div class="grid grid-cols-2 gap-4 mb-4">
        <div>
          <div class="font-semibold mb-2">Status Code:</div>
          <span class="badge {page.status_code >= 200 && page.status_code < 300 ? 'badge-success' : page.status_code >= 400 ? 'badge-error' : 'badge-warning'}">
            {page.status_code}
          </span>
        </div>
        <div>
          <div class="font-semibold mb-2">Response Time:</div>
          <div>{page.response_time_ms}ms</div>
        </div>
      </div>

      <!-- Page Metadata -->
      <div class="mb-4">
        <div class="font-semibold mb-2">Title:</div>
        <div class="break-words">
          {#if page.title}
            {page.title}
          {:else}
            <span class="text-base-content/50">Not set</span>
          {/if}
        </div>
      </div>

      <div class="mb-4">
        <div class="font-semibold mb-2">Meta Description:</div>
        <div class="break-words">
          {#if page.meta_description}
            {page.meta_description}
          {:else}
            <span class="text-base-content/50">Not set</span>
          {/if}
        </div>
      </div>

      {#if page.canonical}
        <div class="mb-4">
          <div class="font-semibold mb-2">Canonical URL:</div>
          <a href={page.canonical} target="_blank" class="link link-primary break-all">
            {page.canonical}
          </a>
        </div>
      {/if}

      <!-- Headings -->
      <div class="mb-4">
        <div class="font-semibold mb-2">Headings:</div>
        <div class="space-y-2">
          {#if page.h1 && page.h1.length > 0}
            <div><span class="font-semibold">H1:</span> {formatHeadingList(page.h1)}</div>
          {/if}
          {#if page.h2 && page.h2.length > 0}
            <div><span class="font-semibold">H2:</span> {formatHeadingList(page.h2)}</div>
          {/if}
          {#if page.h3 && page.h3.length > 0}
            <div><span class="font-semibold">H3:</span> {formatHeadingList(page.h3)}</div>
          {/if}
          {#if (!page.h1 || page.h1.length === 0) && (!page.h2 || page.h2.length === 0) && (!page.h3 || page.h3.length === 0)}
            <div class="text-base-content/50">No headings found</div>
          {/if}
        </div>
      </div>

      <!-- Links -->
      <div class="grid grid-cols-2 gap-4 mb-4">
        <div>
          <div class="font-semibold mb-2">Internal Links:</div>
          <div class="badge badge-ghost">{page.internal_links?.length || 0}</div>
        </div>
        <div>
          <div class="font-semibold mb-2">External Links:</div>
          <div class="badge badge-ghost">{page.external_links?.length || 0}</div>
        </div>
      </div>

      <!-- Images -->
      {#if page.images && page.images.length > 0}
        <div class="mb-4">
          <div class="font-semibold mb-2">Images ({page.images.length}):</div>
          <div class="max-h-32 overflow-y-auto">
            {#each page.images.slice(0, 10) as image}
              <div class="text-sm break-all">
                {image.url}
                {#if image.alt}
                  <span class="text-base-content/70">(alt: {image.alt})</span>
                {:else}
                  <span class="badge badge-warning badge-sm">No alt</span>
                {/if}
              </div>
            {/each}
            {#if page.images.length > 10}
              <div class="text-sm text-base-content/70">... and {page.images.length - 10} more</div>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Redirect Chain -->
      {#if page.redirect_chain && page.redirect_chain.length > 0}
        <div class="mb-4">
          <div class="font-semibold mb-2">Redirect Chain:</div>
          <div class="break-all">{page.redirect_chain.join(' → ')}</div>
        </div>
      {/if}

      <!-- Error -->
      {#if page.error}
        <div class="mb-4">
          <div class="font-semibold mb-2 text-error">Error:</div>
          <div class="text-error">{page.error}</div>
        </div>
      {/if}

      <!-- Issues Section -->
      <div class="divider"></div>
      <div class="mb-4">
        <div class="flex items-center justify-between mb-4">
          <h4 class="font-bold text-xl">Issues ({issues.length})</h4>
          {#if issues.length > 0}
            <button class="btn btn-primary btn-sm" on:click={handleViewIssues}>
              View All Issues
            </button>
          {/if}
        </div>

        {#if issues.length === 0}
          <div class="alert alert-success">
            <span>🎉 No issues found for this page!</span>
          </div>
        {:else}
          <div class="space-y-2">
            {#each issues as issue}
              <div class="alert {getSeverityBadge(issue.severity)} shadow-sm">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-1">
                    <span class="badge {getSeverityBadge(issue.severity)}">
                      {issue.severity}
                    </span>
                    <span class="badge badge-outline">
                      {issue.type.replace(/_/g, ' ')}
                    </span>
                  </div>
                  <h5 class="font-semibold">{issue.message}</h5>
                  {#if issue.value}
                    <div class="text-sm mt-1">
                      <span class="font-semibold">Value:</span> {issue.value}
                    </div>
                  {/if}
                  {#if issue.recommendation}
                    <div class="text-sm mt-1">
                      <span class="font-semibold">Recommendation:</span> {issue.recommendation}
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Modal Actions -->
      <div class="modal-action">
        <button class="btn" on:click={close}>Close</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button type="button" on:click={close}>close</button>
    </form>
  </div>
{/if}

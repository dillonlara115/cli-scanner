<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { triggerCrawl, fetchProjects } from '../lib/data.js';
  
  import CrawlProgress from './CrawlProgress.svelte';
  
  export let projectId = null;
  export let project = null;
  export let className = '';
  
  const dispatch = createEventDispatcher();
  
  let showModal = false;
  let showProgress = false;
  let activeCrawlId = null;
  let loading = false;
  let error = null;
  let loadedProject = null;
  let hasUrl = false;
  
  // Form fields
  let url = '';
  let maxDepth = 3;
  let maxPages = 1000;
  let workers = 10;
  let respectRobots = true;
  let parseSitemap = false;

  // Use loadedProject or prop project
  $: currentProject = project || loadedProject;
  
  // Always sync URL from project settings when project changes
  $: if (currentProject?.settings?.url) {
    url = currentProject.settings.url;
  }
  
  // Reactive check for URL availability
  $: {
    // Check multiple possible locations for URL
    let projectUrl = null;
    
    if (currentProject) {
      // Try settings.url first
      if (currentProject.settings?.url) {
        projectUrl = currentProject.settings.url;
      }
      // If no URL in settings, construct from domain
      else if (currentProject.domain) {
        // Construct URL from domain (assume https)
        const domain = currentProject.domain.trim();
        projectUrl = domain.startsWith('http') ? domain : `https://${domain}`;
      }
    }
    
    const urlVar = url;
    const hasUrlResult = projectUrl || urlVar;
    
    // Debug logging
    console.log('URL check:', { 
      currentProject: !!currentProject,
      projectName: currentProject?.name,
      projectDomain: currentProject?.domain,
      settingsType: typeof currentProject?.settings,
      settings: currentProject?.settings,
      projectUrl, 
      urlVar, 
      hasUrlResult
    });
    
    hasUrl = hasUrlResult;
    
    // Auto-populate URL if we have it from project but url variable is empty
    if (projectUrl && !urlVar) {
      url = projectUrl;
    }
  }

  // Load project if not provided
  onMount(async () => {
    if (!project && projectId) {
      const { data: projects } = await fetchProjects();
      if (projects) {
        loadedProject = projects.find(p => p.id === projectId);
      }
    }
  });

  // Update URL when modal opens
  async function openModal() {
    // Ensure project is loaded
    if (!currentProject && projectId) {
      const { data: projects } = await fetchProjects();
      if (projects) {
        loadedProject = projects.find(p => p.id === projectId);
      }
    }
    
    if (currentProject?.settings?.url) {
      url = currentProject.settings.url;
    }
    showModal = true;
  }

  async function handleSubmit() {
    // Get URL from project settings or use empty string (will be validated on backend)
    const crawlUrl = currentProject?.settings?.url || url;
    
    if (!crawlUrl) {
      error = 'Project URL is required. Please set a URL in project settings.';
      return;
    }

    // Basic URL validation
    try {
      new URL(crawlUrl);
    } catch (e) {
      error = 'Invalid URL format in project settings';
      return;
    }

    loading = true;
    error = null;

    const { data, error: crawlError } = await triggerCrawl(projectId, {
      url: crawlUrl,
      max_depth: maxDepth,
      max_pages: maxPages,
      workers,
      respect_robots: respectRobots,
      parse_sitemap: parseSitemap
    });

    loading = false;

    if (crawlError) {
      error = crawlError.message || 'Failed to trigger crawl';
      return;
    }

    // Success - show progress instead of closing modal
    console.log('Crawl triggered successfully:', data);
    
    // The API returns { crawl_id, status, message }
    const crawlId = data.crawl_id || data.id || data.crawlId;
    console.log('Extracted crawlId:', crawlId, 'from data:', data);
    
    if (!crawlId) {
      console.error('No crawl_id found in response:', data);
      error = 'Failed to get crawl ID from response';
      return;
    }
    
    activeCrawlId = crawlId;
    showProgress = true;
    console.log('Switching to progress view. activeCrawlId:', activeCrawlId, 'showProgress:', showProgress);
    
    // Don't dispatch 'created' event immediately - let the progress component handle completion
    // The parent will redirect when crawl completes via the 'completed' event
    
    // Reset form (but keep URL from project settings)
    if (currentProject?.settings?.url) {
      url = currentProject.settings.url;
    } else {
      url = '';
    }
    maxDepth = 3;
    maxPages = 1000;
    workers = 10;
    respectRobots = true;
    parseSitemap = false;
  }

  function handleCancel() {
    showModal = false;
    showProgress = false;
    activeCrawlId = null;
    error = null;
    // Reset form (but keep URL from project settings)
    if (currentProject?.settings?.url) {
      url = currentProject.settings.url;
    } else {
      url = '';
    }
    maxDepth = 3;
    maxPages = 1000;
    workers = 10;
    respectRobots = true;
    parseSitemap = false;
  }
  
  function handleProgressComplete() {
    // When crawl completes, dispatch both events
    if (projectId && activeCrawlId) {
      dispatch('created', { crawl_id: activeCrawlId });
      dispatch('completed', { crawlId: activeCrawlId });
      // Don't close modal yet - let parent handle navigation
    }
  }
</script>

<button 
  class="btn btn-primary {className}"
  on:click={openModal}
>
  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
  </svg>
  Start Crawl
</button>

{#if showModal}
  <dialog class="modal modal-open">
    <div class="modal-box bg-base-200 text-base-content max-w-2xl overflow-visible">
      {#if showProgress && activeCrawlId}
        <!-- Show progress -->
        <div class="mb-4 flex justify-between items-center">
          <h3 class="font-bold text-lg">Crawl Progress</h3>
          <button 
            class="btn btn-sm btn-ghost"
            on:click={handleCancel}
          >
            ✕
          </button>
        </div>
        <!-- Debug info -->
        <div class="mb-2 text-xs text-base-content opacity-50">
          Crawl ID: {activeCrawlId} | Project ID: {projectId}
        </div>
        <CrawlProgress crawlId={activeCrawlId} {projectId} on:completed={(e) => {
          console.log('CrawlProgress completed event:', e.detail);
          handleProgressComplete();
        }} />
      {:else}
        <!-- Show form -->
        <h3 class="font-bold text-lg mb-4">Start New Crawl for {currentProject?.name} ({currentProject?.domain})</h3>
        
        {#if hasUrl}
          <div class="alert alert-info mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            <span>Crawling: <strong>{url || (currentProject?.domain ? `https://${currentProject.domain}` : '')}</strong></span>
          </div>
        {:else if currentProject}
          <div class="alert alert-warning mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <span>No URL configured for this project. Please update project settings.</span>
          </div>
        {/if}
        
        <!-- Hidden field to ensure URL is available for form submission -->
        {#if url}
          <input type="hidden" name="url" value={url} />
        {/if}
        
        {#if error}
          <div class="alert alert-error mb-4">
            <span>{error}</span>
          </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit}>

        <div class="grid grid-cols-2 gap-4 mb-4">
          <div class="form-control">
            <label class="label" for="max-depth-input">
              <span class="label-text text-base-content">Max Depth</span>
            </label>
            <input 
              id="max-depth-input"
              type="number" 
              class="input input-bordered bg-base-100 text-base-content" 
              min="1"
              max="10"
              bind:value={maxDepth}
              disabled={loading}
            />
            <div class="label">
              <span class="label-text-alt text-base-content opacity-70">Maximum number of link levels to follow from the starting URL. Higher values crawl deeper but take longer.</span>
            </div>
          </div>

          <div class="form-control">
            <label class="label" for="max-pages-input">
              <span class="label-text text-base-content">Max Pages</span>
            </label>
            <input 
              id="max-pages-input"
              type="number" 
              class="input input-bordered bg-base-100 text-base-content" 
              min="1"
              max="10000"
              bind:value={maxPages}
              disabled={loading}
            />
            <div class="label">
              <span class="label-text-alt text-base-content opacity-70">Maximum number of pages to crawl. The crawl will stop once this limit is reached.</span>
            </div>
          </div>
        </div>

        <div class="form-control mb-4">
          <label class="label" for="workers-input">
            <span class="label-text text-base-content">Workers</span>
          </label>
          <input 
            id="workers-input"
            type="number" 
            class="input input-bordered bg-base-100 text-base-content" 
            min="1"
            max="50"
            bind:value={workers}
            disabled={loading}
          />
          <div class="label">
            <span class="label-text-alt text-base-content opacity-70">Number of concurrent workers. Higher values crawl faster but use more resources and may be blocked by servers.</span>
          </div>
        </div>

        <div class="form-control mb-4">
          <label class="label cursor-pointer">
            <span class="label-text text-base-content">Respect robots.txt</span>
            <input 
              type="checkbox" 
              class="toggle toggle-primary" 
              bind:checked={respectRobots}
              disabled={loading}
            />
          </label>
          <div class="label">
            <span class="label-text-alt text-base-content opacity-70">If enabled, the crawler will check robots.txt and skip URLs that are disallowed. Recommended for ethical crawling.</span>
          </div>
        </div>

        <div class="form-control mb-4">
          <label class="label cursor-pointer">
            <span class="label-text text-base-content">Parse sitemap.xml</span>
            <input 
              type="checkbox" 
              class="toggle toggle-primary" 
              bind:checked={parseSitemap}
              disabled={loading}
            />
          </label>
          <div class="label">
            <span class="label-text-alt text-base-content opacity-70">If enabled, the crawler will discover and use URLs from sitemap.xml files. This can help find more pages quickly.</span>
          </div>
        </div>

        <div class="modal-action">
          <button 
            type="button" 
            class="btn btn-ghost" 
            on:click={handleCancel}
            disabled={loading}
          >
            Cancel
          </button>
          <button 
            type="submit" 
            class="btn btn-primary"
            disabled={loading || !hasUrl}
          >
            {#if loading}
              <span class="loading loading-spinner loading-sm"></span>
              Starting...
            {:else}
              Start Crawl
            {/if}
          </button>
        </div>
        </form>
      {/if}
    </div>
    <form method="dialog" class="modal-backdrop">
      <button on:click={handleCancel}>close</button>
    </form>
  </dialog>
{/if}


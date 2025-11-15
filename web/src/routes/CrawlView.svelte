<script>
  import { onMount } from 'svelte';
  import { params, querystring, push } from 'svelte-spa-router';
  import { fetchProjects, fetchCrawls, fetchPages, fetchIssues } from '../lib/data.js';
  import ProjectsView from '../components/ProjectsView.svelte';
  import Dashboard from '../components/Dashboard.svelte';
  import CrawlSelector from '../components/CrawlSelector.svelte';
  import TriggerCrawlButton from '../components/TriggerCrawlButton.svelte';
  
  let projects = [];
  let project = null;
  let selectedProject = null;
  let crawls = [];
  let selectedCrawl = null;
  let summary = null;
  let results = [];
  let loading = true;
  let error = null;
  let currentProjectId = null;
  let currentCrawlId = null;

  $: projectId = $params?.projectId || null;
  $: crawlId = $params?.crawlId || null;
  $: tab = $querystring ? new URLSearchParams($querystring).get('tab') || 'dashboard' : 'dashboard';

  onMount(async () => {
    // Wait a tick for params to be available
    await new Promise(resolve => setTimeout(resolve, 0));
    if ($params?.projectId && $params?.crawlId) {
      await loadData();
    }
  });

  $: if (projectId && crawlId && 
        (projectId !== currentProjectId || crawlId !== currentCrawlId) && 
        $params?.projectId && $params?.crawlId) {
    currentProjectId = projectId;
    currentCrawlId = crawlId;
    loadData();
  }

  async function loadData() {
    if (!projectId || !crawlId) return;
    
    loading = true;
    try {
      // Load projects
      const { data: projectsData, error: projectsError } = await fetchProjects();
      if (projectsError) throw projectsError;
      projects = projectsData || [];
      
      // Find current project
      project = projects.find(p => p.id === projectId);
      selectedProject = project;
      if (!project) {
        error = 'Project not found';
        loading = false;
        return;
      }

      // Load crawls for this project
      const { data: crawlsData, error: crawlsError } = await fetchCrawls(projectId);
      if (crawlsError) throw crawlsError;
      crawls = crawlsData || [];

      // Find selected crawl
      selectedCrawl = crawls.find(c => c.id === crawlId);
      if (!selectedCrawl) {
        error = 'Crawl not found';
        loading = false;
        return;
      }

      // Load crawl data
      await loadCrawlData(crawlId);
    } catch (err) {
      error = err.message;
      loading = false;
    }
  }

  async function loadCrawlData(crawlId) {
    if (!crawlId) return;

    try {
      // Fetch pages and issues in parallel
      const [pagesResult, issuesResult] = await Promise.all([
        fetchPages(crawlId),
        fetchIssues(crawlId)
      ]);

      if (pagesResult.error) throw pagesResult.error;
      if (issuesResult.error) throw issuesResult.error;

      results = pagesResult.data || [];
      const issues = issuesResult.data || [];

      // Calculate link statistics from pages
      let totalInternalLinks = 0;
      let totalExternalLinks = 0;
      let pagesWithRedirects = 0;
      
      results.forEach(page => {
        // Count internal links
        if (Array.isArray(page.internal_links)) {
          totalInternalLinks += page.internal_links.length;
        }
        // Count external links
        if (Array.isArray(page.external_links)) {
          totalExternalLinks += page.external_links.length;
        }
        // Count pages with redirects
        if (Array.isArray(page.redirect_chain) && page.redirect_chain.length > 0) {
          pagesWithRedirects++;
        }
      });

      // Generate summary from data
      summary = {
        total_pages: results.length,
        total_issues: issues.length,
        issues_by_type: {},
        issues: issues.map(issue => ({
          type: issue.type,
          severity: issue.severity,
          url: issue.page_id ? results.find(p => p.id === issue.page_id)?.url || '' : '',
          message: issue.message,
          value: issue.value,
          recommendation: issue.recommendation
        })),
        average_response_time_ms: results.length > 0
          ? Math.round(results.reduce((sum, p) => sum + (p.response_time_ms || 0), 0) / results.length)
          : 0,
        pages_with_errors: results.filter(p => p.status_code >= 400).length,
        total_internal_links: totalInternalLinks,
        total_external_links: totalExternalLinks,
        pages_with_redirects: pagesWithRedirects
      };

      // Count issues by type
      issues.forEach(issue => {
        summary.issues_by_type[issue.type] = (summary.issues_by_type[issue.type] || 0) + 1;
      });

      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
    }
  }

  function handleProjectSelect(selectedProject) {
    push(`/project/${selectedProject.id}`);
  }

  function handleCrawlSelect(crawl) {
    push(`/project/${projectId}/crawl/${crawl.id}?tab=${tab}`);
  }

  async function handleCrawlCreated(e) {
    // Reload crawls and redirect to the new crawl
    const { data: crawlsData, error: crawlsError } = await fetchCrawls(projectId);
    if (!crawlsError && crawlsData && crawlsData.length > 0) {
      crawls = crawlsData;
      // Find the new crawl (should be first/latest)
      const newCrawl = crawlsData.find(c => c.id === e.detail.crawl_id) || crawlsData[0];
      push(`/project/${projectId}/crawl/${newCrawl.id}?tab=${tab}`);
    }
  }
</script>

{#if loading}
  <div class="flex items-center justify-center min-h-screen">
    <span class="loading loading-spinner loading-lg"></span>
  </div>
{:else if error}
  <div class="flex items-center justify-center min-h-screen">
    <div class="alert alert-error max-w-md">
      <span>Error: {error}</span>
    </div>
  </div>
{:else if project && selectedCrawl}
  <ProjectsView {projects} {selectedProject} on:select={(e) => handleProjectSelect(e.detail)} />
  
  <div class="container mx-auto p-4">
    <div class="flex justify-between items-center mb-4">
      <CrawlSelector {crawls} selectedCrawl={selectedCrawl} on:select={(e) => handleCrawlSelect(e.detail)} />
      <TriggerCrawlButton {projectId} project={project} on:created={handleCrawlCreated} />
    </div>
  </div>
  
  <Dashboard {summary} {results} initialTab={tab} projectId={projectId} crawlId={crawlId} project={project} />
{/if}


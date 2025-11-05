<script>
  import { onMount } from 'svelte';
  import { push, params } from 'svelte-spa-router';
  import { fetchProjects, fetchCrawls } from '../lib/data.js';
  import ProjectsView from '../components/ProjectsView.svelte';
  import CrawlSelector from '../components/CrawlSelector.svelte';
  import TriggerCrawlButton from '../components/TriggerCrawlButton.svelte';
  
  let projects = [];
  let project = null;
  let selectedProject = null;
  let crawls = [];
  let loading = true;
  let error = null;
  let currentProjectId = null;

  $: projectId = $params?.id || null;

  onMount(async () => {
    // Wait a tick for params to be available
    await new Promise(resolve => setTimeout(resolve, 0));
    if ($params?.id) {
      await loadData();
    }
  });

  $: if (projectId && projectId !== currentProjectId && $params?.id) {
    currentProjectId = projectId;
    loadData();
  }

  async function loadData() {
    if (!projectId) return;
    
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

      // If crawls exist, redirect to latest crawl
      if (crawls.length > 0) {
        push(`/project/${projectId}/crawl/${crawls[0].id}`);
      }
    } catch (err) {
      error = err.message;
      loading = false;
    } finally {
      loading = false;
    }
  }

  function handleProjectSelect(selectedProject) {
    push(`/project/${selectedProject.id}`);
  }

  function handleCrawlSelect(crawl) {
    push(`/project/${projectId}/crawl/${crawl.id}`);
  }

  async function handleCrawlCreated(e) {
    // Reload crawls and redirect to the new crawl
    const { data: crawlsData, error: crawlsError } = await fetchCrawls(projectId);
    if (!crawlsError && crawlsData && crawlsData.length > 0) {
      crawls = crawlsData;
      // Find the new crawl (should be first/latest)
      const newCrawl = crawlsData.find(c => c.id === e.detail.crawl_id) || crawlsData[0];
      push(`/project/${projectId}/crawl/${newCrawl.id}`);
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
{:else if project}
  <ProjectsView {projects} {selectedProject} on:select={(e) => handleProjectSelect(e.detail)} />
  
  <div class="container mx-auto p-4">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold text-base-content">Crawls</h2>
      <TriggerCrawlButton {projectId} project={project} on:created={handleCrawlCreated} />
    </div>
    
    {#if crawls.length === 0}
      <div class="alert alert-info">
        <span>No crawls found for this project. Start a crawl to get started.</span>
      </div>
    {:else}
      <CrawlSelector {crawls} on:select={(e) => handleCrawlSelect(e.detail)} />
    {/if}
  </div>
{/if}


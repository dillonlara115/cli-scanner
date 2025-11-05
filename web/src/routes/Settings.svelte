<script>
  import { onMount } from 'svelte';
  import { params, push } from 'svelte-spa-router';
  import { fetchProjects } from '../lib/data.js';
  import ProjectGSCSelector from '../components/ProjectGSCSelector.svelte';
  
  let project = null;
  let summary = null; // For enriching issues if needed
  let loading = true;
  let error = null;

  $: projectId = $params?.projectId || null;

  onMount(async () => {
    if (projectId) {
      await loadProject();
    }
  });

  $: if (projectId) {
    loadProject();
  }

  async function loadProject() {
    if (!projectId) return;
    
    loading = true;
    try {
      const { data: projects, error: projectsError } = await fetchProjects();
      if (projectsError) throw projectsError;
      
      project = projects?.find(p => p.id === projectId);
      if (!project) {
        error = 'Project not found';
      }
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function handleEnriched(e) {
    // Navigate to issues tab to see enriched data
    push(`/project/${projectId}/crawl/${$params.crawlId || ''}?tab=issues`);
  }
</script>

<div class="container mx-auto p-6 max-w-4xl">
  <div class="mb-6">
    <button 
      class="btn btn-ghost btn-sm mb-4"
      on:click={() => push(`/project/${projectId}`)}
    >
      ← Back to Project
    </button>
    <h1 class="text-3xl font-bold mb-2">Project Settings</h1>
    <p class="text-base-content/70">
      Configure integration settings and preferences for this project.
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center min-h-[400px]">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else if error}
    <div class="alert alert-error">
      <span>{error}</span>
    </div>
  {:else if project}
    <div class="space-y-6">
      <!-- Google Search Console Integration -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title text-xl mb-4">Google Search Console Integration</h2>
          <ProjectGSCSelector {project} projectId={project.id} {summary} on:enriched={handleEnriched} />
        </div>
      </div>

      <!-- Future integrations can be added here -->
      <!-- 
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title text-xl mb-4">Other Integration</h2>
          <p>Integration settings here...</p>
        </div>
      </div>
      -->
    </div>
  {/if}
</div>


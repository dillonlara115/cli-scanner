<script>
  import { createEventDispatcher } from 'svelte';
  import { link } from 'svelte-spa-router';
  import { createProject } from '../lib/data.js';
  import Auth from './Auth.svelte';
  import Logo from './Logo.svelte';
  
  const dispatch = createEventDispatcher();

  export let projects = [];
  export let selectedProject = null;

  let showCreateModal = false;
  let newProjectName = '';
  let newProjectDomain = '';
  let creating = false;
  let error = null;

  function selectProject(project) {
    selectedProject = project;
    dispatch('select', project);
  }

  async function handleCreateProject() {
    if (!newProjectName || !newProjectDomain) {
      error = 'Name and domain are required';
      return;
    }

    creating = true;
    error = null;

    try {
      const { data, error: createError } = await createProject(
        newProjectName,
        newProjectDomain
      );

      if (createError) throw createError;

      // Add to projects list and select it
      projects = [...projects, data];
      selectProject(data);
      
      // Reset form
      newProjectName = '';
      newProjectDomain = '';
      showCreateModal = false;
    } catch (err) {
      error = err.message || 'Failed to create project';
    } finally {
      creating = false;
    }
  }
</script>

<div class="navbar bg-base-100 shadow-lg border-b border-base-300">
  <div class="flex-1">
    <a href="/" use:link class="btn btn-ghost">
      <Logo size="md" />
    </a>
  </div>
  <div class="flex-none gap-2">
    <Auth />
    <div class="dropdown dropdown-end">
      <label tabindex="0" class="btn btn-ghost">
        {selectedProject ? selectedProject.name : 'Select Project'}
        <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </label>
      <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-64 p-2 shadow-lg">
        {#each projects as project}
          <li>
            <a
              class:active={selectedProject?.id === project.id}
              on:click={() => selectProject(project)}
            >
              <div>
                <div class="font-semibold">{project.name}</div>
                <div class="text-sm opacity-70">{project.domain}</div>
              </div>
            </a>
          </li>
        {/each}
        <li>
          <a on:click={() => showCreateModal = true}>
            <span>+ Create New Project</span>
          </a>
        </li>
      </ul>
    </div>
  </div>
</div>

<!-- Create Project Modal -->
{#if showCreateModal}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">Create New Project</h3>

      {#if error}
        <div class="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      {/if}

      <div class="form-control w-full mb-4">
        <label class="label">
          <span class="label-text">Project Name</span>
        </label>
        <input
          type="text"
          placeholder="My Website"
          class="input input-bordered w-full"
          bind:value={newProjectName}
        />
      </div>

      <div class="form-control w-full mb-4">
        <label class="label">
          <span class="label-text">Domain</span>
        </label>
        <input
          type="text"
          placeholder="example.com"
          class="input input-bordered w-full"
          bind:value={newProjectDomain}
        />
      </div>

      <div class="modal-action">
        <button
          class="btn btn-ghost"
          on:click={() => {
            showCreateModal = false;
            error = null;
          }}
        >
          Cancel
        </button>
        <button
          class="btn btn-primary"
          on:click={handleCreateProject}
          disabled={creating || !newProjectName || !newProjectDomain}
        >
          {#if creating}
            <span class="loading loading-spinner loading-sm"></span>
          {:else}
            Create
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}


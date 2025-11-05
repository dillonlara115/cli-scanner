<script>
  import { createEventDispatcher } from 'svelte';
  import { createProject } from '../lib/data.js';
  
  const dispatch = createEventDispatcher();
  
  export let className = '';
  
  let showCreateModal = false;
  let newProjectName = '';
  let newProjectDomain = '';
  let newProjectUrl = '';
  let creating = false;
  let error = null;

  async function handleCreateProject() {
    if (!newProjectName || !newProjectDomain || !newProjectUrl) {
      error = 'Name, domain, and URL are required';
      return;
    }

    // Validate URL format
    try {
      new URL(newProjectUrl);
    } catch (e) {
      error = 'Invalid URL format';
      return;
    }

    creating = true;
    error = null;

    try {
      const { data, error: createError } = await createProject(
        newProjectName,
        newProjectDomain,
        { url: newProjectUrl }
      );

      if (createError) throw createError;

      // Emit the created project to parent
      dispatch('created', data);
      
      // Reset form and close modal
      newProjectName = '';
      newProjectDomain = '';
      newProjectUrl = '';
      showCreateModal = false;
    } catch (err) {
      error = err.message || 'Failed to create project';
    } finally {
      creating = false;
    }
  }
</script>

<button 
  class="btn btn-primary {className}"
  on:click={() => showCreateModal = true}
>
  Create Project
</button>

<!-- Create Project Modal -->
{#if showCreateModal}
  <div class="modal modal-open">
    <div class="modal-box bg-base-100">
      <h3 class="font-bold text-lg mb-4 text-base-content">Create New Project</h3>

      {#if error}
        <div class="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      {/if}

      <div class="form-control w-full mb-4">
        <label class="label">
          <span class="label-text text-base-content">Project Name</span>
        </label>
        <input
          type="text"
          placeholder="My Website"
          class="input input-bordered w-full bg-base-200 text-base-content placeholder-gray-500 border-base-300 focus:border-primary"
          bind:value={newProjectName}
        />
      </div>

      <div class="form-control w-full mb-4">
        <label class="label">
          <span class="label-text text-base-content">Domain</span>
        </label>
        <input
          type="text"
          placeholder="example.com"
          class="input input-bordered w-full bg-base-200 text-base-content placeholder-gray-500 border-base-300 focus:border-primary"
          bind:value={newProjectDomain}
        />
      </div>

      <div class="form-control w-full mb-4">
        <label class="label">
          <span class="label-text text-base-content">Starting URL</span>
        </label>
        <input
          type="url"
          placeholder="https://example.com"
          class="input input-bordered w-full bg-base-200 text-base-content placeholder-gray-500 border-base-300 focus:border-primary"
          bind:value={newProjectUrl}
        />
        <label class="label">
          <span class="label-text-alt text-base-content opacity-70">This URL will be used as the default starting point for all crawls</span>
        </label>
      </div>

      <div class="modal-action">
        <button
          class="btn btn-ghost text-base-content hover:bg-base-200"
          on:click={() => {
            showCreateModal = false;
            error = null;
          }}
        >
          Cancel
        </button>
        <button
          class="btn btn-primary text-primary-content"
          on:click={handleCreateProject}
          disabled={creating || !newProjectName || !newProjectDomain || !newProjectUrl}
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


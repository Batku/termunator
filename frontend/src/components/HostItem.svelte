<script lang="ts">
  import { Loader2, AlertCircle, PenBox, Computer } from 'lucide-svelte';
  import { createEventDispatcher } from 'svelte';
  import type { Host } from '../types/api';

  export let host: Host;
  export let connecting: boolean = false;
  export let error: string | null = null;
  export let onSelect: (host: Host) => void;

  // For edit popup
  const dispatch = createEventDispatcher<{ editHost: Host }>();


  function editHost() {
    console.log('Sidebar: Edit host clicked, dispatching editHost event');
    dispatch('editHost', host);
  }

  function handleEdit(e: Event) {
    e.stopPropagation();
    editHost();
  }

</script>

<div class="relative group">
  <button 
    class="w-full p-3 text-left hover:bg-slate-700 border-b border-slate-700 transition-colors relative flex items-center gap-3 {connecting ? 'opacity-75' : ''}"
    on:click={() => onSelect(host)}
    disabled={connecting}
  >
    <!-- Host image/icon -->
    {#if host.image}
      <img src={host.image} alt="Host" class="w-8 h-8 rounded mr-2 object-cover bg-slate-700 border border-slate-600" />
    {:else}
      <span class="w-8 h-8 flex items-center justify-center rounded mr-2 bg-slate-700 border border-slate-600">
        <Computer size={22} class="text-slate-400" />
      </span>
    {/if}
    <div class="flex-1 min-w-0">
      <div class="font-medium text-white flex items-center min-w-0">
        <span class="truncate">{host.label}</span>
        {#if connecting}
          <Loader2 size={14} class="ml-2 animate-spin text-blue-400" />
        {:else if error}
          <AlertCircle size={14} class="ml-2 text-red-400" />
        {/if}
      </div>
      <div class="text-sm text-slate-400 truncate">{host.username}@{host.hostname}:{host.port}</div>
      {#if host.tags && host.tags.length > 0}
        <div class="flex flex-wrap gap-1 mt-1">
          {#each host.tags as tag}
            <span class="inline-block px-2 py-0.5 text-xs bg-slate-600 text-slate-300 rounded">
              {tag}
            </span>
          {/each}
        </div>
      {/if}
      {#if error}
        <div class="text-xs text-red-300 mt-1 bg-red-900/20 p-1 rounded">
          {error}
        </div>
      {/if}
    </div>
    <!-- Edit button (pencil) -->
    <button
      type="button"
      class="ml-2 p-1 rounded hover:bg-slate-600 text-slate-400 hover:text-white opacity-0 group-hover:opacity-100 transition-opacity"
      title="Edit Host"
      on:click|stopPropagation={handleEdit}
    >
      <PenBox size={16} />
    </button>
  </button>
</div>

<style>
  .group:hover .group-hover\:opacity-100 {
    opacity: 1 !important;
  }
</style>

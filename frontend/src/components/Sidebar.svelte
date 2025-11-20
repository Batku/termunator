<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Server, Plus, X } from 'lucide-svelte';
  import type { Host } from '../types/api';
  import HostItem from './HostItem.svelte';

  export let hosts: Host[] = [];
  export let collapsed: boolean = false;

  const dispatch = createEventDispatcher<{
    hostSelect: Host;
    addHost: void;
    toggleCollapse: boolean;
    editHost: Host;
  }>();

  // Loading and error states for individual hosts
  let connectingHosts: Set<string> = new Set();
  let hostErrors: Map<string, string> = new Map();

  function selectHost(host: Host) {
    // Clear any previous error for this host
    if (hostErrors.has(host.id)) {
      hostErrors.delete(host.id);
      hostErrors = new Map(hostErrors);
    }
    
    // Set loading state
    connectingHosts.add(host.id);
    connectingHosts = new Set(connectingHosts);
    
    dispatch('hostSelect', host);
  }

  // Function to be called by parent when connection succeeds
  export function onConnectionSuccess(hostId: string) {
    connectingHosts.delete(hostId);
    connectingHosts = new Set(connectingHosts);
    
    if (hostErrors.has(hostId)) {
      hostErrors.delete(hostId);
      hostErrors = new Map(hostErrors);
    }
  }

  // Function to be called by parent when connection fails
  export function onConnectionError(hostId: string, error: string) {
    connectingHosts.delete(hostId);
    connectingHosts = new Set(connectingHosts);
    
    hostErrors.set(hostId, error);
    hostErrors = new Map(hostErrors);
  }

  function addNewHost() {
    console.log('Sidebar: Add host clicked, dispatching addHost event');
    dispatch('addHost');
  }

  function handleAddHostClick(event: Event) {
    event.preventDefault();
    event.stopPropagation();
    console.log('Button click event triggered');
    addNewHost();
  }

  function toggleCollapse() {
    dispatch('toggleCollapse', !collapsed);
  }
</script>

{#if !collapsed}
  <!-- Full Sidebar -->
  <div class="w-80 bg-slate-800 border-r border-slate-700 flex flex-col">
    <!-- Header -->
    <div class="p-4 border-b border-slate-700 flex items-center justify-between">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Server size={20} />
        Hosts
      </h2>
      <div class="flex items-center gap-1">
        <button 
          on:click={handleAddHostClick}
          class="p-1 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
          title="Add Host"
          type="button"
        >
          <Plus size={16} />
        </button>
        <button 
          on:click={toggleCollapse}
          class="p-1 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
          title="Collapse Sidebar"
        >
          <X size={16} />
        </button>
      </div>
    </div>

    <!-- Host List -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
      {#each hosts as host (host.id)}
      <!-- HostItem component extraction -->
        <HostItem
          {host}
          connecting={connectingHosts.has(host.id)}
          error={hostErrors.get(host.id) || null}
          onSelect={selectHost}
          on:editHost={() => dispatch('editHost', host)}
        />
      {:else}
        <div class="p-4 text-center text-slate-400">
          <Server size={32} class="mx-auto mb-2 opacity-50" />
          <p class="text-sm">No hosts configured</p>
          <button 
            on:click={handleAddHostClick}
            class="mt-2 text-blue-400 hover:text-blue-300 text-sm underline"
            type="button"
          >
            Add your first host
          </button>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <!-- Collapsed Sidebar Toggle -->
  <div class="w-12 bg-slate-800 border-r border-slate-700 flex flex-col">
    <button 
      on:click={toggleCollapse}
      class="p-3 text-slate-400 hover:text-white hover:bg-slate-700 transition-colors"
      title="Expand Sidebar"
    >
      <Server size={20} />
    </button>
  </div>
{/if}

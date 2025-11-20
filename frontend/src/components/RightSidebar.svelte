<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Command, Plus, Play, X } from 'lucide-svelte';
  import type { Macro } from '../types/api';

  export let macros: Macro[] = [];
  export let collapsed: boolean = true;

  const dispatch = createEventDispatcher<{
    macroExecute: Macro;
    addMacro: void;
    toggleCollapse: boolean;
  }>();

  function executeMacro(macro: Macro) {
    dispatch('macroExecute', macro);
  }

  function addNewMacro() {
    dispatch('addMacro');
  }

  function toggleCollapse() {
    dispatch('toggleCollapse', !collapsed);
  }
</script>

{#if !collapsed}
  <!-- Full Right Sidebar -->
   <!-- this should also have history (move both to seperate components) -->
  <div class="w-80 bg-slate-800 border-l border-slate-700 flex flex-col">
    <!-- Header -->
    <div class="p-4 border-b border-slate-700 flex items-center justify-between">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Command size={20} />
        Macros
      </h2>
      <div class="flex items-center gap-1">
        <button 
          on:click={addNewMacro}
          class="p-1 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
          title="Add Macro"
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

    <!-- Macro List -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
      {#each macros as macro (macro.id)}
        <div class="p-3 border-b border-slate-700 hover:bg-slate-700 transition-colors">
          <div class="flex items-center justify-between mb-2">
            <h4 class="font-medium text-white truncate">{macro.label}</h4>
            <button 
              on:click={() => executeMacro(macro)}
              class="p-1 text-green-400 hover:text-green-300 hover:bg-slate-600 rounded transition-colors"
              title="Execute Macro"
            >
              <Play size={14} />
            </button>
          </div>
          <div class="text-xs text-slate-500 font-mono bg-slate-900 p-2 rounded">
            {macro.commands.slice(0, 2).join('; ')}
            {#if macro.commands.length > 2}
              <span class="text-slate-600">...</span>
            {/if}
          </div>
        </div>
      {:else}
        <div class="p-4 text-center text-slate-400">
          <Command size={32} class="mx-auto mb-2 opacity-50" />
          <p class="text-sm">No macros configured</p>
          <button 
            on:click={addNewMacro}
            class="mt-2 text-blue-400 hover:text-blue-300 text-sm underline"
          >
            Create your first macro
          </button>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <!-- Collapsed Right Sidebar Toggle -->
  <div class="w-12 bg-slate-800 border-l border-slate-700 flex flex-col">
    <button 
      on:click={toggleCollapse}
      class="p-3 text-slate-400 hover:text-white hover:bg-slate-700 transition-colors"
      title="Expand Macros"
    >
      <Command size={20} />
    </button>
  </div>
{/if}

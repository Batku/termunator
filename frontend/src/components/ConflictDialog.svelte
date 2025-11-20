<script lang="ts">
  import { createEventDispatcher } from "svelte";
  export let conflicts: Array<{ name: string; type: "file" | "folder" }>; // All conflicts
  export let currentIndex: number = 0; // Index of the current conflict
  export let show: boolean = false;
  const dispatch = createEventDispatcher();

  type ConflictAction = "replace" | "keep-both" | "cancel";

  let selectedAction: ConflictAction = "replace";
  let showList = false;

  function handleConfirm() {
    dispatch("resolve", {
      action: selectedAction,
      applyToAll: false,
      single: true,
    });
  }
  function handleApplyToAll() {
    dispatch("resolve", {
      action: selectedAction,
      applyToAll: true,
      single: false,
    });
  }
  function handleCancel() {
    dispatch("resolve", {
      action: "cancel",
      applyToAll: false,
      single: true,
    });
  }
</script>

{#if show}
  <div
    class="fixed inset-0 z-[9999] flex items-center justify-center bg-black bg-opacity-50"
  >
    <div
      class="bg-slate-800 rounded-lg shadow-lg p-6 w-full max-w-md border border-slate-600 z-[10000]"
    >
      <h2 class="text-lg font-bold mb-2 text-slate-100">
        File/Folder Conflict
      </h2>
      <div class="flex items-center justify-between mb-2">
        <span class="text-slate-400 text-sm"
          >Conflict {currentIndex + 1} of {conflicts.length}</span
        >
        <button
          class="text-xs text-blue-400 hover:underline"
          on:click={() => (showList = !showList)}
        >
          {showList ? "Hide" : "Show"} all conflicts
        </button>
      </div>
      {#if showList}
        <ul
          class="mb-2 max-h-24 overflow-y-auto text-slate-200 border border-slate-600 rounded p-2 bg-slate-700"
        >
          {#each conflicts as conflict, i}
            <li
              class="mb-1 flex items-center {i === currentIndex
                ? 'font-bold text-blue-300'
                : ''}"
            >
              <span class="mr-2"
                >{conflict.type === "folder" ? "📁" : "📄"}</span
              >
              <span>{conflict.name}</span>
              {#if i === currentIndex}
                <span class="ml-2 text-xs text-blue-400">(current)</span>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
      <div class="mb-4 flex items-center gap-2">
        <span class="mr-2"
          >{conflicts[currentIndex].type === "folder" ? "📁" : "📄"}</span
        >
        <span class="font-semibold text-slate-200"
          >{conflicts[currentIndex].name}</span
        >
      </div>
      <div class="mb-6 flex flex-col gap-2">
        <div class="flex gap-3 justify-center">
          <button
            type="button"
            class="flex flex-col items-center px-4 py-2 rounded border border-slate-600 bg-slate-700 text-slate-100 hover:bg-blue-700 focus:ring-2 focus:ring-blue-400 transition-all {selectedAction ===
            'replace'
              ? 'ring-2 ring-blue-400 border-blue-500'
              : ''}"
            on:click={() => (selectedAction = "replace")}
            aria-pressed={selectedAction === "replace"}
          >
            <span class="font-semibold">Replace</span>
            <span class="text-xs text-slate-400">Overwrite existing</span>
          </button>
          <button
            type="button"
            class="flex flex-col items-center px-4 py-2 rounded border border-slate-600 bg-slate-700 text-slate-100 hover:bg-blue-700 focus:ring-2 focus:ring-blue-400 transition-all {selectedAction ===
            'keep-both'
              ? 'ring-2 ring-blue-400 border-blue-500'
              : ''}"
            on:click={() => (selectedAction = "keep-both")}
            aria-pressed={selectedAction === "keep-both"}
          >
            <span class="font-semibold">Keep Both</span>
            <span class="text-xs text-slate-400">Auto-rename new</span>
          </button>
          <button
            type="button"
            class="flex flex-col items-center px-4 py-2 rounded border border-slate-600 bg-slate-700 text-slate-100 hover:bg-blue-700 focus:ring-2 focus:ring-blue-400 transition-all {selectedAction ===
            'cancel'
              ? 'ring-2 ring-blue-400 border-blue-500'
              : ''}"
            on:click={() => (selectedAction = "cancel")}
            aria-pressed={selectedAction === "cancel"}
          >
            <span class="font-semibold">Skip</span>
            <span class="text-xs text-slate-400">Skip this file</span>
          </button>
        </div>
      </div>
      <div class="flex justify-end space-x-2 mt-4">
        <button
          class="px-4 py-2 rounded bg-slate-600 text-white hover:bg-slate-500"
          on:click={handleCancel}>Cancel</button
        >
        {#if conflicts.length > 1}
          <button
            type="button"
            class="px-4 py-2 rounded border border-blue-500 bg-blue-600 text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-400 transition-all"
            on:click={handleApplyToAll}
          >
            Apply to all
          </button>
        {/if}
        <button
          class="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-500"
          on:click={handleConfirm}>Confirm</button
        >
      </div>
    </div>
  </div>
{/if}

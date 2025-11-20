<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { TerminalIcon, X, AlertCircle } from 'lucide-svelte';
  import type { Session } from '../types/api';

  export let sessions: Session[] = [];
  export let activeSessionId: string | null = null;
  export let sessionErrors: Map<string, string> = new Map();

  const dispatch = createEventDispatcher<{
    sessionClosed: string;
    sessionActivated: string;
  }>();

  function selectSession(sessionId: string) {
    dispatch('sessionActivated', sessionId);
  }

  function closeSession(sessionId: string) {
    dispatch('sessionClosed', sessionId);
  }
</script>


<div class="flex items-center bg-slate-800 border-b border-slate-700 min-h-[48px] w-full min-w-0 custom-scrollbar">
  <div class="flex-1 flex items-center overflow-x-auto whitespace-nowrap scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-slate-800 w-full min-w-0">
    {#each sessions as session (session.id)}
      <div class="flex items-center">
        <button
          class="px-4 py-2 bg-slate-800 border-t border-l border-r border-slate-700 rounded-t-lg text-sm font-medium transition-colors flex-shrink {activeSessionId === session.id ? 'bg-slate-900 border-slate-600 text-white' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700'}"
          on:click={() => selectSession(session.id)}
        >
          <div class="flex items-center min-w-0">
            <TerminalIcon size={14} class="mr-2" />
            <span
              class="truncate min-w-[4ch] max-w-[12ch]"
              style="line-height:1;"
              title={session.displayName || session.host?.label || 'Unknown Host'}
            >
              {session.displayName || session.host?.label || 'Unknown Host'}
            </span>
            {#if sessionErrors.has(session.id)}
              <AlertCircle size={12} class="ml-1 text-red-400" />
            {/if}
            <button
              class="ml-2 hover:bg-slate-600 rounded p-0.5"
              on:click|stopPropagation={() => closeSession(session.id)}
            >
              <X size={12} />
            </button>
          </div>
        </button>
      </div>
    {:else}
      <div class="px-4 py-2 text-sm text-slate-400">No active sessions</div>
    {/each}
  </div>
</div>

<style>
  /* Custom scrollbar for horizontal tab bar */
  .scrollbar-thin {
    scrollbar-width: thin;
  }
  .scrollbar-thumb-slate-700::-webkit-scrollbar-thumb {
    background-color: #334155;
  }
  .scrollbar-track-slate-800::-webkit-scrollbar-track {
    background-color: #1e293b;
  }
  .scrollbar-thin::-webkit-scrollbar {
    height: 6px;
  }
  .truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { X, Check, AlertTriangle, Loader2, Shield, Key, Wifi } from 'lucide-svelte';
  import type { Host } from '../types/api';

  export let host: Host;
  export let onAccept: () => void;
  export let onReject: () => void;

  const dispatch = createEventDispatcher<{
    close: void;
    accept: void;
    reject: void;
  }>();

  interface ConnectionStep {
    id: string;
    title: string;
    status: 'pending' | 'in-progress' | 'completed' | 'error' | 'warning';
    message?: string;
    icon?: any;
  }

  export let steps: ConnectionStep[] = [
    { id: 'resolving', title: 'Resolving hostname', status: 'pending', icon: Wifi },
    { id: 'connecting', title: 'Establishing connection', status: 'pending', icon: Wifi },
    { id: 'handshake', title: 'SSH handshake', status: 'pending', icon: Key },
    { id: 'hostkey', title: 'Verifying host key', status: 'pending', icon: Shield },
    { id: 'auth', title: 'Authenticating', status: 'pending', icon: Key },
    { id: 'session', title: 'Starting session', status: 'pending', icon: Check }
  ];

  export let hostKeyInfo: {
    fingerprint: string;
    algorithm: string;
    isNewHost: boolean;
  } | null = null;

  export let showHostKeyDialog = false;

  function handleAccept() {
    dispatch('accept');
    onAccept();
  }

  function handleReject() {
    dispatch('reject');
    onReject();
  }

  function handleClose() {
    dispatch('close');
    onReject();
  }
</script>

<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
  <div class="bg-slate-800 rounded-lg border border-slate-700 w-full max-w-lg">
    <!-- Header -->
    <div class="flex items-center justify-between p-6 border-b border-slate-700">
      <div class="flex items-center gap-3">
        <Wifi size={20} class="text-blue-400" />
        <div>
          <h2 class="text-lg font-semibold text-white">Connecting to {host.label}</h2>
          <p class="text-sm text-slate-400">{host.username}@{host.hostname}:{host.port}</p>
        </div>
      </div>
      <button 
        on:click={handleClose}
        class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
      >
        <X size={20} />
      </button>
    </div>

    <!-- Connection Steps -->
    <div class="p-6 space-y-4">
      {#each steps as step}
        <div class="flex items-center gap-3">
          <div class="flex-shrink-0">
            {#if step.status === 'completed'}
              <div class="w-6 h-6 bg-green-500 rounded-full flex items-center justify-center">
                <Check size={14} class="text-white" />
              </div>
            {:else if step.status === 'error'}
              <div class="w-6 h-6 bg-red-500 rounded-full flex items-center justify-center">
                <X size={14} class="text-white" />
              </div>
            {:else if step.status === 'warning'}
              <div class="w-6 h-6 bg-yellow-500 rounded-full flex items-center justify-center">
                <AlertTriangle size={14} class="text-white" />
              </div>
            {:else if step.status === 'in-progress'}
              <div class="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center">
                <Loader2 size={14} class="text-white animate-spin" />
              </div>
            {:else}
              <div class="w-6 h-6 bg-slate-600 rounded-full flex items-center justify-center">
                <svelte:component this={step.icon} size={14} class="text-slate-400" />
              </div>
            {/if}
          </div>
          <div class="flex-1">
            <div class="text-sm font-medium text-white">{step.title}</div>
            {#if step.message}
              <div class="text-xs text-slate-400">{step.message}</div>
            {/if}
          </div>
        </div>
      {/each}
    </div>

    <!-- Host Key Verification Dialog -->
    {#if showHostKeyDialog && hostKeyInfo}
      <div class="border-t border-slate-700 p-6 bg-slate-750">
        <div class="flex items-center gap-2 mb-4">
          <Shield size={20} class="text-yellow-400" />
          <h3 class="text-lg font-semibold text-white">Host Key Verification</h3>
        </div>
        
        <div class="space-y-3 mb-6">
          {#if hostKeyInfo.isNewHost}
            <p class="text-sm text-slate-300">
              The authenticity of host '<span class="font-mono text-blue-400">{host.hostname}</span>' can't be established.
            </p>
          {:else}
            <p class="text-sm text-slate-300">
              The host key for '<span class="font-mono text-blue-400">{host.hostname}</span>' has changed.
            </p>
          {/if}
          
          <div class="bg-slate-900 p-3 rounded border border-slate-600">
            <div class="text-xs text-slate-400 mb-1">Host Key Fingerprint:</div>
            <div class="font-mono text-sm text-white break-all">
              {hostKeyInfo.algorithm} {hostKeyInfo.fingerprint}
            </div>
          </div>
          
          <p class="text-sm text-slate-300">
            Are you sure you want to continue connecting?
          </p>
        </div>

        <div class="flex justify-end gap-3">
          <button 
            on:click={handleReject}
            class="px-4 py-2 text-slate-300 border border-slate-600 rounded-md hover:bg-slate-700 transition-colors"
          >
            Cancel
          </button>
          <button 
            on:click={handleAccept}
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            Accept & Continue
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>
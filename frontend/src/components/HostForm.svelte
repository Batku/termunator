<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { X } from 'lucide-svelte';
  import type { HostCreateRequest, Host } from '../types/api';
  import { HostAPI } from '../lib/api';
  import { addNotification, hosts } from '../types/stores';
  import PrivateKeySelector from './PrivateKeySelector.svelte';
  export let host: Host | null = null;
  const dispatch = createEventDispatcher<{
    close: void;
    saved: void;
  }>();

  let isEditing = host !== null;
  let hostForm: HostCreateRequest = {
    label: '',
    hostname: '',
    port: 22,
    username: '',
    auth_method: 'password',
    password: '',
    private_key: '',
    tags: []
  };
  
  let newTag = '';
  let privateKeyFileName = '';
  // Initialize form with host data if editing
  if (host) {
    console.log('Editing host:', host);
    hostForm = {
      label: host.label,
      hostname: host.hostname,
      port: host.port,
      username: host.username,
      auth_method: host.auth_method,
      password: host.password || '',
      private_key: host.private_key || '',
      tags: host.tags || []
    };
    // Set private key filename if editing and has a key
    if (host.private_key) {
      privateKeyFileName = 'Previously uploaded key';
    }
  }


  function addTag() {
    if (newTag.trim() && !hostForm.tags.includes(newTag.trim())) {
      hostForm.tags = [...hostForm.tags, newTag.trim()];
      newTag = '';
    }
  }

  function removeTag(tagToRemove: string) {
    hostForm.tags = hostForm.tags.filter(tag => tag !== tagToRemove);
  }

  function handleTagKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    }
  }

  function validateForm(): string | null {
    if (!hostForm.label.trim()) return 'Display name is required';
    if (!hostForm.hostname.trim()) return 'Hostname is required';
    if (!hostForm.username.trim()) return 'Username is required';
    if (hostForm.port < 1 || hostForm.port > 65535) return 'Port must be between 1 and 65535';
    
    if (hostForm.auth_method === 'password' && (!hostForm.password || !hostForm.password.trim())) {
      return 'Password is required for password authentication';
    }
    
    if (hostForm.auth_method === 'private_key' && (!hostForm.private_key || !hostForm.private_key.trim())) {
      return 'Private key file is required for private key authentication';
    }
    
    // Basic private key format validation
    if (hostForm.auth_method === 'private_key' && hostForm.private_key && hostForm.private_key.trim()) {
      const key = hostForm.private_key.trim();
      if (!key.includes('BEGIN') || !key.includes('PRIVATE KEY')) {
        return 'Invalid private key format. Please ensure it\'s a valid private key file.';
      }
    }
    
    return null;
  }

  async function saveHost() {
    const validationError = validateForm();
    if (validationError) {
      addNotification({
        type: 'error',
        title: validationError
      });
      return;
    }

    try {
      if (isEditing && host) {
        console.log('Updating host with data:', hostForm);
        const updatedHost = await HostAPI.update(host.id, hostForm);
        console.log('Host updated successfully:', updatedHost);
        
        // Update the hosts store
        hosts.update(currentHosts => 
          currentHosts.map(h => h.id === host.id ? updatedHost : h)
        );
        
        addNotification({
          type: 'success',
          title: 'Host updated successfully'
        });
      } else {
        console.log('Creating host with data:', hostForm);
        const newHost = await HostAPI.create(hostForm);
        console.log('Host created successfully:', newHost);
        hosts.update(currentHosts => [...currentHosts, newHost]);
        addNotification({
          type: 'success',
          title: 'Host created successfully'
        });
      }
      dispatch('saved');
      dispatch('close');
    } catch (error) {
      console.error('Failed to save host:', error);
      addNotification({
        type: 'error',
        title: isEditing ? 'Failed to update host' : 'Failed to create host'
      });
    }
  }

  // Delete host logic
  let showDeleteConfirm = false;
  async function deleteHost() {
    if (!host) return;
    try {
      await HostAPI.delete(host.id);
      hosts.update(currentHosts => currentHosts.filter(h => h.id !== host.id));
      addNotification({
        type: 'success',
        title: 'Host deleted successfully'
      });
      dispatch('saved');
      dispatch('close');
    } catch (error) {
      console.error('Failed to delete host:', error);
      addNotification({
        type: 'error',
        title: 'Failed to delete host'
      });
    }
  }

  function cancelForm() {
    dispatch('close');
  }
</script>

<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
  <div class="bg-slate-800 rounded-lg border border-slate-700 w-full max-w-2xl max-h-[90vh] flex flex-col">
    <div class="flex items-center justify-between p-6 border-b border-slate-700">
      <h2 class="text-xl font-semibold text-white">
        {isEditing ? 'Edit Host' : 'Add New Host'}
      </h2>
      <button 
        on:click={cancelForm}
        class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
      >
        <X size={20} />
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-6 custom-scrollbar">
      <form on:submit|preventDefault={saveHost} class="space-y-6">
          <!-- Basic Information -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label for="host-label" class="block text-sm font-medium text-slate-300 mb-2">
                Display Name *
              </label>
              <input 
                id="host-label"
                bind:value={hostForm.label}
                type="text" 
                required
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
                placeholder="My Server"
              />
            </div>
            <div>
              <label for="host-hostname" class="block text-sm font-medium text-slate-300 mb-2">
                Hostname/IP *
              </label>
              <input 
                id="host-hostname"
                bind:value={hostForm.hostname}
                type="text" 
                required
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
                placeholder="192.168.1.100"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label for="host-port" class="block text-sm font-medium text-slate-300 mb-2">
                Port
              </label>
              <input 
                id="host-port"
                bind:value={hostForm.port}
                type="number" 
                min="1" 
                max="65535"
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
              />
            </div>
            <div>
              <label for="host-username" class="block text-sm font-medium text-slate-300 mb-2">
                Username *
              </label>
              <input 
                id="host-username"
                bind:value={hostForm.username}
                type="text" 
                required
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
                placeholder="username"
              />
            </div>
          </div>

          <!-- Authentication Method -->
          <fieldset>
            <legend class="block text-sm font-medium text-slate-300 mb-3">
              Authentication Method *
            </legend>
            <div class="space-y-2">
              <label class="flex items-center">
                <input 
                  type="radio" 
                  bind:group={hostForm.auth_method} 
                  value="password"
                  class="text-blue-500 bg-slate-700 border-slate-600 focus:ring-blue-500 focus:ring-offset-slate-800"
                />
                <span class="ml-2 text-slate-300">Password</span>
              </label>
              <label class="flex items-center">
                <input 
                  type="radio" 
                  bind:group={hostForm.auth_method} 
                  value="private_key"
                  class="text-blue-500 bg-slate-700 border-slate-600 focus:ring-blue-500 focus:ring-offset-slate-800"
                />
                <span class="ml-2 text-slate-300">Private Key</span>
              </label>
              <label class="flex items-center">
                <input 
                  type="radio" 
                  bind:group={hostForm.auth_method} 
                  value="ssh_agent"
                  class="text-blue-500 bg-slate-700 border-slate-600 focus:ring-blue-500 focus:ring-offset-slate-800"
                />
                <span class="ml-2 text-slate-300">SSH Agent</span>
              </label>
            </div>
          </fieldset>

          <!-- Conditional Authentication Fields -->
          {#if hostForm.auth_method === 'password'}
            <div>
              <label for="host-password" class="block text-sm font-medium text-slate-300 mb-2">
                Password *
              </label>
              <input 
                id="host-password"
                bind:value={hostForm.password}
                type="password" 
                required
                class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
                placeholder="Enter password"
              />
            </div>
          {:else if hostForm.auth_method === 'private_key'}
            <div>
              <div class="block text-sm font-medium text-slate-300 mb-2">
                Private Key *
              </div>
              
              <PrivateKeySelector
                bind:selectedKeyData={hostForm.private_key}
                bind:selectedKeyName={privateKeyFileName}
                on:keySelected={(event) => {
                  hostForm.private_key = event.detail.data;
                  privateKeyFileName = event.detail.name;
                }}
              />
              
              
            </div>
          {/if}

          <!-- Tags -->
          <div>
            <span class="block text-sm font-medium text-slate-300 mb-2">
              Tags
            </span>
            <div class="flex flex-wrap gap-2 mb-2">
              {#each hostForm.tags as tag}
                <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {tag}
                  <button 
                    type="button"
                    on:click={() => removeTag(tag)}
                    class="ml-1 hover:text-blue-600"
                  >
                    <X size={12} />
                  </button>
                </span>
              {/each}
            </div>
            <div class="flex gap-2">
              <input 
                bind:value={newTag}
                on:keydown={handleTagKeydown}
                type="text" 
                class="flex-1 px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" 
                placeholder="Add a tag..."
                aria-label="New tag"
              />
              <button 
                type="button"
                on:click={addTag}
                class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                Add
              </button>
            </div>
          </div>

          <!-- Form Actions -->
          <div class="flex justify-between gap-3 pt-6 border-t border-slate-700">
            {#if isEditing}
              <div class="flex items-center">
                <button
                  type="button"
                  class="px-4 py-2 text-red-400 border border-red-600 rounded-md hover:bg-red-700 hover:text-white transition-colors"
                  on:click={() => showDeleteConfirm = true}
                >
                  Delete Host
                </button>
              </div>
            {/if}
            <div class="flex gap-3 justify-end flex-1">
              <button 
                type="button"
                on:click={cancelForm}
                class="px-4 py-2 text-slate-300 border border-slate-600 rounded-md hover:bg-slate-700 transition-colors"
              >
                Cancel
              </button>
              <button 
                type="submit"
                class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                {isEditing ? 'Update Host' : 'Create Host'}
              </button>
            </div>
          </div>

          {#if showDeleteConfirm}
            <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-60">
              <div class="bg-slate-800 border border-slate-700 rounded-lg p-8 max-w-sm w-full text-center">
                <h3 class="text-lg font-semibold text-red-400 mb-4">Delete Host?</h3>
                <p class="text-slate-300 mb-4">Are you sure you want to delete this host? <br/>This action cannot be undone.</p>
                <div class="flex justify-center gap-4">
                  <button
                    class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
                    on:click={deleteHost}
                  >
                    Yes, Delete
                  </button>
                  <button
                    class="px-4 py-2 border border-slate-600 text-slate-300 rounded-md hover:bg-slate-700 transition-colors"
                    on:click={() => showDeleteConfirm = false}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          {/if}
        </form>
    </div>
  </div>
</div>

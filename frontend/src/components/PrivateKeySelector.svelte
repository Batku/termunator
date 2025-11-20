<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { X, Upload, Key, Trash2 } from "lucide-svelte";
  import { addNotification } from "../types/stores";
  import * as App from "../../wailsjs/go/main/App";
  import { models } from "../../wailsjs/go/models";

  export let selectedKeyId: string | null = null;
  export let selectedKeyData: string = "";
  export let selectedKeyName: string = "";

  const dispatch = createEventDispatcher<{
    keySelected: { id: string | null; data: string; name: string };
  }>();

  let savedKeys: models.PrivateKeyInfo[] = [];
  let showUploadDialog = false;
  let uploadKeyName = "";
  let uploadKeyData = "";
  let uploadMethod: "file" | "paste" = "file";
  let dragActive = false;

  onMount(async () => {
    await loadSavedKeys();
  });

  async function loadSavedKeys() {
    try {
      savedKeys = await App.GetPrivateKeys();
    } catch (error) {
      console.error("Failed to load saved keys:", error);
      addNotification({
        type: "error",
        title: "Failed to load saved keys",
      });
    }
  }

  async function selectSavedKey(key: models.PrivateKeyInfo) {
    try {
      const fullKey = await App.GetPrivateKey(key.id);
      if (fullKey && fullKey.key_data) {
        selectedKeyId = key.id;
        selectedKeyData = fullKey.key_data;
        selectedKeyName = key.name;
        dispatch("keySelected", {
          id: selectedKeyId,
          data: selectedKeyData,
          name: selectedKeyName,
        });
      }
    } catch (error) {
      console.error("Failed to load private key:", error);
      addNotification({
        type: "error",
        title: "Failed to load private key",
      });
    }
  }


  async function saveNewKey() {
    if (!uploadKeyName.trim() || !uploadKeyData.trim()) {
      addNotification({
        type: "error",
        title: "Please provide both a name and key data",
      });
      return;
    }

    try {
      const newKey = await App.CreatePrivateKey({
        name: uploadKeyName,
        key_data: uploadKeyData,
      });

      addNotification({
        type: "success",
        title: "Private key saved successfully",
      });

      // Refresh the list and select the new key
      await loadSavedKeys();
      const keyInfo = models.PrivateKeyInfo.createFrom({
        id: newKey.id,
        name: newKey.name,
        fingerprint: newKey.fingerprint,
        key_type: newKey.key_type,
        created_at: newKey.created_at,
      });
      await selectSavedKey(keyInfo);

      // Reset upload form
      uploadKeyName = "";
      uploadKeyData = "";
      uploadMethod = "file";
      showUploadDialog = false;
    } catch (error) {
      console.error("Failed to save private key:", error);
      addNotification({
        type: "error",
        title: "Failed to save private key",
      });
    }
  }

  async function deleteKey(key: models.PrivateKeyInfo) {
    if (
      !confirm(`Are you sure you want to delete the private key "${key.name}"?`)
    ) {
      return;
    }

    try {
      await App.DeletePrivateKey(key.id);
      addNotification({
        type: "success",
        title: "Private key deleted",
      });

      if (selectedKeyId === key.id) {
        selectedKeyId = null;
        selectedKeyData = "";
        selectedKeyName = "";
        dispatch("keySelected", {
          id: null,
          data: "",
          name: "",
        });
      }

      await loadSavedKeys();
    } catch (error) {
      console.error("Failed to delete private key:", error);
      addNotification({
        type: "error",
        title: "Failed to delete private key",
      });
    }
  }

  function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        uploadKeyData = e.target?.result as string;
        if (!uploadKeyName) {
          uploadKeyName = file.name.replace(/\.[^/.]+$/, ""); // Remove extension
        }
      };
      reader.readAsText(file);
    }
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    dragActive = true;
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault();
    dragActive = false;
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    dragActive = false;

    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      const file = files[0];
      const reader = new FileReader();
      reader.onload = (e) => {
        uploadKeyData = e.target?.result as string;
        if (!uploadKeyName) {
          uploadKeyName = file.name.replace(/\.[^/.]+$/, ""); // Remove extension
        }
      };
      reader.readAsText(file);
    }
  }

  function resetUploadData() {
    uploadKeyData = "";
    uploadKeyName = "";
  }

  function switchMethod(method: "file" | "paste") {
    uploadMethod = method;
    resetUploadData();
  }
</script>

<div class="space-y-4">
  <!-- Saved Keys Section -->
  <div>
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-medium text-slate-300">Saved Private Keys</h3>
      <button
        type="button"
        on:click={() => (showUploadDialog = true)}
        class="text-xs px-2 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
      >
        Add New Key
      </button>
    </div>

    {#if !savedKeys || savedKeys.length === 0}
      <div
        class="text-sm text-slate-400 py-4 text-center border border-slate-600 rounded-md"
      >
        No saved private keys found
      </div>
    {:else}
      <div class="space-y-2 max-h-32 overflow-y-auto">
        {#each savedKeys as key}
          <div
            class="flex items-center justify-between p-3 border rounded-md transition-colors cursor-pointer
                   {selectedKeyId === key.id
              ? 'border-blue-500 bg-blue-500/10'
              : 'border-slate-600 hover:border-slate-500'}"
            on:click={() => selectSavedKey(key)}
          >
            <div class="flex items-center gap-3">
              <Key size={16} class="text-slate-400" />
              <div>
                <div class="text-sm font-medium text-white">{key.name}</div>
                <div class="text-xs text-slate-400">
                  {key.key_type} • {key.fingerprint.substring(0, 20)}...
                </div>
              </div>
            </div>
            <button
              type="button"
              on:click|stopPropagation={() => deleteKey(key)}
              class="p-1 text-slate-400 hover:text-red-400 transition-colors"
            >
              <Trash2 size={14} />
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Upload Dialog -->
{#if showUploadDialog}
  <div
    class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
  >
    <div
      class="bg-slate-800 rounded-lg border border-slate-700 w-full max-w-md p-6"
    >
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-white">Add New Private Key</h3>
        <button
          on:click={() => (showUploadDialog = false)}
          class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
        >
          <X size={16} />
        </button>
      </div>
      <div class="mb-2 text-xs text-yellow-400 font-bold">
        [DEBUG: Upload Dialog Open]
      </div>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">
            Key Name *
          </label>
          <input
            bind:value={uploadKeyName}
            type="text"
            required
            class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
            placeholder="My Server Key"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">
            Input Method *
          </label>
          <div class="flex gap-2 mb-3">
            <button
              type="button"
              class="flex-1 px-3 py-2 text-sm rounded-md border transition-colors {uploadMethod ===
              'file'
                ? 'bg-blue-600 border-blue-600 text-white'
                : 'bg-slate-700 border-slate-600 text-slate-300 hover:bg-slate-600'}"
              on:click={() => switchMethod("file")}
            >
              Upload File
            </button>
            <button
              type="button"
              class="flex-1 px-3 py-2 text-sm rounded-md border transition-colors {uploadMethod ===
              'paste'
                ? 'bg-blue-600 border-blue-600 text-white'
                : 'bg-slate-700 border-slate-600 text-slate-300 hover:bg-slate-600'}"
              on:click={() => switchMethod("paste")}
            >
              Paste Key
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">
            Private Key *
          </label>

          {#if uploadMethod === "file"}
            <!-- File Upload Area -->
            <div
              class="border-2 border-dashed rounded-lg p-6 text-center transition-colors cursor-pointer
                     {dragActive
                ? 'border-blue-400 bg-blue-500/10'
                : 'border-slate-600 hover:border-slate-500'}"
              on:dragover={handleDragOver}
              on:dragleave={handleDragLeave}
              on:drop={handleDrop}
              on:click={() =>
                document.getElementById("upload-key-file")?.click()}
            >
              {#if uploadKeyData}
                <div class="text-green-400 mb-2">
                  ✓ Key loaded ({uploadKeyData.length} characters)
                </div>
                <button
                  type="button"
                  on:click|stopPropagation={() => {
                    uploadKeyData = "";
                    uploadKeyName = "";
                  }}
                  class="text-red-400 hover:text-red-300 text-sm"
                >
                  Remove key
                </button>
              {:else}
                <Upload size={32} class="mx-auto mb-3 text-slate-400" />
                <div class="text-slate-400 mb-2">
                  Drag and drop your private key file here
                </div>
                <div class="text-sm text-slate-500 mb-3">or</div>
                <div
                  class="inline-block px-4 py-2 text-sm border border-slate-600 rounded text-slate-300 bg-slate-700 hover:bg-slate-600 transition-colors"
                >
                  Browse Files
                </div>
              {/if}

              <input
                type="file"
                accept=".pem,.key,.ppk,.openssh,*"
                on:change={handleFileSelect}
                class="hidden"
                id="upload-key-file"
              />
            </div>

            <div class="mt-2 text-xs text-slate-400">
              Supported formats: .pem, .key, .ppk, .openssh or any text file
            </div>
          {:else}
            <!-- Manual Paste Area -->
            <textarea
              bind:value={uploadKeyData}
              class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white placeholder-slate-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              rows="8"
              placeholder="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----"
            ></textarea>

            <div class="mt-2 text-xs text-slate-400">
              Paste your private key content here (including BEGIN/END lines)
            </div>
          {/if}
        </div>

        <div class="flex justify-end gap-3">
          <button
            type="button"
            on:click={() => (showUploadDialog = false)}
            class="px-4 py-2 text-slate-300 border border-slate-600 rounded-md hover:bg-slate-700 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            on:click={saveNewKey}
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            Save Key
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

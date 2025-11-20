<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { X, Palette, Shield, User, Terminal, Save, Trash2, RotateCcw } from 'lucide-svelte';
  import { terminalTheme, type TerminalTheme } from '../types/stores';
  import * as App from '../../wailsjs/go/main/App';
  
  const dispatch = createEventDispatcher<{
    close: void;
    saveSettings: any;
  }>();

  export let show = false;

  // Settings state
  let activeTab: 'terminal' | 'known-hosts' | 'account' = 'terminal';
  
  // Terminal theme presets
  const themePresets: TerminalTheme[] = [
    {
      name: 'Default Dark',
      background: '#1e293b',
      foreground: '#f1f5f9',
      cursor: '#06b6d4',
      cursorAccent: '#1e293b',
      black: '#1e293b',
      red: '#ef4444',
      green: '#22c55e',
      yellow: '#eab308',
      blue: '#3b82f6',
      magenta: '#a855f7',
      cyan: '#06b6d4',
      white: '#f1f5f9',
      brightBlack: '#475569',
      brightRed: '#f87171',
      brightGreen: '#4ade80',
      brightYellow: '#facc15',
      brightBlue: '#60a5fa',
      brightMagenta: '#c084fc',
      brightCyan: '#22d3ee',
      brightWhite: '#ffffff'
    },
    {
      name: 'Monokai',
      background: '#272822',
      foreground: '#f8f8f2',
      cursor: '#f8f8f0',
      cursorAccent: '#272822',
      black: '#272822',
      red: '#f92672',
      green: '#a6e22e',
      yellow: '#f4bf75',
      blue: '#66d9ef',
      magenta: '#ae81ff',
      cyan: '#a1efe4',
      white: '#f8f8f2',
      brightBlack: '#75715e',
      brightRed: '#f92672',
      brightGreen: '#a6e22e',
      brightYellow: '#f4bf75',
      brightBlue: '#66d9ef',
      brightMagenta: '#ae81ff',
      brightCyan: '#a1efe4',
      brightWhite: '#f9f8f5'
    },
    {
      name: 'Dracula',
      background: '#282a36',
      foreground: '#f8f8f2',
      cursor: '#f8f8f0',
      cursorAccent: '#282a36',
      black: '#21222c',
      red: '#ff5555',
      green: '#50fa7b',
      yellow: '#f1fa8c',
      blue: '#bd93f9',
      magenta: '#ff79c6',
      cyan: '#8be9fd',
      white: '#f8f8f2',
      brightBlack: '#6272a4',
      brightRed: '#ff6e6e',
      brightGreen: '#69ff94',
      brightYellow: '#ffffa5',
      brightBlue: '#d6acff',
      brightMagenta: '#ff92df',
      brightCyan: '#a4ffff',
      brightWhite: '#ffffff'
    },
    {
      name: 'One Dark',
      background: '#282c34',
      foreground: '#abb2bf',
      cursor: '#528bff',
      cursorAccent: '#282c34',
      black: '#282c34',
      red: '#e06c75',
      green: '#98c379',
      yellow: '#e5c07b',
      blue: '#61afef',
      magenta: '#c678dd',
      cyan: '#56b6c2',
      white: '#abb2bf',
      brightBlack: '#5c6370',
      brightRed: '#e06c75',
      brightGreen: '#98c379',
      brightYellow: '#e5c07b',
      brightBlue: '#61afef',
      brightMagenta: '#c678dd',
      brightCyan: '#56b6c2',
      brightWhite: '#ffffff'
    },
    {
      name: 'Solarized Dark',
      background: '#002b36',
      foreground: '#839496',
      cursor: '#93a1a1',
      cursorAccent: '#002b36',
      black: '#073642',
      red: '#dc322f',
      green: '#859900',
      yellow: '#b58900',
      blue: '#268bd2',
      magenta: '#d33682',
      cyan: '#2aa198',
      white: '#eee8d5',
      brightBlack: '#002b36',
      brightRed: '#cb4b16',
      brightGreen: '#586e75',
      brightYellow: '#657b83',
      brightBlue: '#839496',
      brightMagenta: '#6c71c4',
      brightCyan: '#93a1a1',
      brightWhite: '#fdf6e3'
    },
    {
    name: "Catppuccin Mocha",
      background: "#1e1e2e",
      foreground: "#cdd6f4",
      cursor: "#f5e0dc",
      cursorAccent: "#1e1e2e",
      black: "#45475a",
      red: "#f38ba8",
      green: "#a6e3a1",
      yellow: "#f9e2af",
      blue: "#89b4fa",
      magenta: "#f5c2e7",
      cyan: "#94e2d5",
      white: "#bac2de",
      brightBlack: "#585b70",
      brightRed: "#f38ba8",
      brightGreen: "#a6e3a1",
      brightYellow: "#f9e2af",
      brightBlue: "#89b4fa",
      brightMagenta: "#f5c2e7",
      brightCyan: "#94e2d5",
      brightWhite: "#a6adc8"
  },
  {
    name: "Catppuccin Latte",
    background: "#eff1f5",
    foreground: "#4c4f69",
    cursor: "#dc8a78",
    cursorAccent: "#eff1f5",
    black: "#9ca0b0",
    red: "#d20f39",
    green: "#40a02b",
    yellow: "#df8e1d",
    blue: "#1e66f5",
    magenta: "#ea76cb",
    cyan: "#179299",
    white: "#acb0be",
    brightBlack: "#bcc0cc",
    brightRed: "#d20f39",
    brightGreen: "#40a02b",
    brightYellow: "#df8e1d",
    brightBlue: "#1e66f5",
    brightMagenta: "#ea76cb",
    brightCyan: "#179299",
    brightWhite: "#acb0be"
  },
  {
    name: "Sakura",
    background: "#feedf3",
    foreground: "#564448",
    cursor: "#ff5f8a",
    cursorAccent: "#feedf3",
    black: "#564448",
    red: "#ff5f8a",
    green: "#7bb972",
    yellow: "#e6c446",
    blue: "#6c99bb",
    magenta: "#e5a2cd",
    cyan: "#7bb972",
    white: "#ffffff",
    brightBlack: "#877379",
    brightRed: "#ff5f8a",
    brightGreen: "#7bb972",
    brightYellow: "#e6c446",
    brightBlue: "#6c99bb",
    brightMagenta: "#e5a2cd",
    brightCyan: "#7bb972",
    brightWhite: "#ffffff"
  },
  {
    name: "Nord",
    background: "#2e3440",
    foreground: "#d8dee9",
    cursor: "#d8dee9",
    cursorAccent: "#2e3440",
    black: "#3b4252",
    red: "#bf616a",
    green: "#a3be8c",
    yellow: "#ebcb8b",
    blue: "#81a1c1",
    magenta: "#b48ead",
    cyan: "#88c0d0",
    white: "#e5e9f0",
    brightBlack: "#4c566a",
    brightRed: "#bf616a",
    brightGreen: "#a3be8c",
    brightYellow: "#ebcb8b",
    brightBlue: "#81a1c1",
    brightMagenta: "#b48ead",
    brightCyan: "#8fbcbb",
    brightWhite: "#eceff4"
  },
  {
    name: "Gruvbox Dark",
    background: "#282828",
    foreground: "#ebdbb2",
    cursor: "#ebdbb2",
    cursorAccent: "#282828",
    black: "#282828",
    red: "#cc241d",
    green: "#98971a",
    yellow: "#d79921",
    blue: "#458588",
    magenta: "#b16286",
    cyan: "#689d6a",
    white: "#a89984",
    brightBlack: "#928374",
    brightRed: "#fb4934",
    brightGreen: "#b8bb26",
    brightYellow: "#fabd2f",
    brightBlue: "#83a598",
    brightMagenta: "#d3869b",
    brightCyan: "#8ec07c",
    brightWhite: "#ebdbb2"
  },
  {
    name: "Tokyo Night",
    background: "#1a1b26",
    foreground: "#a9b1d6",
    cursor: "#c0caf5",
    cursorAccent: "#1a1b26",
    black: "#32344a",
    red: "#f7768e",
    green: "#9ece6a",
    yellow: "#e0af68",
    blue: "#7aa2f7",
    magenta: "#ad8ee6",
    cyan: "#449dab",
    white: "#787c99",
    brightBlack: "#444b6a",
    brightRed: "#ff7a93",
    brightGreen: "#b9f27c",
    brightYellow: "#ff9e64",
    brightBlue: "#7da6ff",
    brightMagenta: "#bb9af7",
    brightCyan: "#0db9d7",
    brightWhite: "#acb0d0"
  },
  {
    name: "Rose Pine",
    background: "#191724",
    foreground: "#e0def4",
    cursor: "#e0def4",
    cursorAccent: "#191724",
    black: "#191724",
    red: "#eb6f92",
    green: "#9ccfd8",
    yellow: "#f6c177",
    blue: "#31748f",
    magenta: "#c4a7e7",
    cyan: "#ebbcba",
    white: "#e0def4",
    brightBlack: "#555169",
    brightRed: "#eb6f92",
    brightGreen: "#9ccfd8",
    brightYellow: "#f6c177",
    brightBlue: "#31748f",
    brightMagenta: "#c4a7e7",
    brightCyan: "#ebbcba",
    brightWhite: "#e0def4"
  }
];

  // Current theme state
  let selectedPreset = 0;
  let customTheme = { ...themePresets[0] };
  let useCustomTheme = false;

  // Reactive statement to update terminal theme when custom theme changes
  $: if (useCustomTheme) {
    terminalTheme.set(customTheme);
  }

  // Known hosts state
  let knownHosts: Array<{
    hostname: string;
    algorithm: string;
    fingerprint: string;
    addedDate: string;
  }> = [];

  // Account state
  let accountSettings = {
    username: 'user',
    email: '',
    theme: 'dark',
    autoSave: true,
    notifications: true
  };

  // Load known hosts from backend
  onMount(async () => {
    if (show) {
      await loadKnownHosts();
    }
  });

  // Reload known hosts when dialog is shown
  $: if (show) {
    loadKnownHosts();
  }

  async function loadKnownHosts() {
    try {
      const hosts = await App.GetKnownHosts();
      knownHosts = hosts.map(host => ({
        hostname: host.hostname,
        algorithm: host.algorithm,
        fingerprint: host.fingerprint,
        addedDate: host.addedDate
      }));
    } catch (error) {
      console.error('Failed to load known hosts:', error);
    }
  }

  function selectPreset(index: number) {
    selectedPreset = index;
    customTheme = { ...themePresets[index] };
    useCustomTheme = false;
    
    // Update the terminal theme store
    terminalTheme.set(themePresets[index]);
  }

  function enableCustomTheme() {
    useCustomTheme = true;
  }

  function resetToPreset() {
    customTheme = { ...themePresets[selectedPreset] };
    useCustomTheme = false;
    
    // Update the terminal theme store
    terminalTheme.set(themePresets[selectedPreset]);
  }

  async function removeKnownHost(index: number) {
    const host = knownHosts[index];
    try {
      await App.RemoveKnownHost(host.hostname);
      knownHosts = knownHosts.filter((_, i) => i !== index);
    } catch (error) {
      console.error('Failed to remove known host:', error);
    }
  }

  async function clearAllKnownHosts() {
    if (confirm('Are you sure you want to remove all known hosts? This will require you to verify host keys again on next connection.')) {
      try {
        await App.ClearKnownHosts();
        knownHosts = [];
      } catch (error) {
        console.error('Failed to clear known hosts:', error);
      }
    }
  }

  function saveSettings() {
    const settings = {
      terminal: {
        theme: useCustomTheme ? customTheme : themePresets[selectedPreset],
        useCustomTheme,
        selectedPreset
      },
      knownHosts,
      account: accountSettings
    };
    
    dispatch('saveSettings', settings);
    dispatch('close');
  }

  function handleClose() {
    dispatch('close');
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      handleClose();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
    <div class="bg-slate-800 rounded-lg border border-slate-700 w-full max-w-4xl max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-slate-700">
        <h2 class="text-xl font-semibold text-white">Settings</h2>
        <button 
          on:click={handleClose}
          class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
        >
          <X size={20} />
        </button>
      </div>

      <div class="flex flex-1 min-h-0">
        <!-- Sidebar -->
        <div class="w-64 bg-slate-750 border-r border-slate-700 p-4">
          <nav class="space-y-2">
            <button
              class="w-full flex items-center gap-3 px-3 py-2 text-left rounded-md transition-colors {activeTab === 'terminal' ? 'bg-slate-600 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'}"
              on:click={() => activeTab = 'terminal'}
            >
              <Terminal size={18} />
              Terminal Themes
            </button>
            <button
              class="w-full flex items-center gap-3 px-3 py-2 text-left rounded-md transition-colors {activeTab === 'known-hosts' ? 'bg-slate-600 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'}"
              on:click={() => activeTab = 'known-hosts'}
            >
              <Shield size={18} />
              Known Hosts
            </button>
            <button
              class="w-full flex items-center gap-3 px-3 py-2 text-left rounded-md transition-colors {activeTab === 'account' ? 'bg-slate-600 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'}"
              on:click={() => activeTab = 'account'}
            >
              <User size={18} />
              Account
            </button>
          </nav>
        </div>

        <!-- Content -->
  <div class="flex-1 p-6 overflow-y-auto custom-scrollbar" style="overflow-x: hidden;">
          {#if activeTab === 'terminal'}
            <div class="space-y-6">
              <div>
                <h3 class="text-lg font-medium text-white mb-4">Terminal Appearance</h3>
                
                <!-- Theme Presets -->
                <div class="mb-6">
                  <h4 class="text-sm font-medium text-slate-300 mb-3">Theme Presets</h4>
                  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                    {#each themePresets as preset, index}
                      <button
                        class="p-3 rounded-lg border transition-all {selectedPreset === index && !useCustomTheme ? 'border-blue-500 bg-blue-500/10' : 'border-slate-600 hover:border-slate-500'}"
                        on:click={() => selectPreset(index)}
                      >
                        <div class="text-left">
                          <div class="text-sm font-medium text-white mb-2">{preset.name}</div>
                          <div class="flex gap-1 mb-2">
                            <div class="w-3 h-3 rounded-full" style="background-color: {preset.background}"></div>
                            <div class="w-3 h-3 rounded-full" style="background-color: {preset.foreground}"></div>
                            <div class="w-3 h-3 rounded-full" style="background-color: {preset.red}"></div>
                            <div class="w-3 h-3 rounded-full" style="background-color: {preset.green}"></div>
                            <div class="w-3 h-3 rounded-full" style="background-color: {preset.blue}"></div>
                          </div>
                          <div 
                            class="text-xs p-2 rounded font-mono"
                            style="background-color: {preset.background}; color: {preset.foreground};"
                          >
                            user@host:~$
                          </div>
                        </div>
                      </button>
                    {/each}
                  </div>
                </div>

                <!-- Custom Theme Section -->
                <div class="border-t border-slate-700 pt-6">
                  <div class="flex items-center justify-between mb-4">
                    <h4 class="text-sm font-medium text-slate-300">Custom Theme</h4>
                    <div class="flex gap-2">
                      <button
                        class="px-3 py-1 text-xs bg-slate-700 text-slate-300 rounded hover:bg-slate-600 transition-colors"
                        on:click={resetToPreset}
                      >
                        <RotateCcw size={14} class="inline mr-1" />
                        Reset
                      </button>
                      <button
                        class="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                        on:click={enableCustomTheme}
                      >
                        <Palette size={14} class="inline mr-1" />
                        Customize
                      </button>
                    </div>
                  </div>

                  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Background</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.background}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.background}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Foreground</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.foreground}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.foreground}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Cursor</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.cursor}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.cursor}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <!-- Standard Colors -->
                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Red</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.red}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.red}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Green</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.green}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.green}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Blue</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.blue}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.blue}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-xs text-slate-400 mb-1">Yellow</label>
                      <div class="flex items-center gap-2">
                        <input
                          type="color"
                          bind:value={customTheme.yellow}
                          on:input={enableCustomTheme}
                          class="w-8 h-8 rounded border border-slate-600"
                        />
                        <input
                          type="text"
                          bind:value={customTheme.yellow}
                          on:input={enableCustomTheme}
                          class="flex-1 px-2 py-1 text-xs bg-slate-700 border border-slate-600 rounded text-white"
                        />
                      </div>
                    </div>
                  </div>

                  <!-- Preview -->
                  <div class="mt-4">
                    <label class="block text-xs text-slate-400 mb-2">Preview</label>
                    <div 
                      class="p-4 rounded-lg font-mono text-sm"
                      style="background-color: {customTheme.background}; color: {customTheme.foreground};"
                    >
                      <div style="color: {customTheme.green};">user@hostname:~$</div>
                      <div style="color: {customTheme.blue};">ls -la</div>
                      <div style="color: {customTheme.red};">-rw-r--r-- 1 user user 1024 Sep 15 10:30 file.txt</div>
                      <div style="color: {customTheme.yellow};">drwxr-xr-x 2 user user 4096 Sep 15 10:31 directory</div>
                      <div>total 5120</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

          {:else if activeTab === 'known-hosts'}
            <div class="space-y-6">
              <div class="flex items-center justify-between">
                <h3 class="text-lg font-medium text-white">Known Hosts</h3>
                <button
                  class="px-3 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
                  on:click={clearAllKnownHosts}
                >
                  <Trash2 size={16} class="inline mr-2" />
                  Clear All
                </button>
              </div>

              <p class="text-sm text-slate-400">
                Manage SSH host keys that have been previously accepted. Removing a host will require you to verify its key again on the next connection.
              </p>

              {#if knownHosts.length === 0}
                <div class="text-center py-8 text-slate-400">
                  <Shield size={48} class="mx-auto mb-4 opacity-50" />
                  <p>No known hosts</p>
                  <p class="text-sm">Host keys will appear here after you accept them during SSH connections.</p>
                </div>
              {:else}
                <div class="space-y-3">
                  {#each knownHosts as host, index}
                    <div class="bg-slate-750 rounded-lg p-4 border border-slate-700">
                      <div class="flex items-start justify-between">
                        <div class="flex-1">
                          <div class="flex items-center gap-3 mb-2">
                            <h4 class="font-medium text-white">{host.hostname}</h4>
                            <span class="px-2 py-1 text-xs bg-slate-600 text-slate-300 rounded">
                              {host.algorithm}
                            </span>
                          </div>
                          <div class="text-sm text-slate-400 mb-1">
                            <strong>Fingerprint:</strong>
                          </div>
                          <div class="font-mono text-xs bg-slate-800 p-2 rounded border border-slate-600 text-slate-300 break-all">
                            {host.fingerprint}
                          </div>
                          <div class="text-xs text-slate-500 mt-2">
                            Added: {host.addedDate}
                          </div>
                        </div>
                        <button
                          class="ml-4 p-2 text-slate-400 hover:text-red-400 hover:bg-slate-700 rounded transition-colors"
                          on:click={() => removeKnownHost(index)}
                          title="Remove this host key"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>

          {:else if activeTab === 'account'}
            <div class="space-y-6">
              <h3 class="text-lg font-medium text-white">Account Settings</h3>
              
              <div class="bg-slate-750 rounded-lg p-6 border border-slate-700">
                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-slate-300 mb-2">Username</label>
                    <input
                      type="text"
                      bind:value={accountSettings.username}
                      disabled
                      class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white disabled:opacity-50 disabled:cursor-not-allowed"
                      placeholder="Enter username"
                    />
                    <p class="text-xs text-slate-500 mt-1">Username management will be available in a future update</p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-slate-300 mb-2">Email</label>
                    <input
                      type="email"
                      bind:value={accountSettings.email}
                      disabled
                      class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white disabled:opacity-50 disabled:cursor-not-allowed"
                      placeholder="Enter email address"
                    />
                    <p class="text-xs text-slate-500 mt-1">Email integration coming soon</p>
                  </div>

                  <div class="border-t border-slate-700 pt-4">
                    <h4 class="text-sm font-medium text-slate-300 mb-3">Preferences</h4>
                    
                    <div class="space-y-3">
                      <label class="flex items-center gap-3">
                        <input
                          type="checkbox"
                          bind:checked={accountSettings.autoSave}
                          disabled
                          class="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded disabled:opacity-50"
                        />
                        <span class="text-sm text-slate-300">Auto-save settings</span>
                      </label>

                      <label class="flex items-center gap-3">
                        <input
                          type="checkbox"
                          bind:checked={accountSettings.notifications}
                          disabled
                          class="w-4 h-4 text-blue-600 bg-slate-700 border-slate-600 rounded disabled:opacity-50"
                        />
                        <span class="text-sm text-slate-300">Enable notifications</span>
                      </label>
                    </div>
                  </div>

                  <div class="bg-blue-900/20 border border-blue-800 rounded-lg p-4">
                    <div class="flex items-start gap-3">
                      <User size={20} class="text-blue-400 flex-shrink-0 mt-0.5" />
                      <div>
                        <h5 class="text-sm font-medium text-blue-300 mb-1">Account Features Coming Soon</h5>
                        <p class="text-xs text-blue-200">
                          User accounts, cloud sync, and advanced preferences will be available in future updates.
                          For now, all settings are stored locally.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="border-t border-slate-700 p-6 flex justify-end gap-3">
        <button
          class="px-4 py-2 text-slate-300 border border-slate-600 rounded-md hover:bg-slate-700 transition-colors"
          on:click={handleClose}
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors flex items-center gap-2"
          on:click={saveSettings}
        >
          <Save size={16} />
          Save Settings
        </button>
      </div>
    </div>
  </div>
{/if}
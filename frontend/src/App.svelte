<script lang="ts">
  import { onMount } from 'svelte';
  
  // Import our TypeScript stores and types
  import {
    hosts,
    macros,
    activeSessions,
    activeTab,
    showHostModal,
    addNotification,
    sessionCache,
    sessionSFTPLayouts
  } from './types/stores';

  let editingHost: Host | null = null;
  // Global sessionRawOutput cache (Map<sessionId, string>)
  export const sessionRawOutput = new Map<string, string>();
  
  import type { Host, Macro } from './types/api';
  
  // Import Wails App API
  import * as App from '../wailsjs/go/main/App';
  import { SessionAPI, MacroAPI, HostAPI } from './lib/api';
  
  // Import components
  import Sidebar from './components/Sidebar.svelte';
  import Terminal from './components/Terminal.svelte';
  import SFTPPanel from './components/SFTPPanel.svelte';
  import RightSidebar from './components/RightSidebar.svelte';
  import HostForm from './components/HostForm.svelte';
  import ConnectionDialog from './components/ConnectionDialog.svelte';
  import SettingsModal from './components/Settings.svelte';
  import TerminalTabs from './components/TerminalTabs.svelte';
  
  // Icons from lucide-svelte
  import { 
    Settings, 
    FolderOpen, 
    Bell, 
    Wifi, 
    WifiOff
  } from 'lucide-svelte';

  // State
  // Ping tracking
  let lastPing: number | null = null;
  let pingDisplay: string = '';
  let pingInterval: any = null;

  // Poll backend for ping every 2s for active session
  $: if (activeSession && activeSession.id) {
    if (pingInterval) clearInterval(pingInterval);
    pingInterval = setInterval(async () => {
      const ping = await SessionAPI.getPing(activeSession.id);
      lastPing = ping;
      pingDisplay = lastPing !== null && lastPing > 0 ? `${lastPing} ms` : '';
    }, 2000);
  } else if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
    lastPing = null;
    pingDisplay = '';
  }
  let leftSidebarCollapsed = false;
  let rightSidebarCollapsed = true;
  let forceRerender = 0;
  let sidebar: Sidebar; // Reference to sidebar component for error handling
  let terminal: Terminal; // Reference to terminal component for dimensions
  let activeSession: any; // Active session object
  let sftpToggleTimeout: number | null = null; // Debounce timeout for SFTP toggle


  // Connection dialog state
  let showConnectionDialog = false;
  let connectingHost: Host | null = null;

  type ConnectionStep = {
    id: string;
    title: string;
    status: 'pending' | 'in-progress' | 'completed' | 'error' | 'warning';
    message?: string;
  };

  let connectionSteps: ConnectionStep[] = [
    { id: 'resolving', title: 'Resolving hostname', status: 'pending' },
    { id: 'connecting', title: 'Establishing connection', status: 'pending' },
    { id: 'handshake', title: 'SSH handshake', status: 'pending' },
    { id: 'hostkey', title: 'Verifying host key', status: 'pending' },
    { id: 'auth', title: 'Authenticating', status: 'pending' },
    { id: 'session', title: 'Starting session', status: 'pending' }
  ];
  let hostKeyInfo: { fingerprint: string; algorithm: string; publicKey?: string; hostname?: string; isNewHost: boolean } | null = null;
  let showHostKeyDialog = false;

  // Settings state
  let showSettings = false;

  // Load initial data
  onMount(async () => {
    console.log('=== APP MOUNT START ===');
    console.log('App mounted, loading initial data...');
    
    // Add detailed environment logging
    console.log('Window object check:', {
      hasWindow: typeof window !== 'undefined',
      hasGo: typeof window !== 'undefined' && !!window.go,
      hasMain: typeof window !== 'undefined' && !!window.go?.main,
      hasApp: typeof window !== 'undefined' && !!window.go?.main?.App,
      windowGoKeys: typeof window !== 'undefined' && window.go ? Object.keys(window.go) : 'no window.go'
    });
    
    try {
      // Load hosts
      console.log('=== LOADING HOSTS ===');
      const loadedHosts = await HostAPI.getAll();
      console.log('Loaded hosts count:', loadedHosts.length);
      console.log('Loaded hosts data:', loadedHosts);
      hosts.set(loadedHosts);
      
      // Load macros
      console.log('=== LOADING MACROS ===');
      const loadedMacros = await MacroAPI.getAll();
      console.log('Loaded macros count:', loadedMacros.length);
      console.log('Loaded macros data:', loadedMacros);
      macros.set(loadedMacros);
      
      // Load active sessions
      console.log('=== LOADING SESSIONS ===');
      const loadedSessions = await SessionAPI.getActiveSessions();
      console.log('Loaded sessions count:', loadedSessions.length);
      console.log('Loaded sessions data:', loadedSessions);
      // Populate host data for sessions and filter out invalid ones
      const sessionsWithHosts = loadedSessions
        .map(session => {
          console.log('Processing session:', session.id, 'hostId:', session.hostId);
          const host = loadedHosts.find(h => h.id === session.hostId);
          console.log('Found host for session:', session.id, ':', host?.label || 'NOT FOUND');
          return { ...session, host };
        })
        .filter(session => {
          // Only keep sessions that have a valid hostId and found host
          const isValid = session.hostId && session.host;
          if (!isValid) {
            console.log('Filtering out invalid session:', session.id, 'hostId:', session.hostId, 'hasHost:', !!session.host);
          }
          return isValid;
        });
      console.log('Valid sessions with hosts populated:', sessionsWithHosts);
      activeSessions.set(sessionsWithHosts);
      console.log('=== INITIAL DATA LOADING COMPLETED ===');
      
    } catch (error) {
      console.error('=== FAILED TO LOAD INITIAL DATA ===');
      console.error('Error details:', error);
      if (error instanceof Error) {
        console.error('Error stack:', error.stack);
      }
      addNotification({
        type: 'error',
        title: 'Failed to load application data'
      });
    }
  });

  // Event handlers
  function handleHostSelect(event: CustomEvent<Host>) {
    const host = event.detail;
    connectToHost(host);
  }

  function handleAddHost() {
    
    showHostModal.set(true);
    
    // Force component re-render for WebView2
    forceRerender++;
    
    setTimeout(() => {
      showHostModal.set(true);
      forceRerender++;
      console.log('Forced re-render completed, new forceRerender:', forceRerender);
    }, 10);
  }

  function handleEditHost(event: CustomEvent<Host>) {
    editingHost = event.detail;
    showHostModal.set(true);
    
    // Force component re-render for WebView2
    forceRerender++;
    
    setTimeout(() => {
      showHostModal.set(true);
      forceRerender++;
      console.log('Forced re-render completed, new forceRerender:', forceRerender);
    }, 10);
  }

  function handleHostFormClose() {
    showHostModal.set(false);
    editingHost = null
  }

  function handleLeftSidebarToggle(event: CustomEvent<boolean>) {
    leftSidebarCollapsed = event.detail;
  }

  function handleRightSidebarToggle(event: CustomEvent<boolean>) {
    rightSidebarCollapsed = event.detail;
  }

  function handleSessionClosed(event: CustomEvent<string>) {
    const sessionId = event.detail;
    const closedSession = $activeSessions.find(s => s.id === sessionId);
    
    activeSessions.update(sessions => {
      const filtered = sessions.filter(s => s.id !== sessionId);
      
      // Renumber sessions for the same host
      if (closedSession) {
        const hostSessions = filtered.filter(s => s.hostId === closedSession.hostId);
        hostSessions.forEach((session, index) => {
          const sessionNumber = index + 1;
          session.displayName = sessionNumber === 1 ? 
            session.host?.label || 'Unknown Host' : 
            `${session.host?.label || 'Unknown Host'} (${sessionNumber})`;
        });
      }
      
      return filtered;
    });
    
    // Clean up session cache
    sessionCache.delete(sessionId);
    
    // Clean up sessionRawOutput cache
    sessionRawOutput.delete(sessionId);
    
    // Clean up SFTP layout for this session
    sessionSFTPLayouts.delete(sessionId);
    
    // If this was the active tab, clear it or switch to another
    if ($activeTab === sessionId) {
      const remainingSessions = $activeSessions.filter(s => s.id !== sessionId);
      activeTab.set(remainingSessions.length > 0 ? remainingSessions[0].id : null);
    }
  }

  function handleSessionActivated(event: CustomEvent<string>) {
    activeTab.set(event.detail);
  }

  function handleMacroExecute(event: CustomEvent<Macro>) {
    const macro = event.detail;
    executeMacro(macro);
  }

  function toggleSFTP() {
    if (!activeSession) return;
    
    const currentLayout = sessionSFTPLayouts.get(activeSession.id);
    console.log('toggleSFTP called, current layout for session', activeSession.id, ':', currentLayout);
    
    // Debounce rapid clicks
    if (sftpToggleTimeout) {
      console.log('toggleSFTP: Ignoring rapid click (debounced)');
      return;
    }
    
    sftpToggleTimeout = setTimeout(() => {
      sftpToggleTimeout = null;
    }, 500); // 500ms debounce
    
    const newLayout = currentLayout === 'hidden' ? 'fullscreen' : 'hidden';
    sessionSFTPLayouts.set(activeSession.id, newLayout);
    console.log('toggleSFTP completed, new layout for session', activeSession.id, ':', newLayout);
    
    // Force reactivity update
    activeSession = activeSession;
  }

  function selectSession(sessionId: string) {
    activeTab.set(sessionId);
  }

  function closeSession(sessionId: string) {
    handleSessionClosed({ detail: sessionId } as CustomEvent<string>);
  }

  function resetConnectionState() {
    console.log('Resetting connection state');
    showConnectionDialog = false;
    showHostKeyDialog = false;
    connectingHost = null;
    hostKeyInfo = null;
    connectionSteps = connectionSteps.map(step => ({ ...step, status: 'pending' as const, message: undefined }));
  }

  async function connectToHost(host: Host) {
    console.log('connectToHost called for:', host.label, 'showConnectionDialog:', showConnectionDialog);
    
    // Prevent multiple concurrent connections
    if (showConnectionDialog) {
      console.log('Connection already in progress, ignoring duplicate request');
      return;
    }

    // Reset connection state
    resetConnectionState();
    connectingHost = host;
    showConnectionDialog = true;
    
    console.log('Starting connection to:', host.label);

    try {
      await performConnection(host);
    } catch (error) {
      console.error('Connection failed:', error);
      handleConnectionError(host, error);
    }
  }

  async function performConnection(host: Host) {
 
    updateConnectionStep('resolving', 'in-progress', 'Looking up hostname...');
    await new Promise(resolve => setTimeout(resolve, 500)); // Simulate delay
    updateConnectionStep('resolving', 'completed', 'Hostname resolved');

  
    updateConnectionStep('connecting', 'in-progress', 'Connecting to server...');
    
    try {
  
      updateConnectionStep('connecting', 'completed');
      updateConnectionStep('handshake', 'in-progress', 'Negotiating encryption...');
      
   
      updateConnectionStep('handshake', 'completed');
      updateConnectionStep('hostkey', 'in-progress', 'Checking host key...');
      

      const dimensions = terminal?.getTerminalDimensions?.() || { cols: 80, rows: 24 };
      console.log('Calling ConnectSSHWithHostKeyVerification for host:', host.label);
      const result = await App.ConnectSSHWithHostKeyVerification(host.id, dimensions.cols, dimensions.rows);
      console.log('ConnectSSHWithHostKeyVerification result:', result);
      
      if (result.needsHostKeyVerification) {
        console.log('Host key verification needed, showing dialog');
   
        hostKeyInfo = {
          fingerprint: result.hostKeyInfo.fingerprint,
          algorithm: result.hostKeyInfo.algorithm,
          publicKey: result.hostKeyInfo.publicKey,
          hostname: result.hostKeyInfo.hostname,
          isNewHost: result.hostKeyInfo.isNewHost
        };
        showHostKeyDialog = true;
        updateConnectionStep('hostkey', 'warning', 'Host key verification required');
        
 
        return;
      } else {
        console.log('Host key already verified, continuing with session setup');
        updateConnectionStep('hostkey', 'completed', 'Host key verified');
        await finalizeConnection(host, result.sessionId);
      }
      
    } catch (error) {
      console.error('Error in performConnection:', error);
      throw error;
    }
  }

  async function continueConnection(host: Host) {
    updateConnectionStep('auth', 'in-progress', 'Authenticating user...');
    
    const connectionPromise = (async (): Promise<string> => {
      if (terminal && terminal.getTerminalDimensions) {
        const dimensions = terminal.getTerminalDimensions();
        if (dimensions) {
          return await SessionAPI.connectWithDimensions(host.id, dimensions.cols, dimensions.rows);
        } else {
          return await SessionAPI.connect(host.id);
        }
      } else {
        return await SessionAPI.connect(host.id);
      }
    })();

    const sessionId = await connectionPromise;
    await finalizeConnection(host, sessionId);
  }

  async function finalizeConnection(host: Host, sessionId: string) {
    updateConnectionStep('auth', 'completed', 'Authentication successful');
    
    updateConnectionStep('session', 'in-progress', 'Starting terminal session...');
    await new Promise(resolve => setTimeout(resolve, 500)); // Brief delay
    updateConnectionStep('session', 'completed', 'Session established');
    
    // Close dialog after a brief success display
    setTimeout(() => {
      resetConnectionState();
    }, 1000);
    
    // Generate display name for numbered sessions
    const existingSessions = $activeSessions.filter(s => s.hostId === host.id);
    const sessionNumber = existingSessions.length + 1;
    const displayName = sessionNumber === 1 ? host.label : `${host.label} (${sessionNumber})`;
    
    const newSession = {
      id: sessionId,
      hostId: host.id,
      host: host,
      isActive: true,
      createdAt: new Date().toISOString(),
      displayName: displayName
    };
    
    activeSessions.update(sessions => [...sessions, newSession]);
    activeTab.set(sessionId);
    
    // Notify sidebar of successful connection
    if (sidebar) {
      sidebar.onConnectionSuccess(host.id);
    }
    
    addNotification({
      type: 'success',
      title: `Connected to ${host.label}`,
      message: `Session ${sessionId} established`
    });
    
    // Check for initial output and ensure terminal is ready
    setTimeout(async () => {
      try {
        const output = await App.GetSessionOutput(sessionId);
        console.log('Initial session output check:', { sessionId, outputLength: output?.length || 0 });
        
        if (output && output.length > 0) {
          // Store in sessionRawOutput cache
          sessionRawOutput.set(sessionId, output);
          
          let retries = 0;
          const maxRetries = 20;
          const writeOutput = () => {
            if (terminal && terminal.writeToTerminal && terminal.writeToTerminal(output)) {
              sessionCache.set(sessionId, output);
              console.log('Successfully wrote initial output to terminal');
              return;
            } else if (retries < maxRetries) {
              retries++;
              setTimeout(writeOutput, 250);
            } else {
              sessionCache.set(sessionId, output);
              console.log('Max retries reached, output stored in cache');
            }
          };
          writeOutput();
        } 
        await App.CheckSessionHealth(sessionId);
      } catch (healthError) {
        console.error('Session health check failed:', healthError);
        addNotification({
          type: 'warning',
          title: `Connection established but session may be unresponsive`,
          message: 'Try sending a command or reconnecting if needed'
        });
      }
    }, 1500);
  }

  function updateConnectionStep(stepId: string, status: 'pending' | 'in-progress' | 'completed' | 'error' | 'warning', message?: string) {
    connectionSteps = connectionSteps.map(step => 
      step.id === stepId ? { ...step, status, message } : step
    );
  }

  function handleConnectionError(host: Host, error: any) {
    resetConnectionState();
    
    // Update the step that failed
    const failedStep = connectionSteps.find(step => step.status === 'in-progress');
    if (failedStep) {
      updateConnectionStep(failedStep.id, 'error', error.message || 'Connection failed');
    }
    
    let errorMessage = 'Unknown connection error';
    if (error instanceof Error) {
      errorMessage = error.message;
    } else if (typeof error === 'string') {
      errorMessage = error;
    }
    
    if (sidebar) {
      sidebar.onConnectionError(host.id, errorMessage);
    }
    
    addNotification({
      type: 'error',
      title: `Failed to connect to ${host.label}`,
      message: errorMessage
    });
  }

  async function handleHostKeyAccept() {
    if (!connectingHost || !hostKeyInfo) return;
    
    try {
  // Save the host key to known_hosts using the marshaled public key provided by backend
  await App.AcceptHostKey(hostKeyInfo.hostname || connectingHost.hostname, hostKeyInfo.publicKey || (hostKeyInfo.algorithm + ' ' + hostKeyInfo.fingerprint));
      
      showHostKeyDialog = false;
      updateConnectionStep('hostkey', 'completed', 'Host key accepted and saved');
      
      await continueConnection(connectingHost);
    } catch (error) {
      console.error('Failed to accept host key:', error);
      addNotification({
        type: 'error',
        title: 'Failed to save host key',
        message: error instanceof Error ? error.message : 'Unknown error occurred'
      });
      handleConnectionError(connectingHost, error);
    }
  }

  function handleHostKeyReject() {
    const hostLabel = connectingHost?.label || 'host';
    resetConnectionState();
    
    addNotification({
      type: 'info',
      title: 'Connection cancelled',
      message: `Host key verification rejected for ${hostLabel}`
    });
  }

  function handleConnectionDialogClose() {
    if (showHostKeyDialog) {
      handleHostKeyReject();
    } else {
      resetConnectionState();
    }
  }

  function handleSettingsOpen() {
    showSettings = true;
  }

  function handleSettingsClose() {
    showSettings = false;
  }

  function handleSettingsSave(event: CustomEvent<any>) {
    const settings = event.detail;
    console.log('"Saving" settings:', settings);
    
    // TODO: Save settings
    
    addNotification({
      type: 'success',
      title: 'Settings saved',
      message: 'Your preferences have been updated'
    });
  }

  async function executeMacro(macro: Macro) {
    if (!$activeTab) {
      addNotification({
        type: 'warning',
        title: 'No active session to execute macro'
      });
      return;
    }

    try {
      await MacroAPI.execute($activeTab, macro.id);
      addNotification({
        type: 'success',
        title: `Executed macro: ${macro.label}`
      });
    } catch (error) {
      console.error('Failed to execute macro:', error);
      addNotification({
        type: 'error',
        title: `Failed to execute macro: ${macro.label}`
      });
    }
  }

  // Reactive statements
  $: activeSession = $activeSessions.find(s => s.id === $activeTab);
  $: currentSFTPLayout = activeSession ? sessionSFTPLayouts.get(activeSession.id) : 'hidden';
</script>

<div class="h-screen flex flex-col bg-slate-900 text-white">
  <!-- Header Bar -->
  <div class="bg-slate-800 border-b border-slate-700 px-4 py-2 flex items-center justify-between flex-shrink-0">
    <div class="flex items-center gap-4">
      <h1 class="text-xl font-bold">Termunator</h1>
      {#if activeSession}
        <div class="flex items-center gap-2 text-sm text-slate-400">
          <div class="flex items-center gap-1">
            {#if activeSession.isActive}
              <span title={pingDisplay ? `Ping: ${pingDisplay}` : 'Connected'}>
                <Wifi size={16} class="text-green-400" />
              </span>
            {:else}
              <WifiOff size={16} class="text-red-400" />
            {/if}
            Connected to {activeSession.host?.label}
          </div>
        </div>
      {/if}
    </div>
    
    <div class="flex items-center gap-2">
      <!-- SFTP Toggle -->
      {#if activeSession}
        <button 
          class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors {currentSFTPLayout !== 'hidden' ? 'bg-slate-700 text-white' : ''}"
          on:click={toggleSFTP}
          title="Toggle SFTP Panel"
        >
          <FolderOpen size={16} />
        </button>
      {/if}
      
      <!-- Notifications -->
      <button class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors">
        <Bell size={16} />
      </button>
      
      <!-- Settings -->
      <button 
        class="p-2 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"
        on:click={handleSettingsOpen}
        title="Settings"
      >
        <Settings size={16} />
      </button>
    </div>
  </div>

  <!-- Main Layout -->
  <div class="flex-1 flex overflow-hidden">
    <!-- Left Sidebar -->
    <Sidebar 
      bind:this={sidebar}
      hosts={$hosts}
      collapsed={leftSidebarCollapsed}
      on:hostSelect={handleHostSelect}
      on:addHost={handleAddHost}
      on:editHost={handleEditHost}
      on:toggleCollapse={handleLeftSidebarToggle}
    />

    <!-- Main Content Area -->
  <div class="flex-1 flex flex-col min-h-0 min-w-0 overflow-hidden flex-shrink-0">
      
      {#if $showHostModal}
        {#key forceRerender}
          <HostForm 
            host={editingHost}
            on:close={handleHostFormClose}
            on:saved={handleHostFormClose}
          />
        {/key}
      {/if}
        <!-- Terminal and SFTP Area with flexible layouts -->
  <div class="flex-1 flex flex-col min-h-0 h-full">
          <!-- Session Tabs - Always at top -->
          {#if $activeSessions.length > 0}
              <TerminalTabs
                sessions={$activeSessions}
                activeSessionId={$activeTab}
                sessionErrors={new Map()}
                on:sessionClosed={event => closeSession(event.detail)}
                on:sessionActivated={event => selectSession(event.detail)}
              />
            {/if}



          <!-- Content Area - Terminal and SFTP -->
          <div class="flex-1 flex flex-col min-h-0 h-full overflow-hidden">
            {#if currentSFTPLayout === 'fullscreen'}
              <!-- SFTP Fullscreen -->
              <SFTPPanel 
                activeSession={activeSession}
                layout={currentSFTPLayout}
                on:layoutChange={(e) => {
                  if (activeSession) {
                    sessionSFTPLayouts.set(activeSession.id, e.detail);
                    activeSession = activeSession; // Force reactivity
                    if (terminal && terminal.triggerResize) {
                      setTimeout(() => terminal.triggerResize(), 200);
                    }
                  }
                }}
                on:close={() => {
                  if (activeSession) {
                    sessionSFTPLayouts.set(activeSession.id, 'hidden');
                    activeSession = activeSession; // Force reactivity
                    if (terminal && terminal.triggerResize) {
                      setTimeout(() => terminal.triggerResize(), 200);
                    }
                  }
                }}
              />
            {:else if currentSFTPLayout === 'top'}
              <!-- SFTP on top, terminal on bottom -->
              <div class="flex flex-col flex-1 min-h-0 h-full">
                <div class="flex-1 min-h-0 h-full border-b border-slate-700">
                  <SFTPPanel 
                    activeSession={activeSession}
                    layout={currentSFTPLayout}
                    on:layoutChange={(e) => {
                      if (activeSession) {
                        sessionSFTPLayouts.set(activeSession.id, e.detail);
                        activeSession = activeSession; // Force reactivity
                        if (terminal && terminal.triggerResize) {
                          setTimeout(() => terminal.triggerResize(), 200);
                        }
                      }
                    }}
                    on:close={() => {
                      if (activeSession) {
                        sessionSFTPLayouts.set(activeSession.id, 'hidden');
                        activeSession = activeSession; // Force reactivity
                        if (terminal && terminal.triggerResize) {
                          setTimeout(() => terminal.triggerResize(), 200);
                        }
                      }
                    }}
                  />
                </div>
                <div class="flex-1 min-h-0 h-full">
                  <Terminal 

                    bind:this={terminal}
                    sessions={$activeSessions}
                    activeSessionId={$activeTab}
                    on:sessionClosed={handleSessionClosed}
                    on:sessionActivated={handleSessionActivated}
                    sessionRawOutput={sessionRawOutput}
                  />
                </div>
              </div>
            {:else if currentSFTPLayout === 'bottom'}
              <!-- Terminal on top, SFTP on bottom -->
              <div class="flex flex-col flex-1 min-h-0 h-full">
                <div class="flex-1 min-h-0 h-full border-b border-slate-700">
                  <Terminal 
                    bind:this={terminal}
                    sessions={$activeSessions}
                    activeSessionId={$activeTab}
                    on:sessionClosed={handleSessionClosed}
                    on:sessionActivated={handleSessionActivated}
                    sessionRawOutput={sessionRawOutput}
                  />
                </div>
                <div class="flex-1 min-h-0 h-full">
                  <SFTPPanel 
                    activeSession={activeSession}
                    layout={currentSFTPLayout}
                    on:layoutChange={(e) => {
                      if (activeSession) {
                        sessionSFTPLayouts.set(activeSession.id, e.detail);
                        activeSession = activeSession; // Force reactivity
                        if (terminal && terminal.triggerResize) {
                          setTimeout(() => terminal.triggerResize(), 200);
                        }
                      }
                    }}
                    on:close={() => {
                      if (activeSession) {
                        sessionSFTPLayouts.set(activeSession.id, 'hidden');
                        activeSession = activeSession; // Force reactivity
                        if (terminal && terminal.triggerResize) {
                          setTimeout(() => terminal.triggerResize(), 200);
                        }
                      }
                    }}
                  />
                </div>
              </div>
            {:else}
              <!-- Terminal only (SFTP hidden) -->
              <div class="flex-1 min-h-0 h-full">
                <Terminal 
                  bind:this={terminal}
                  sessions={$activeSessions}
                  activeSessionId={$activeTab}
                  on:sessionClosed={handleSessionClosed}
                  on:sessionActivated={handleSessionActivated}
                  sessionRawOutput={sessionRawOutput}
                />
              </div>
            {/if}
          </div>
        </div>
      
    </div>

    <!-- Right Sidebar -->
    <RightSidebar 
      macros={$macros}
      collapsed={rightSidebarCollapsed}
      on:macroExecute={handleMacroExecute}
      on:addMacro={() => addNotification({
        type: 'info',
        title: 'Macro creation will be implemented'
      })}
      on:toggleCollapse={handleRightSidebarToggle}
    />
  </div>
</div>

<!-- Connection Dialog -->
{#if showConnectionDialog && connectingHost}
  <ConnectionDialog
    host={connectingHost}
    steps={connectionSteps}
    {hostKeyInfo}
    {showHostKeyDialog}
    onAccept={handleHostKeyAccept}
    onReject={handleHostKeyReject}
    on:close={handleConnectionDialogClose}
    on:accept={handleHostKeyAccept}
    on:reject={handleHostKeyReject}
  />
{/if}

<!-- Settings Dialog -->
<SettingsModal
  show={showSettings}
  on:close={handleSettingsClose}
  on:saveSettings={handleSettingsSave}
/>

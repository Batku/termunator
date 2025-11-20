<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { TerminalIcon, AlertCircle, Loader2 } from "lucide-svelte";
  import { Terminal } from "xterm";
  import { FitAddon } from "@xterm/addon-fit";
  import { WebLinksAddon } from "@xterm/addon-web-links";
  import type { Session } from "../types/api";
  import {
    sessionCache,
    terminalTheme,
    type TerminalTheme,
  } from "../types/stores";
  import { activeSessions } from "../types/stores";
  import { get } from "svelte/store";
  import * as App from "../../wailsjs/go/main/App";

  // Props
  export let sessions: Session[] = [];
  export let activeSessionId: string | null = null;
  export let sessionRawOutput: Map<string, string>;

  let terminalElement: HTMLElement;
  let terminal: Terminal | null = null;
  let fitAddon: FitAddon | null = null;
  let outputPollingInterval: number | null = null;
  let currentSessionId: string | null = null;
  let resizeHandler: (() => void) | null = null;
  let resizeObserver: ResizeObserver | null = null; // For terminal element
  let scrollTimeout: number; // Timeout for hiding scrollbar doesn't actually work :3

  // Loading and error states
  let terminalReady = false;
  let terminalInitializing = false;
  let isConnecting = false;
  let connectionError: string | null = null;
  let sessionErrors: Map<string, string> = new Map();

  // Subscribe to theme changes ( only when changing while terminal is running )
  $: if (terminal && $terminalTheme) {
    updateTerminalTheme($terminalTheme);
  }

  function updateTerminalTheme(theme: TerminalTheme) {
    if (!terminal) return;

    terminal.options.theme = {
      background: theme.background,
      foreground: theme.foreground,
      cursor: theme.cursor,
      cursorAccent: theme.cursorAccent,
      black: theme.black,
      red: theme.red,
      green: theme.green,
      yellow: theme.yellow,
      blue: theme.blue,
      magenta: theme.magenta,
      cyan: theme.cyan,
      white: theme.white,
      brightBlack: theme.brightBlack,
      brightRed: theme.brightRed,
      brightGreen: theme.brightGreen,
      brightYellow: theme.brightYellow,
      brightBlue: theme.brightBlue,
      brightMagenta: theme.brightMagenta,
      brightCyan: theme.brightCyan,
      brightWhite: theme.brightWhite,
    };
  }

  // Export function to get terminal dimensions
  export function getTerminalDimensions() {
    if (fitAddon && terminal) {
      const dimensions = fitAddon.proposeDimensions();
      console.log(
        "getTerminalDimensions called, proposed dimensions:",
        dimensions
      );
      if (terminalElement) {
        const rect = terminalElement.getBoundingClientRect();
        console.log(
          "Terminal element dimensions:",
          rect.width,
          "x",
          rect.height
        );
      }
      return dimensions;
    }
    console.log(
      "getTerminalDimensions called but fitAddon or terminal not ready"
    );
    return null;
  }

  // Export function to trigger resize manually
  export function triggerResize() {
    console.log("Manual resize triggered");
    if (fitAddon && terminal) {
      setTimeout(() => {
        if (fitAddon) {
          fitAddon.fit();
          const dimensions = fitAddon.proposeDimensions();
          if (dimensions && activeSessionId) {
            try {
              App.ResizeTerminal(
                activeSessionId,
                dimensions.cols,
                dimensions.rows
              );
              console.log(
                "Manual resize completed:",
                dimensions.cols,
                "x",
                dimensions.rows
              );
            } catch (error) {
              console.error("Failed to manually resize terminal:", error);
            }
          }
        }
      }, 100);
    }
  }

  // Export function to write output directly to terminal
  export function writeToTerminal(output: string) {
    if (terminal && terminalReady) {
      console.log(
        "Writing output directly to terminal:",
        output.length,
        "bytes"
      );
      terminal.write(output);
      terminal.scrollToBottom();
      return true;
    } else {
      console.warn(
        "Cannot write to terminal - terminal not ready (terminal:",
        !!terminal,
        "ready:",
        terminalReady,
        ")"
      );
      return false;
    }
  }

  // Default theme as fallback
  const defaultTheme: TerminalTheme = {
    name: "Default",
    background: "#1e1e2e",
    foreground: "#cdd6f4",
    cursor: "#f5e0dc",
    cursorAccent: "#1e1e2e",
    black: "#181825",
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
    brightWhite: "#a6adc8",
  };

  function createTerminal() {
    //console.log("createTerminal called, terminalElement:", !!terminalElement);
    if (!terminalElement || !terminalElement.isConnected) {
      //console.log("meow"); debugging
      setTimeout(createTerminal, 50);
      return;
    }

    destroyTerminal(); // Always clean up previous terminal before creating a new one

    terminalInitializing = true;
    console.log("Creating new terminal instance...");

    // Get current theme from store when creating terminal (uses fallback default if none)
    let currentTheme: TerminalTheme | undefined;
    const unsubscribe = terminalTheme.subscribe((theme) => {
      currentTheme = theme;
    });
    unsubscribe(); // Immediately unsubscribe after getting current value as createTerminal only runs once*

    const themeToUse = currentTheme ?? defaultTheme;

    terminal = new Terminal({
      theme: {
        background: themeToUse.background,
        foreground: themeToUse.foreground,
        cursor: themeToUse.cursor,
        cursorAccent: themeToUse.cursorAccent,
        black: themeToUse.black,
        red: themeToUse.red,
        green: themeToUse.green,
        yellow: themeToUse.yellow,
        blue: themeToUse.blue,
        magenta: themeToUse.magenta,
        cyan: themeToUse.cyan,
        white: themeToUse.white,
        brightBlack: themeToUse.brightBlack,
        brightRed: themeToUse.brightRed,
        brightGreen: themeToUse.brightGreen,
        brightYellow: themeToUse.brightYellow,
        brightBlue: themeToUse.brightBlue,
        brightMagenta: themeToUse.brightMagenta,
        brightCyan: themeToUse.brightCyan,
        brightWhite: themeToUse.brightWhite,
      },
      // These settings might be made configurable later, especially font size
      fontFamily: 'JetBrains Mono, Consolas, Monaco, "Courier New", monospace',
      fontSize: 14,
      lineHeight: 1.2,
      cursorBlink: true,
      cursorStyle: "block",
      scrollback: 10000,
      convertEol: true,
      allowTransparency: false,
      disableStdin: false,
      allowProposedApi: true,
    });

    // sets up and runs FitAddon
    console.log("Terminal instance created, setting up addons...");
    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());
    console.log("Opening terminal in DOM element...");
    terminal.open(terminalElement);
    fitAddon.fit();
    console.log("Terminal opened and fitted to container");

    // Ensure terminal is fully initialized before marking as ready
    setTimeout(() => {
      terminalInitializing = false;
      terminalReady = true;
      console.log("Terminal is now ready for writing");

      // Set up scroll handler after terminal is fully ready (might not be working) TODO: fix
      function setupScrollHandler(retryCount = 0) {
        // Only proceed if terminalElement is still attached
        if (!terminalElement || !terminalElement.isConnected) return;
        const viewport = terminalElement.querySelector(".xterm-viewport");
        if (viewport) {
          console.log("Setting up scroll handler for viewport");
          const handleScroll = () => {
            terminalElement.classList.add("scrolling");
            if (scrollTimeout) clearTimeout(scrollTimeout);
            scrollTimeout = setTimeout(() => {
              terminalElement.classList.remove("scrolling");
            }, 1500);
          };
          viewport.addEventListener("scroll", handleScroll);
        } else if (retryCount < 20) {
          // Retry after a short delay if viewport is not yet available, up to 20 times (1 second) otherwise give up as it isn't important
          setTimeout(() => setupScrollHandler(retryCount + 1), 50);
        } else {
          console.warn(
            "Could not find xterm viewport for scroll handling after max retries"
          );
        }
      }
      setupScrollHandler();
    }, 100); // Small delay to ensure DOM is ready

    // Handle user input
    terminal.onData(async (data) => {
      if (activeSessionId) {
        try {
          await App.SendInput(activeSessionId, data);
        } catch (error) {
          console.error("Failed to send input:", error);
        }
      }
    });

    // Handle resize
    let resizeRetryCount = 0;
    const MAX_RESIZE_RETRIES = 40;
    const handleResize = () => {
      if (!fitAddon || !terminal || !terminalElement) return;
      if (!terminalElement.isConnected) {
        console.warn("Terminal element not attached to DOM, skipping fit");
        return;
      }
      const rect = terminalElement.getBoundingClientRect();
      console.log(
        "handleResize called. Container size:",
        rect.width,
        "x",
        rect.height
      );
      if (rect.width === 0 || rect.height === 0) {
        if (resizeRetryCount < MAX_RESIZE_RETRIES) {
          resizeRetryCount++;
          console.warn(
            "Terminal container not ready for fit (0 size), retrying... (attempt",
            resizeRetryCount,
            ")"
          );
          requestAnimationFrame(handleResize);
        } else {
          console.error(
            "Max resize retries reached, aborting fit. Container size:",
            rect.width,
            "x",
            rect.height
          );
        }
        return;
      }
      fitAddon.fit();
      // Only send resize if dimensions are valid
      if (activeSessionId) {
        const dimensions = fitAddon.proposeDimensions();
        if (
          dimensions &&
          Number.isFinite(dimensions.cols) &&
          Number.isFinite(dimensions.rows) &&
          dimensions.cols > 0 &&
          dimensions.rows > 0
        ) {
          console.log(
            "Proposed dimensions after fit:",
            dimensions.cols,
            "x",
            dimensions.rows
          );
          try {
            App.ResizeTerminal(
              activeSessionId,
              dimensions.cols,
              dimensions.rows
            );
            console.log(
              "Terminal resized:",
              dimensions.cols,
              "x",
              dimensions.rows
            );
            resizeRetryCount = 0;
          } catch (error) {
            console.error("Failed to resize terminal:", error);
          }
        } else {
          if (resizeRetryCount < MAX_RESIZE_RETRIES) {
            resizeRetryCount++;
            console.warn(
              "Proposed dimensions after fit are invalid (",
              dimensions,
              "), retrying... (attempt",
              resizeRetryCount,
              ")"
            );
            requestAnimationFrame(handleResize);
          } else {
            console.error(
              "Max resize retries reached, aborting fit. Last proposed dimensions:",
              dimensions
            );
          }
        }
      }
    };

    window.addEventListener("resize", handleResize);

    // Create ResizeObserver to detect container size changes
    if (terminalElement && window.ResizeObserver) {
      resizeObserver = new ResizeObserver(() => {
        // Immediate resize for container changes
        console.log("Terminal container resized, triggering immediate fit");
        handleResize();
      });
      resizeObserver.observe(terminalElement);
    }


    // Handle scrolling with mouse wheel and keyboard shortcuts TODO: make configurable in settings and ctrl + shift + x to not collide with terminal apps
    terminal.attachCustomKeyEventHandler((event) => {
      // Handle Ctrl+C, Ctrl+V for copy/paste
      if (event.ctrlKey || event.metaKey) {
        switch (event.key) {
          case "c":
            if (event.type === "keydown" && terminal) {
              // Check if there's a selection
              if (terminal.hasSelection()) {
                navigator.clipboard.writeText(terminal.getSelection());
                return false; // Prevent default
              }
            }
            break;
          case "v":
            if (event.type === "keydown") {
              // Paste from clipboard
              navigator.clipboard
                .readText()
                .then((text) => {
                  if (activeSessionId && text) {
                    App.SendInput(activeSessionId, text);
                  }
                })
                .catch((err) => {
                  console.error("Failed to read clipboard:", err);
                });
              return false; // Prevent default
            }
            break;
          case "a":
            if (event.type === "keydown" && terminal) {
              // Select all
              terminal.selectAll();
              return false; // Prevent default
            }
            break;
        }
      }

      // Allow all other standard key events
      return true;
    });

    // Store resize handler for cleanup
    resizeHandler = handleResize;
  }

  function destroyTerminal() {
    terminalReady = false; // Mark terminal as not ready
    terminalInitializing = false; // Reset initializing state

    if (outputPollingInterval) {
      clearInterval(outputPollingInterval);
      outputPollingInterval = null;
    }

    if (terminal) {
      // Clean up resize handler
      if (resizeHandler) {
        window.removeEventListener("resize", resizeHandler);
        resizeHandler = null;
      }

      terminal.dispose();
      terminal = null;
    }

    // Clean up ResizeObserver
    if (resizeObserver) {
      resizeObserver.disconnect();
      resizeObserver = null;
    }

    fitAddon = null;
    currentSessionId = null;
  }

  async function startOutputPolling(sessionId: string) {
    if (outputPollingInterval) {
      clearInterval(outputPollingInterval);
    }

    // Don't poll mock sessions (only exist for testing and browser demo)
    if (sessionId.startsWith("mock_session_")) {
      console.log("Not polling mock session:", sessionId);
      return;
    }

    console.log("Starting output polling for real session:", sessionId);

    let consecutiveErrors = 0;
    let lastOutputTime = Date.now();
    const maxConsecutiveErrors = 5;
    const sessionTimeout = 30000; // 30 seconds without output before health check

    // More frequent polling for better responsiveness
    outputPollingInterval = setInterval(async () => {
      //console.log(`POLLING ATTEMPT - Session: ${sessionId}, Terminal exists: ${!!terminal}`);

      if (!terminal || !sessionId) {
        console.log("POLLING SKIP - No terminal or sessionId");
        return;
      }

      try {
        //console.log(`CALLING GetSessionOutput for session: ${sessionId}`);
        const output = await App.GetSessionOutput(sessionId);
        //console.log(`GetSessionOutput RESPONSE - Session: ${sessionId}, Output length: ${output ? output.length : 'null'}, Output: ${output ? JSON.stringify(output.substring(0, 100)) : 'null'}`);

        if (output && output.length > 0) {
          //console.log(`WRITING TO TERMINAL - Session: ${sessionId}, Length: ${output.length}, Content: ${JSON.stringify(output)}`);
          // Write the output directly to terminal
          terminal.write(output);

          // Auto-scroll to bottom after writing output
          terminal.scrollToBottom();

          // Accumulate raw output for this session
          const currentRaw = sessionRawOutput.get(sessionId) || "";
          sessionRawOutput.set(sessionId, currentRaw + output);

          // Also update the session cache for fallback
          sessionCache.set(sessionId, sessionRawOutput.get(sessionId) || "");

          // Reset error counter and update last output time
          consecutiveErrors = 0;
          lastOutputTime = Date.now();
          //console.log(`OUTPUT WRITTEN SUCCESSFULLY - Session: ${sessionId}`);
        } else {
          //console.log(`NO OUTPUT - Session: ${sessionId}, checking timeout...`);
          // Check if we haven't received output for too long
          const timeSinceLastOutput = Date.now() - lastOutputTime;
          if (timeSinceLastOutput > sessionTimeout) {
            console.log(
              `TIMEOUT DETECTED - Session: ${sessionId}, Time since last output: ${timeSinceLastOutput}ms, checking session health`
            );
            try {
              await App.CheckSessionHealth(sessionId);
              console.log(`HEALTH CHECK PASSED - Session: ${sessionId}`);
              lastOutputTime = Date.now(); // Reset timeout
            } catch (healthError) {
              console.error(
                `HEALTH CHECK FAILED - Session: ${sessionId}:`,
                healthError
              );
              const errorMessage =
                "Session appears unresponsive (health check failed)";
              // Only set error and write warning if not already set for this session
              if (!sessionErrors.has(sessionId)) {
                sessionErrors.set(sessionId, errorMessage);
                sessionErrors = new Map(sessionErrors);
                if (terminal) {
                  terminal.write(
                    `\r\n\x1b[31mWarning: ${errorMessage}\x1b[0m\r\n`
                  );
                }
              }
            }
          }
        }

        // Clear any previous errors for this session
        if (sessionErrors.has(sessionId)) {
          sessionErrors.delete(sessionId);
          sessionErrors = new Map(sessionErrors);
        }
      } catch (error) {
        consecutiveErrors++;
        console.error(
          `POLLING ERROR - Session: ${sessionId}, Error #${consecutiveErrors}/${maxConsecutiveErrors}:`,
          error
        );
        console.error("Error details:", JSON.stringify(error, null, 2));

        if (consecutiveErrors >= maxConsecutiveErrors) {
          const errorMessage = `Too many consecutive errors (${consecutiveErrors}). Session may be dead.`;
          console.error(
            `MAX ERRORS REACHED - Session: ${sessionId}, stopping polling`
          );
          sessionErrors.set(sessionId, errorMessage);
          sessionErrors = new Map(sessionErrors);

          // Display error in terminal
          if (terminal) {
            terminal.write(`\r\n\x1b[31mError: ${errorMessage}\x1b[0m\r\n`);
          }

          // Stop polling to prevent spam
          if (outputPollingInterval) {
            clearInterval(outputPollingInterval);
            outputPollingInterval = null;
          }
        }
      }
    }, 20); // Reduced to 20ms for even better responsiveness with new streaming system
  }

  function setupSession(sessionId: string | null) {
    if (currentSessionId === sessionId) return;

    // Check if there are any other active sessions (besides the current one)
    const sessionsList = get(activeSessions);
    if (!sessionsList || sessionsList.length === 1) {
      // if it's the only session createTerminal is called because it fails to add to DOM otherwise 
      // TODO: remove createTerminal from mount or change this hacky solution
      createTerminal();
    }

    if (outputPollingInterval) {
      clearInterval(outputPollingInterval);
      outputPollingInterval = null;
    }

    // Clear connection error when switching sessions
    connectionError = null;
    isConnecting = false;

    if (!sessionId || !terminal) {
      currentSessionId = null;
      return;
    }

    console.log("Setting up terminal session:", sessionId);
    currentSessionId = sessionId;

    // Check if we have accumulated raw output for this session (like the "welcome" message hosts often send OR previous session output)
    console.log("Checking for accumulated output for session:", sessionId);
    //console.log(
    //  "sessionRawOutput has keys:",
    //  Array.from(sessionRawOutput.keys())
    //);
    const rawOutput = sessionRawOutput.get(sessionId);
    if (rawOutput) {
      console.log(
        "Restoring raw output for session:",
        sessionId,
        rawOutput.length,
        "characters"
      );
      // Fully reset and clear the terminal before restoring output
      terminal.reset();
      terminal.clear();
      terminal.write(rawOutput);
      terminal.scrollToBottom();
    } else {
      console.log(
        "No accumulated output for session:",
        sessionId,
        "- starting fresh"
      );

      // For new sessions, try to get any initial output from the backend
      setTimeout(async () => {
        try {
          const initialOutput = await App.GetSessionOutput(sessionId);
          if (initialOutput && initialOutput.length > 0 && terminal) {
            console.log(
              "Found initial output for new session:",
              sessionId,
              initialOutput.length,
              "characters"
            );
            terminal.write(initialOutput);
            sessionRawOutput.set(sessionId, initialOutput);
            sessionCache.set(sessionId, initialOutput);
            terminal.scrollToBottom();
          }
        } catch (error) {
          console.warn(
            "Failed to get initial output for session:",
            sessionId,
            error
          );
        }
      }, 100);
    }

    // Ensure terminal is properly fitted and send dimensions
    if (fitAddon) {
      fitAddon.fit();
      const dimensions = fitAddon.proposeDimensions();
      if (dimensions) {
        try {
          App.ResizeTerminal(sessionId, dimensions.cols, dimensions.rows);
          console.log(
            "Terminal resized to:",
            dimensions.cols,
            "x",
            dimensions.rows
          );
        } catch (error) {
          console.error("Failed to resize terminal:", error);
        }
      }
    }
    startOutputPolling(sessionId);
  }

  $: activeSession = sessions.find((s) => s.id === activeSessionId);
  $: if (terminal && terminalReady) setupSession(activeSessionId);

  onMount(() => {
    createTerminal();
  });

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect();
      resizeObserver = null;
    }
    fitAddon = null;
    destroyTerminal();
  });
</script>

<div class="flex-1 flex flex-col min-h-0 h-full debug-root-flex">
  <!-- Terminal Content -->
  <div class="flex-1 flex flex-col min-h-0 h-full bg-slate-900 debug-main-bg">
    {#if isConnecting}
      <div
        class="flex-1 min-h-0 h-full flex flex-col items-center justify-center text-slate-400"
      >
        <div class="text-center">
          <Loader2 size={48} class="mx-auto mb-4 animate-spin text-blue-400" />
          <p>Connecting to SSH server...</p>
        </div>
      </div>
    {:else if connectionError}
      <div
        class="flex-1 min-h-0 h-full flex flex-col items-center justify-center text-red-400"
      >
        <div class="text-center max-w-md">
          <AlertCircle size={48} class="mx-auto mb-4" />
          <h3 class="text-lg font-semibold mb-2">Connection Failed</h3>
          <p
            class="text-sm text-slate-300 bg-slate-800 p-4 rounded-lg border border-slate-700"
          >
            {connectionError}
          </p>
        </div>
      </div>
    {:else if activeSession}
      <div class="flex-1 flex flex-col p-4 min-h-0 h-full debug-inner-p4">
        <!-- Session error display -->
        {#if sessionErrors.has(activeSession.id)}
          <div
            class="mb-4 p-3 bg-red-900/20 border border-red-500/50 rounded-lg flex-shrink-0"
          >
            <div class="flex items-center mb-2">
              <AlertCircle size={16} class="text-red-400 mr-2" />
              <span class="text-red-300 text-sm">
                Session Error: {sessionErrors.get(activeSession.id)}
              </span>
            </div>
            <!-- TODO: fix -->
            <button
              class="px-3 py-1 bg-blue-700 hover:bg-blue-800 text-white text-xs rounded transition-colors"
              on:click={() => {
                // Clear error and re-initialize terminal for this session
                sessionErrors.delete(activeSession.id);
                sessionErrors = new Map(sessionErrors);
                // Optionally, clear output and re-poll
                if (terminal) {
                  terminal.reset();
                  terminal.clear();
                }
                // Re-setup session (triggers polling and output restore)
                setupSession(activeSession.id);
              }}
            >
              Refresh Terminal
            </button>
          </div>
        {/if}

        <!-- xterm.js terminal -->
        <div
          class="terminal-container flex-1 flex flex-col min-h-0 h-full rounded-lg border border-slate-700 bg-slate-900 overflow-hidden relative debug-terminal-container"
          bind:this={terminalElement}
        >
          {#if !terminalElement && activeSession}
            {() => {
              console.warn(
                "Fallback: terminalElement not present but session is active, calling createTerminal()"
              );
              createTerminal();
              return null;
            }}
          {/if}
          <!-- Terminal initialization loading overlay, usually not visible -->
          {#if terminalInitializing || (!terminal && !terminalReady)}
            <div
              class="absolute inset-0 flex items-center justify-center bg-slate-900/80 z-10"
            >
              <div class="text-center">
                <Loader2
                  size={32}
                  class="animate-spin text-blue-400 mx-auto mb-2"
                />
                <div class="text-sm text-slate-400">
                  Initializing terminal...
                </div>
              </div>
            </div>
          {/if}
          <!-- Terminal will be rendered here -->
        </div>
      </div>
    {:else}
      <div
        class="flex-1 min-h-0 h-full flex flex-col items-center justify-center text-slate-400"
      >
        <div class="text-center">
          <TerminalIcon size={48} class="mx-auto mb-4 opacity-50" />
          <p>Select a host from the sidebar to start a session</p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Custom scrollbar for terminal */
  .terminal-container :global(.xterm-viewport) {
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;
    transition: scrollbar-color 0.3s ease;
  }

  /* Webkit scrollbar styling */
  .terminal-container :global(.xterm-viewport)::-webkit-scrollbar {
    width: 12px;
  }

  .terminal-container :global(.xterm-viewport)::-webkit-scrollbar-track {
    background: transparent;
    border-radius: 6px;
  }

  .terminal-container :global(.xterm-viewport)::-webkit-scrollbar-thumb {
    background: transparent;
    border-radius: 6px;
    border: 2px solid transparent;
    background-clip: content-box;
    transition: background 0.3s ease;
  }

  /* Show blue scrollbar thumb when hovering the scrollbar area */
  :global(.xterm-viewport)::-webkit-scrollbar-thumb:hover {
    background: rgba(96, 165, 250, 0.8);
    background-clip: content-box;
  }

  .terminal-container :global(.xterm-viewport)::-webkit-scrollbar-thumb:active {
    background: rgba(96, 165, 250, 1);
    background-clip: content-box;
  }

  /* Show blue scrollbar thumb while scrolling */

  /* Smooth transitions */
  .terminal-container :global(.xterm-viewport) {
    transition: all 0.2s ease-in-out;
  }

  /* Force scrollbar to always be present but invisible by default */
  .terminal-container :global(.xterm-viewport)::-webkit-scrollbar {
    background: transparent;
  }
</style>

import { writable, derived, type Writable } from 'svelte/store';
import type { Host, Macro, Session, HistoryEntry } from './api';

// Session cache for storing terminal output
class SessionCache {
  private cache = new Map<string, string>();
  
  set(sessionId: string, content: string) {
    this.cache.set(sessionId, content);
  }
  
  get(sessionId: string): string | undefined {
    return this.cache.get(sessionId);
  }
  
  has(sessionId: string): boolean {
    return this.cache.has(sessionId);
  }
  
  delete(sessionId: string) {
    this.cache.delete(sessionId);
  }
  
  clear() {
    this.cache.clear();
  }
}

// Session SFTP layout cache
class SessionSFTPLayouts {
  private layouts = new Map<string, 'bottom' | 'top' | 'fullscreen' | 'hidden'>();
  
  set(sessionId: string, layout: 'bottom' | 'top' | 'fullscreen' | 'hidden') {
    this.layouts.set(sessionId, layout);
  }
  
  get(sessionId: string): 'bottom' | 'top' | 'fullscreen' | 'hidden' {
    return this.layouts.get(sessionId) || 'hidden';
  }
  
  delete(sessionId: string) {
    this.layouts.delete(sessionId);
  }
  
  clear() {
    this.layouts.clear();
  }
}

export const sessionCache = new SessionCache();
export const sessionSFTPLayouts = new SessionSFTPLayouts();

// Application state stores
export const hosts: Writable<Host[]> = writable([]);
export const macros: Writable<Macro[]> = writable([]);
export const activeSessions: Writable<Session[]> = writable([]);
export const activeTab: Writable<string | null> = writable(null);
export const showSFTP: Writable<boolean> = writable(false);
// Remove global sftpLayout - now handled per session
export const history: Writable<HistoryEntry[]> = writable([]);

// UI state stores
export const sidebarCollapsed: Writable<boolean> = writable(false);
export const rightSidebarTab: Writable<'macros' | 'history'> = writable('macros');
export const showHostModal: Writable<boolean> = writable(false);
export const showMacroModal: Writable<boolean> = writable(false);
export const isDarkMode: Writable<boolean> = writable(true);

// Terminal theme interface
export interface TerminalTheme {
  name: string;
  background: string;
  foreground: string;
  cursor: string;
  cursorAccent: string;
  black: string;
  red: string;
  green: string;
  yellow: string;
  blue: string;
  magenta: string;
  cyan: string;
  white: string;
  brightBlack: string;
  brightRed: string;
  brightGreen: string;
  brightYellow: string;
  brightBlue: string;
  brightMagenta: string;
  brightCyan: string;
  brightWhite: string;
}

// Default terminal theme (current hardcoded theme)
export const defaultTerminalTheme: TerminalTheme = {
  name: 'Default Dark',
  background: '#0f172a',
  foreground: '#f1f5f9',
  cursor: '#60a5fa',
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
};

// Terminal theme store
export const terminalTheme: Writable<TerminalTheme> = writable(defaultTerminalTheme);

// Derived stores
export const activeSession = derived(
  [activeSessions, activeTab],
  ([$activeSessions, $activeTab]) => {
    if (!$activeTab) return null;
    return $activeSessions.find(session => session.id === $activeTab) || null;
  }
);

export const hostsGroupedByTags = derived(hosts, ($hosts) => {
  const grouped: Record<string, Host[]> = {};
  
  $hosts.forEach(host => {
    if (host.tags.length === 0) {
      if (!grouped['Untagged']) grouped['Untagged'] = [];
      grouped['Untagged'].push(host);
    } else {
      host.tags.forEach(tag => {
        if (!grouped[tag]) grouped[tag] = [];
        grouped[tag].push(host);
      });
    }
  });
  
  return grouped;
});

export const recentHosts = derived(hosts, ($hosts) => {
  return $hosts
    .filter(host => host.last_used)
    .sort((a, b) => {
      const aTime = a.last_used ? new Date(a.last_used).getTime() : 0;
      const bTime = b.last_used ? new Date(b.last_used).getTime() : 0;
      return bTime - aTime;
    })
    .slice(0, 5);
});

export const macrosForActiveHost = derived(
  [macros, activeSession],
  ([$macros, $activeSession]) => {
    if (!$activeSession) return $macros;
    
    return $macros.filter(macro => 
      macro.host_ids.length === 0 || 
      macro.host_ids.indexOf($activeSession.hostId) !== -1
    );
  }
);

// Terminal state
export const terminalOutput: Writable<Record<string, string[]>> = writable({});
export const terminalInput: Writable<string> = writable('');

// SFTP state
export const localDirectory: Writable<string> = writable('');
export const remoteDirectory: Writable<string> = writable('');
export const localFiles: Writable<any[]> = writable([]);
export const remoteFiles: Writable<any[]> = writable([]);

// Notification system
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  duration?: number;
}

export const notifications: Writable<Notification[]> = writable([]);

// Helper functions for stores
export const addNotification = (notification: Omit<Notification, 'id'>) => {
  const id = Date.now().toString();
  const newNotification: Notification = {
    ...notification,
    id,
    duration: notification.duration || 5000,
  };
  
  notifications.update(notifs => [...notifs, newNotification]);
  
  // Auto-remove after duration
  if (newNotification.duration && newNotification.duration > 0) {
    setTimeout(() => {
      notifications.update(notifs => notifs.filter(n => n.id !== id));
    }, newNotification.duration);
  }
};

export const removeNotification = (id: string) => {
  notifications.update(notifs => notifs.filter(n => n.id !== id));
};

export const addTerminalOutput = (sessionId: string, output: string) => {
  terminalOutput.update(outputs => ({
    ...outputs,
    [sessionId]: [...(outputs[sessionId] || []), output]
  }));
};

export const clearTerminalOutput = (sessionId: string) => {
  terminalOutput.update(outputs => ({
    ...outputs,
    [sessionId]: []
  }));
};

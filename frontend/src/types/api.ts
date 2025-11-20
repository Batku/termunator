// Re-export types from generated Wails models
import { models } from '../../wailsjs/go/models';

// Use the generated types
export type Host = models.Host;
export type HostCreateRequest = models.HostCreateRequest;
export type Macro = models.Macro;
export type MacroCreateRequest = models.MacroCreateRequest;
export type SFTPFileInfo = models.SFTPFileInfo;
export type HistoryEntry = models.HistoryEntry;

// Session type based on the Go SSHSession
export interface Session {
  id: string;
  hostId: string;
  host?: Host;
  isActive: boolean;
  createdAt: string;
  displayName?: string; // For numbered sessions like "host (2)"
  cachedOutput?: string; // Cached terminal output
}

// Auth method type for convenience
export type AuthMethod = 'password' | 'private_key' | 'ssh_agent';

// SFTP System Types
export interface FileItem {
  name: string;
  path: string;
  size: number;
  modTime: string;
  isDir: boolean;
  permissions: string;
  owner?: string;
  group?: string;
}

export interface TransferItem {
  id: string;
  sourceFile: FileItem;
  destinationPath: string;
  direction: 'upload' | 'download';
  status: 'pending' | 'in-progress' | 'completed' | 'failed' | 'paused' | 'cancelled';
  progress: number; // 0-100
  bytesTransferred: number;
  totalBytes: number;
  speed: number; // bytes per second
  errorMessage?: string;
  startTime: Date;
  endTime?: Date;
}

export interface TransferQueue {
  items: TransferItem[];
  activeTransfers: number;
  maxConcurrency: number;
  totalItems: number;
  completedItems: number;
  failedItems: number;
}

export interface ConflictResolution {
  action: 'replace' | 'keep-both' | 'cancel';
  applyToAll: boolean;
}

export interface SFTPConnection {
  sessionId: string;
  isConnected: boolean;
  remotePath: string;
  localPath: string;
  lastError?: string;
}

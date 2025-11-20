import type { 
  Host, 
  HostCreateRequest, 
  Macro, 
  MacroCreateRequest, 
  Session, 
  SFTPFileInfo
} from '../types/api';

// Import Wails runtime
import * as App from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';

// Extend Window interface to include Wails runtime
declare global {
  interface Window {
    go?: {
      main?: {
        App?: any;
      };
    };
  }
}

// Function to wait for Wails runtime to be ready (for new WebView2Loader)
async function waitForWailsRuntime(maxWaitMs: number = 5000): Promise<boolean> {
  console.log('=== WAITING FOR WAILS RUNTIME ===');
  const startTime = Date.now();
  let attempts = 0;
  
  while (Date.now() - startTime < maxWaitMs) {
    attempts++;
    console.log(`Runtime check attempt ${attempts}:`, {
      hasWindow: typeof window !== 'undefined',
      hasGo: typeof window !== 'undefined' && !!window.go,
      hasMain: typeof window !== 'undefined' && !!window.go?.main,
      hasApp: typeof window !== 'undefined' && !!window.go?.main?.App,
      hasCreateHost: typeof window !== 'undefined' && typeof window.go?.main?.App?.CreateHost === 'function',
      windowKeys: typeof window !== 'undefined' ? Object.keys(window) : 'no window',
      goKeys: typeof window !== 'undefined' && window.go ? Object.keys(window.go) : 'no go'
    });
    
    if (typeof window !== 'undefined' && 
        window.go && 
        window.go.main && 
        window.go.main.App &&
        typeof window.go.main.App.CreateHost === 'function') {
      console.log('=== WAILS RUNTIME IS READY! ===');
      return true;
    }
    
    // Wait 50ms before checking again
    await new Promise(resolve => setTimeout(resolve, 50));
  }
  
  console.log('=== WAILS RUNTIME FAILED TO LOAD WITHIN TIMEOUT ===');
  return false;
}

// Enhanced environment detection with async runtime check
async function checkWailsEnvironment(): Promise<boolean> {
  console.log('=== CHECKING WAILS ENVIRONMENT ===');
  
  // First do immediate check
  const immediateCheck = typeof window !== 'undefined' && 
    window.go && 
    window.go.main && 
    window.go.main.App &&
    typeof window.go.main.App.CreateHost === 'function';
    
  console.log('Immediate environment check:', {
    result: immediateCheck,
    hasWindow: typeof window !== 'undefined',
    hasGo: typeof window !== 'undefined' && !!window.go,
    hasMain: typeof window !== 'undefined' && !!window.go?.main,
    hasApp: typeof window !== 'undefined' && !!window.go?.main?.App,
    hasCreateHost: typeof window !== 'undefined' && typeof window.go?.main?.App?.CreateHost === 'function'
  });
    
  if (immediateCheck) {
    console.log('=== WAILS RUNTIME IMMEDIATELY AVAILABLE ===');
    return true;
  }
  
  // If not immediately available, wait for it (new WebView2Loader behavior)
  console.log('=== WAILS RUNTIME NOT IMMEDIATELY AVAILABLE, WAITING... ===');
  return await waitForWailsRuntime();
}

// Global variable to cache environment check
let wailsEnvironmentChecked = false;
let isWailsEnvironment = false;

// Initialize environment check
async function initializeEnvironment() {
  
  if (wailsEnvironmentChecked) {
    //console.log('Environment already checked, result:', isWailsEnvironment);
    return isWailsEnvironment;
  }
  
  console.log('Running first-time environment check...');
  isWailsEnvironment = await checkWailsEnvironment();
  wailsEnvironmentChecked = true;
  
  console.log('=== ENVIRONMENT CHECK COMPLETE ===', {
    hasWindow: typeof window !== 'undefined',
    hasWindowGo: typeof window !== 'undefined' && !!window.go,
    hasWindowGoMain: typeof window !== 'undefined' && !!window.go?.main,
    hasWindowGoMainApp: typeof window !== 'undefined' && !!window.go?.main?.App,
    hasCreateHostMethod: typeof window !== 'undefined' && typeof window.go?.main?.App?.CreateHost === 'function',
    hasApp: typeof App !== 'undefined',
    hasCreateHost: typeof App.CreateHost === 'function',
    finalResult: isWailsEnvironment,
    willUseMockData: !isWailsEnvironment
  });
  
  return isWailsEnvironment;
}

// Mock data for web browser development
const mockHosts: Host[] = [
  models.Host.createFrom({
    id: '1',
    label: 'Development Server',
    hostname: '192.168.1.100',
    port: 22,
    username: 'dev',
    auth_method: 'password',
    password: '',
    private_key: '',
    tags: ['development', 'testing'],
    last_used: null,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }),
  models.Host.createFrom({
    id: '2', 
    label: 'Production Server',
    hostname: 'prod.example.com',
    port: 22,
    username: 'admin',
    auth_method: 'password',
    password: '',
    private_key: '',
    tags: ['production'],
    last_used: null,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  })
];

const mockMacros: Macro[] = [
  models.Macro.createFrom({
    id: '1',
    label: 'System Update',
    commands: ['sudo apt update', 'sudo apt upgrade -y'],
    host_ids: ['1', '2'],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  })
];

// Host Management API
export class HostAPI {
  static async create(request: HostCreateRequest): Promise<Host> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      const newHost = models.Host.createFrom({
        ...request,
        id: Date.now().toString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      });
      mockHosts.push(newHost);
      return newHost;
    }

    try {
      return await App.CreateHost(request);
    } catch (error) {
      console.error('Failed to create host:', error);
      throw error;
    }
  }

  static async getAll(): Promise<Host[]> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Return mock data for browser
      console.log('Using mock hosts data (not Wails environment)');
      return [...mockHosts];
    }

    try {
      console.log('Attempting to get hosts from Wails backend...');
      const backendHosts = await App.GetHosts();
      console.log('Backend hosts received:', backendHosts);
      
      // If backend returns empty, provide mock data for development
      if (!backendHosts || backendHosts.length === 0) {
        console.log('Backend returned no hosts, using mock data for development');
        return [...mockHosts];
      }
      
      return backendHosts;
    } catch (error) {
      console.error('Failed to get hosts from backend:', error);
      // Fallback to mock data if backend fails
      console.log('Falling back to mock hosts data due to backend error');
      return [...mockHosts];
    }
  }

  static async update(hostId: string, request: HostCreateRequest): Promise<Host> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      const index = mockHosts.findIndex(h => h.id === hostId);
      if (index > -1) {
        const updatedHost = models.Host.createFrom({
          ...request,
          id: hostId,
          created_at: mockHosts[index].created_at,
          updated_at: new Date().toISOString()
        });
        mockHosts[index] = updatedHost;
        return updatedHost;
      }
      throw new Error('Host not found');
    }

    try {
      return await App.UpdateHost(hostId, request);
    } catch (error) {
      console.error('Failed to update host:', error);
      throw error;
    }
  }

  static async delete(hostId: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      const index = mockHosts.findIndex(h => h.id === hostId);
      if (index > -1) {
        mockHosts.splice(index, 1);
      }
      return;
    }

    try {
      await App.DeleteHost(hostId);
    } catch (error) {
      console.error('Failed to delete host:', error);
      throw error;
    }
  }
}

// SSH Session API
export class SessionAPI {
  static async connect(hostId: string): Promise<string> {
    return this.connectWithDimensions(hostId, 80, 24);
  }

  static async connectWithDimensions(hostId: string, cols: number, rows: number): Promise<string> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Creating session for host', hostId, 'with dimensions', cols, 'x', rows);
      return `session_${Date.now()}`;
    }

    try {
      console.log('Attempting to connect to host via backend:', hostId, 'with dimensions', cols, 'x', rows);
      const sessionId = await App.ConnectSSHWithDimensions(hostId, cols, rows);
      console.log('Backend connection successful, session ID:', sessionId);
      return sessionId;
    } catch (error) {
      console.error('Failed to connect to host via backend:', error);
      // Re-throw the error instead of creating a mock session
      throw error;
    }
  }

  static async getActiveSessions(): Promise<Session[]> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      return [];
    }

    try {
      const sessions = await App.GetActiveSessions();
      // Convert the Go map to an array
      return Object.entries(sessions).map(([id, session]) => ({
        id,
        hostId: session.HostID || '',
        host: undefined, // Will be populated by the store
        isActive: session.IsConnected || false,
        createdAt: new Date().toISOString()
      }));
    } catch (error) {
      console.error('Failed to get active sessions:', error);
      return [];
    }
  }

  static async closeSession(sessionId: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Closing session', sessionId);
      return;
    }

    try {
      await App.CloseSSHSession(sessionId);
    } catch (error) {
      console.error('Failed to close session:', error);
      throw error;
    }
  }

  static async getPing(sessionId: string): Promise<number | null> {
    const isWails = await initializeEnvironment();
    if (!isWails) return null;
    try {
      return await App.GetSessionPing(sessionId);
    } catch (error) {
      console.error('Failed to get session ping:', error);
      return null;
    }
  }
}

// Macro API
export class MacroAPI {
  static async create(request: MacroCreateRequest): Promise<Macro> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      const newMacro = models.Macro.createFrom({
        ...request,
        id: Date.now().toString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      });
      mockMacros.push(newMacro);
      return newMacro;
    }

    try {
      return await App.CreateMacro(request);
    } catch (error) {
      console.error('Failed to create macro:', error);
      throw error;
    }
  }

  static async getAll(): Promise<Macro[]> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Using mock macros data (not Wails environment)');
      return [...mockMacros];
    }

    try {
      console.log('Attempting to get macros from Wails backend...');
      const backendMacros = await App.GetMacros();
      console.log('Backend macros received:', backendMacros);
      
      // If backend returns empty, provide mock data for development
      if (!backendMacros || backendMacros.length === 0) {
        console.log('Backend returned no macros, using mock data for development');
        return [...mockMacros];
      }
      
      return backendMacros;
    } catch (error) {
      console.error('Failed to get macros from backend:', error);
      return [...mockMacros];
    }
  }

  static async execute(sessionId: string, macroId: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Executing macro', macroId, 'on session', sessionId);
      return;
    }

    try {
      await App.ExecuteMacro(sessionId, macroId);
    } catch (error) {
      console.error('Failed to execute macro:', error);
      throw error;
    }
  }
}

// SFTP API
export class SFTPAPI {
  /**
   * Reads a local file as a Blob using the Go backend. Returns a Blob for upload logic.
   * @param path Local file path
   */
  static async readLocalFileAsBlob(path: string): Promise<Blob> {
    const isWails = await initializeEnvironment();
    if (!isWails) {
      // Mock: return an empty Blob
      console.log('Mock: readLocalFileAsBlob for', path);
      return new Blob();
    }
    try {
      // Call Go backend to get base64-encoded file contents
      const base64Data = await App.ReadLocalFileAsBytes(path);
      // Decode base64 to binary
      const binaryString = atob(base64Data);
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      return new Blob([bytes]);
    } catch (error) {
      console.error('Failed to read local file as blob:', error);
      throw error;
    }
  }
  static async connect(hostId: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Connecting SFTP to host', hostId);
      return;
    }

    try {
      await App.ConnectSFTP(hostId);
    } catch (error) {
      console.error('Failed to connect SFTP:', error);
      throw error;
    }
  }

  static async listDirectory(hostId: string, path: string): Promise<SFTPFileInfo[] | null> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Listing remote directory', path, 'for host', hostId);
      return [
        models.SFTPFileInfo.createFrom({
          name: 'Documents',
          path: '/home/user/Documents',
          size: 0,
          mode: 'drwxr-xr-x',
          mod_time: new Date(),
          is_dir: true,
          permissions: '755'
        }),
        models.SFTPFileInfo.createFrom({
          name: 'file.txt',
          path: '/home/user/file.txt',
          size: 1024,
          mode: '-rw-r--r--',
          mod_time: new Date(),
          is_dir: false,
          permissions: '644'
        })
      ];
    }

    try {
      return await App.ListDirectory(hostId, path);
    } catch (error) {
      console.error('listDirectory error object:', error);
      // Handle both string and object errors safely
      let msg = '';
      if (typeof error === 'string') {
        msg = error;
      } else if (typeof error === 'object' && error !== null && 'message' in error) {
        msg = (error as any).message;
      }
      if (msg && String(msg).includes('file does not exist')) {
        return null;
      }
      console.error('Failed to list directory:', error);
      throw error;
    }
  }

  static async listLocalDirectory(path: string): Promise<SFTPFileInfo[]> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Listing local directory', path);
      return [
        models.SFTPFileInfo.createFrom({
          name: 'Desktop',
          path: 'C:\\Users\\User\\Desktop',
          size: 0,
          mode: 'drwxr-xr-x',
          mod_time: new Date(),
          is_dir: true,
          permissions: '755'
        }),
        models.SFTPFileInfo.createFrom({
          name: 'test.txt',
          path: 'C:\\Users\\User\\test.txt',
          size: 512,
          mode: '-rw-r--r--',
          mod_time: new Date(),
          is_dir: false,
          permissions: '644'
        })
      ];
    }

    try {
      return await App.ListLocalDirectory(path);
    } catch (error) {
      console.error('Failed to list local directory:', error);
      return [];
    }
  }

  static async getClientHome(): Promise<string> {
    const isWails = await initializeEnvironment();
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Getting client home directory');
      return '/home/user';
    }
    try {
      return await App.GetClientHome();
    } catch (error) {
      console.error('Failed to get client home directory:', error);
      return '/';
    }
  }
  

  static async downloadFile(hostId: string, remotePath: string, localPath: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Downloading file', remotePath, 'to', localPath);
      return;
    }

    try {
      await App.DownloadFile(hostId, remotePath, localPath);
    } catch (error) {
      console.error('Failed to download file:', error);
      throw error;
    }
  }

  static async uploadFile(hostId: string, localPath: string, remotePath: string): Promise<void> {
    const isWails = await initializeEnvironment();
    
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Uploading file', localPath, 'to', remotePath);
      return;
    }

    try {
      await App.UploadFile(hostId, localPath, remotePath);
    } catch (error) {
      console.error('Failed to upload file:', error);
      throw error;
    }
  }

  /**
   * Upload a file/blob to the remote server at the given remotePath.
   * In browser/mock mode, just logs. In Wails, must be implemented in backend.
   */
  static async uploadFileFromBlob(hostId: string, file: File | Blob, remotePath: string): Promise<void> {
    const isWails = await initializeEnvironment();
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Uploading file/blob to', remotePath, 'for host', hostId, file);
      return;
    }
    // Convert File/Blob to base64 string
    const arrayBuffer = await file.arrayBuffer();
    const uint8Array = new Uint8Array(arrayBuffer);
    let binary = '';
    for (let i = 0; i < uint8Array.length; i++) {
      binary += String.fromCharCode(uint8Array[i]);
    }
    const base64Data = btoa(binary);
    // Call Go backend method
    return await App.UploadFileFromBytes(hostId, remotePath, base64Data);
  }

  static async makeDirectory(hostId: string, dirPath: string): Promise<void> {
    const isWails = await initializeEnvironment();
    if (!isWails) {
      // Mock implementation for browser
      console.log('Mock: Creating remote directory', dirPath, 'for host', hostId);
      return;
    }
    try {
      await App.MakeDirectory(hostId, dirPath);
    } catch (error) {
      console.error('Failed to create remote directory:', error);
      throw error;
    }
  }
}

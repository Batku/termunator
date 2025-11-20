<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import {
    FolderOpen,
    File as FileIcon,
    Upload,
    RefreshCw,
    Home,
    ChevronRight,
    ChevronLeft,
    Search,
    MoreVertical,
    X,
    Play,
    Pause,
    Settings,
    Maximize2,
    Square,
  } from "lucide-svelte";

  import type { Session, TransferQueue, SFTPFileInfo } from "../types/api";
  import { SFTPAPI } from "../lib/api";
  import { addNotification } from "../types/stores";
  import ConflictDialog from "./ConflictDialog.svelte";
  import { tick } from "svelte";

  export let activeSession: Session | undefined = undefined;
  export let layout: "bottom" | "top" | "fullscreen" | "hidden" = "hidden";

  const dispatch = createEventDispatcher<{
    layoutChange: "bottom" | "top" | "fullscreen" | "hidden";
    close: void;
  }>();

  // State management - per session state caching
  let sessionFileState = new Map<
    string,
    {
      localFiles: SFTPFileInfo[];
      remoteFiles: SFTPFileInfo[];
      localPath: string;
      remotePath: string;
      selectedLocalFiles: Set<string>;
      selectedRemoteFiles: Set<string>;
    }
  >();

  let localFiles: SFTPFileInfo[] = [];
  let remoteFiles: SFTPFileInfo[] = [];
  let localPath = ""; // Will be set to user home directory
  let remotePath = "/";
  let localSearchQuery = "";
  let remoteSearchQuery = "";
  let selectedLocalFiles: Set<string> = new Set();
  let selectedRemoteFiles: Set<string> = new Set();
  let transferQueue: TransferQueue = {
    items: [],
    activeTransfers: 0,
    maxConcurrency: 3,
    totalItems: 0,
    completedItems: 0,
    failedItems: 0,
  };
  let showTransferQueue = false;
  // let showConflictDialog = false;
  // let conflictResolution: ConflictResolution = { action: 'replace', applyToAll: false };
  let dragOverLocal = false;
  let dragOverRemote = false;
  let loadingLocal = false;
  let loadingRemote = false;
  let previousSessionId: string | undefined = undefined;
  let previousLayout: "bottom" | "top" | "fullscreen" | "hidden" | undefined =
    undefined;

  // Conflict dialog state
  let showConflictDialog = false;
  let conflictList: Array<{ name: string; type: "file" | "folder" }> = [];
  let conflictDialogPromise: ((result: any) => void) | null = null;
  let currentConflictIndex: number = 0;

  // Helper to show the conflict dialog and await result
  //function showConflictDialogAsync(conflicts: Array<{ name: string; type: 'file' | 'folder' }>) {
  //  showConflictDialog = true;
  //  conflictList = conflicts;
  //  return new Promise((resolve) => {
  //    conflictDialogPromise = resolve;
  //  });
  //}

  function handleConflictDialogResolve(result: {
    action: string;
    applyToAll: boolean;
  }) {
    showConflictDialog = false;
    if (conflictDialogPromise) conflictDialogPromise(result);
    conflictDialogPromise = null;
  }
  // Layout management
  function changeLayout(newLayout: typeof layout) {
    console.log(
      "SFTPPanel: changeLayout called - from",
      layout,
      "to",
      newLayout
    );
    console.log("Stack trace:", new Error().stack);
    layout = newLayout;
    dispatch("layoutChange", newLayout);
  }

  function closePanel() {
    console.log("SFTPPanel: closePanel called");
    console.log("Stack trace:", new Error().stack);
    layout = "hidden";
    dispatch("close");
  }

  // Session state management
  function saveSessionState(sessionId: string) {
    if (!sessionId) return;
    sessionFileState.set(sessionId, {
      localFiles: [...localFiles],
      remoteFiles: [...remoteFiles],
      localPath,
      remotePath,
      selectedLocalFiles: new Set(selectedLocalFiles),
      selectedRemoteFiles: new Set(selectedRemoteFiles),
    });
  }

  function restoreSessionState(sessionId: string) {
    const state = sessionFileState.get(sessionId);
    if (state) {
      localFiles = state.localFiles;
      remoteFiles = state.remoteFiles;
      localPath = state.localPath;
      remotePath = state.remotePath;
      selectedLocalFiles = state.selectedLocalFiles;
      selectedRemoteFiles = state.selectedRemoteFiles;
      console.log("Restored SFTP state for session:", sessionId);
      return true;
    }
    return false;
  }

  // File operations
  async function loadLocalFiles() {
    loadingLocal = true;
    try {
      console.log("Loading local files from:", localPath);
      localFiles = await SFTPAPI.listLocalDirectory(localPath);
      console.log("Local files loaded:", localFiles.length, "files");
    } catch (error) {
      console.error("Failed to load local files:", error);
      localFiles = [];
    } finally {
      loadingLocal = false;
    }
  }

  async function initializeLocalPath() {
    try {
      localPath = await SFTPAPI.getClientHome();
      console.log("User home directory:", localPath);
      await loadLocalFiles();
    } catch (error) {
      console.error("Failed to get user home directory:", error);
      const isWindows = navigator.platform.includes("Win");
      localPath = isWindows ? "C:\\" : "/";
    }
  }

  async function loadRemoteFiles() {
    console.log("loadRemoteFiles called, activeSession:", activeSession);
    console.log("activeSession?.hostId:", activeSession?.hostId);
    console.log("activeSession?.id:", activeSession?.id);
    console.log("activeSession?.host:", activeSession?.host);

    if (!activeSession?.hostId) {
      console.log("No active session or hostId, skipping remote file load");
      return;
    }
    loadingRemote = true;
    try {
      console.log(
        "Loading remote files from:",
        remotePath,
        "for host:",
        activeSession.hostId
      );
      // First make sure we have an SFTP connection
      await SFTPAPI.connect(activeSession.hostId);
      console.log("SFTP connection established");
      remoteFiles =
        (await SFTPAPI.listDirectory(activeSession.hostId, remotePath)) ?? [];
      console.log("Remote files loaded:", remoteFiles.length, "files");
    } catch (error) {
      console.error("Failed to load remote files:", error);
      remoteFiles = [];
    } finally {
      loadingRemote = false;
    }
  }

  // Navigation functions
  function navigateLocal(path: string) {
    localPath = path;
    loadLocalFiles();
  }

  function navigateRemote(path: string) {
    remotePath = path;
    loadRemoteFiles();
  }

  function goUpLocal() {
    const isWindows = localPath.includes("\\");
    const separator = isWindows ? "\\" : "/";
    const parts = localPath.split(separator).filter((part) => part.length > 0);

    if (parts.length > 1) {
      parts.pop();
      localPath = isWindows
        ? parts.join(separator) + separator
        : "/" + parts.join("/");
    } else if (isWindows && parts.length === 1) {
      // On Windows, go to root drives list (C:\, D:\, etc.)
      localPath = "";
    } else {
      localPath = "/";
    }
    loadLocalFiles();
  }

  function goUpRemote() {
    const parts = remotePath.split("/").filter((part) => part.length > 0);
    if (parts.length > 0) {
      parts.pop();
      remotePath = "/" + parts.join("/");
      if (remotePath === "/") remotePath = "/";
    } else {
      remotePath = "/";
    }
    loadRemoteFiles();
  }

  function navigateToLocalPathSegment(index: number) {
    const isWindows = localPath.includes("\\");
    const separator = isWindows ? "\\" : "/";
    const parts = localPath.split(separator).filter((part) => part.length > 0);

    if (index === -1) {
      // Navigate to root
      localPath = isWindows ? "" : "/";
    } else {
      const newParts = parts.slice(0, index + 1);
      localPath = isWindows
        ? newParts.join(separator) + separator
        : "/" + newParts.join("/");
    }
    loadLocalFiles();
  }

  function navigateToRemotePathSegment(index: number) {
    const parts = remotePath.split("/").filter((part) => part.length > 0);

    if (index === -1) {
      // Navigate to root
      remotePath = "/";
    } else {
      const newParts = parts.slice(0, index + 1);
      remotePath = "/" + newParts.join("/");
    }
    loadRemoteFiles();
  }

  // File selection
  function toggleLocalSelection(fileName: string, event?: MouseEvent) {
    if (event && event.ctrlKey) {
      // Multi-select with Ctrl
      if (selectedLocalFiles.has(fileName)) {
        selectedLocalFiles.delete(fileName);
      } else {
        selectedLocalFiles.add(fileName);
      }
    } else {
      // Single select
      if (selectedLocalFiles.size === 1 && selectedLocalFiles.has(fileName)) {
        selectedLocalFiles.clear();
      } else {
        selectedLocalFiles.clear();
        selectedLocalFiles.add(fileName);
      }
    }
    selectedLocalFiles = new Set(selectedLocalFiles);
  }

  function handleLocalDragStart(event: DragEvent, fileName: string) {
    // If not already selected, select this file only
    if (!selectedLocalFiles.has(fileName)) {
      selectedLocalFiles.clear();
      selectedLocalFiles.add(fileName);
      selectedLocalFiles = new Set(selectedLocalFiles);
    }
    // Set drag data as JSON array of selected file names
    event.dataTransfer?.setData(
      "application/json",
      JSON.stringify(Array.from(selectedLocalFiles))
    );
    event.dataTransfer?.setData("text/plain", "local-to-remote");

    // Custom drag image for multi-select
    if (event.dataTransfer) {
      let dragImage: HTMLElement;
      if (selectedLocalFiles.size > 1) {
        dragImage = document.createElement("div");
        dragImage.style.cssText =
          "padding: 8px 16px; background: #1e293b; color: #fff; border-radius: 6px; border: 2px solid #3b82f6; font-size: 14px; font-family: inherit; display: flex; align-items: center; gap: 8px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        dragImage.innerHTML =
          `<svg width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="#3b82f6" style="margin-right:6px;"><rect x="3" y="5" width="18" height="14" rx="2" stroke-width="2"/></svg>` +
          `<span>${selectedLocalFiles.size} items</span>`;
      } else {
        dragImage = document.createElement("div");
        dragImage.style.cssText =
          "padding: 8px 16px; background: #1e293b; color: #fff; border-radius: 6px; border: 2px solid #3b82f6; font-size: 14px; font-family: inherit; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        dragImage.textContent = fileName;
      }
      document.body.appendChild(dragImage);
      event.dataTransfer.setDragImage(
        dragImage,
        dragImage.offsetWidth / 2,
        dragImage.offsetHeight / 2
      );
      // Remove drag image after a short delay
      setTimeout(() => document.body.removeChild(dragImage), 0);
    }
  }

  function handleRemoteDragStart(event: DragEvent, fileName: string) {
    if (!selectedRemoteFiles.has(fileName)) {
      selectedRemoteFiles.clear();
      selectedRemoteFiles.add(fileName);
      selectedRemoteFiles = new Set(selectedRemoteFiles);
    }
    event.dataTransfer?.setData(
      "application/json",
      JSON.stringify(Array.from(selectedRemoteFiles))
    );
    event.dataTransfer?.setData("text/plain", "remote-to-local");

    // Custom drag image for multi-select
    if (event.dataTransfer) {
      let dragImage: HTMLElement;
      if (selectedRemoteFiles.size > 1) {
        dragImage = document.createElement("div");
        dragImage.style.cssText =
          "padding: 8px 16px; background: #1e293b; color: #fff; border-radius: 6px; border: 2px solid #22c55e; font-size: 14px; font-family: inherit; display: flex; align-items: center; gap: 8px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        dragImage.innerHTML =
          `<svg width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="#22c55e" style="margin-right:6px;"><rect x="3" y="5" width="18" height="14" rx="2" stroke-width="2"/></svg>` +
          `<span>${selectedRemoteFiles.size} items</span>`;
      } else {
        dragImage = document.createElement("div");
        dragImage.style.cssText =
          "padding: 8px 16px; background: #1e293b; color: #fff; border-radius: 6px; border: 2px solid #22c55e; font-size: 14px; font-family: inherit; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;";
        dragImage.textContent = fileName;
      }
      document.body.appendChild(dragImage);
      event.dataTransfer.setDragImage(
        dragImage,
        dragImage.offsetWidth / 2,
        dragImage.offsetHeight / 2
      );
      setTimeout(() => document.body.removeChild(dragImage), 0);
    }
  }

  // Helper: Ensure remote directory exists (recursively create if needed)
  async function ensureRemoteDirExists(hostId: string, dirPath: string) {
    console.log("ensureRemoteDirExists called for", hostId, dirPath);
    if (!dirPath || dirPath === "/" || dirPath === "") return;
    const parent = dirPath.substring(0, dirPath.lastIndexOf("/")) || "/";
    if (parent && parent !== dirPath) {
      await ensureRemoteDirExists(hostId, parent);
    }
    console.log("About to listDirectory", hostId, dirPath);
    const files = await SFTPAPI.listDirectory(hostId, dirPath);
    if (files && Array.isArray(files)) {
      console.log("listDirectory succeeded for", dirPath);
      return;
    }
    // If we reach here, directory does not exist, try to create it
    console.log("Directory does not exist, calling makeDirectory for", dirPath);
    try {
      await SFTPAPI.makeDirectory(hostId, dirPath);
      console.log("makeDirectory called for", dirPath);
    } catch (err) {
      // Ignore error if directory already exists (common SFTP quirk)
      const msg = String(err);
      if (msg.includes("exists") || msg.includes("Failure")) {
        console.warn(
          `Ignored error from makeDirectory (likely already exists):`,
          msg
        );
        return;
      }
      addNotification({
        type: "error",
        title: `Failed to create remote directory`,
        message: msg,
      });
      throw err;
    }
  }

  function toggleRemoteSelection(fileName: string, event?: MouseEvent) {
    if (event && event.ctrlKey) {
      if (selectedRemoteFiles.has(fileName)) {
        selectedRemoteFiles.delete(fileName);
      } else {
        selectedRemoteFiles.add(fileName);
      }
    } else {
      if (selectedRemoteFiles.size === 1 && selectedRemoteFiles.has(fileName)) {
        selectedRemoteFiles.clear();
      } else {
        selectedRemoteFiles.clear();
        selectedRemoteFiles.add(fileName);
      }
    }
    selectedRemoteFiles = new Set(selectedRemoteFiles);
  }

  // Drag and drop handlers
  function handleDragOver(event: DragEvent, target: "local" | "remote") {
    event.preventDefault();
    if (target === "local") {
      dragOverLocal = true;
    } else {
      dragOverRemote = true;
    }
  }

  function handleDragLeave(target: "local" | "remote") {
    if (target === "local") {
      dragOverLocal = false;
    } else {
      dragOverRemote = false;
    }
  }

  async function handleDrop(event: DragEvent, target: "local" | "remote") {
    event.preventDefault();
    dragOverLocal = false;
    dragOverRemote = false;

    console.log("[handleDrop] Drop event triggered", { target, event });
    // Support both internal drag (custom type) and File API drag (from desktop/Explorer)
    const dragType = event.dataTransfer?.getData("text/plain");
    const fileListJson = event.dataTransfer?.getData("application/json");
    let fileNames: string[] = [];
    try {
      if (fileListJson) fileNames = JSON.parse(fileListJson);
    } catch {}

    // File API drag-in (from desktop/Explorer)
    const fileList = event.dataTransfer?.files;
    const items = event.dataTransfer?.items;
    console.log("[handleDrop] fileList:", fileList, "items:", items);

    if (target === "remote" && fileList && fileList.length > 0) {
      console.log(
        "[handleDrop] Handling remote drop with fileList:",
        fileList,
        "items:",
        items
      );
      if (!activeSession?.hostId) {
        addNotification({
          type: "error",
          title: "No active session",
          message: "No active session or hostId for upload.",
        });
        return;
      }
      // Recursively queue files/folders for upload
      const uploadQueue: File[] = [];
      await collectFilesFromDataTransferItems(items, uploadQueue);
      console.log(
        "[handleDrop] uploadQueue after collectFilesFromDataTransferItems:",
        uploadQueue
      );

      // Build set of top-level folders
      const topFolders = new Set<string>();
      for (const file of uploadQueue) {
        const parts = file.name.split("/");
        if (parts.length > 1) topFolders.add(parts[0]);
      }

      // Check for conflicts (files and folders)
      const conflicts: Array<{ name: string; type: "file" | "folder" }> = [];
      const conflictMap: Record<
        string,
        { name: string; type: "file" | "folder" }
      > = {};
      // Check folder conflicts first
      for (const folderName of topFolders) {
        const dirListing = await SFTPAPI.listDirectory(
          activeSession.hostId,
          remotePath
        );
        if (
          dirListing &&
          dirListing.some((f) => f.name === folderName && f.is_dir)
        ) {
          const conflict = { name: folderName, type: "folder" as const };
          conflicts.push(conflict);
          conflictMap[folderName] = conflict;
        }
      }
      // Check file conflicts
      for (const file of uploadQueue) {
        // If file is inside a top-level folder that is a conflict, skip file conflict check
        const topFolder = file.name.includes('/') ? file.name.split('/')[0] : null;
        if (topFolder && conflictMap[topFolder] && conflictMap[topFolder].type === 'folder') {
          continue;
        }
        const remoteFilePath = remotePath.endsWith("/")
          ? remotePath + file.name
          : remotePath + "/" + file.name;
        const parentDir = remoteFilePath.substring(
          0,
          remoteFilePath.lastIndexOf("/")
        );
        await ensureRemoteDirExists(activeSession.hostId, parentDir);
        const baseName = file.name.split("/").pop() || file.name;
        const dirListing = await SFTPAPI.listDirectory(
          activeSession.hostId,
          parentDir
        );
        if (
          dirListing &&
          dirListing.some((f) => f.name === baseName && !f.is_dir)
        ) {
          const conflict = { name: file.name, type: "file" as const };
          conflicts.push(conflict);
          conflictMap[file.name] = conflict;
        }
      }

      // Track user choices for each conflict
  let applyToAllAction: string | null = null;
  const conflictResults: Record<string, string> = {};

      // Handle folder conflicts first
      const folderConflicts = conflicts.filter(c => c.type === 'folder');
      for (let i = 0; i < folderConflicts.length; i++) {
        const folderName = folderConflicts[i].name;
        if (conflictMap[folderName]) {
          let action: string | null = applyToAllAction;
          if (!applyToAllAction) {
            currentConflictIndex = conflicts.findIndex(c => c.name === folderName && c.type === 'folder');
            showConflictDialog = false; await tick(); showConflictDialog = true;
            conflictList = conflicts;
            let dialogResult = await new Promise<{
              action: string;
              applyToAll: boolean;
            }>((resolve) => {
              conflictDialogPromise = resolve;
              conflictList = conflicts;
            });
            action = dialogResult.action;
            if (dialogResult.applyToAll) {
              applyToAllAction = action;
            }
          }
          if (action === "cancel" || action === "skip") {
            conflictResults[folderName] = "cancel";
            continue;
          }
          conflictResults[folderName] = action || "";
        }
      }

      // Handle file conflicts (not inside a conflicting folder)
      const fileConflicts = conflicts.filter(c => c.type === 'file');
      for (let i = 0; i < fileConflicts.length; i++) {
        const fileName = fileConflicts[i].name;
        const topFolder = fileName.includes("/") ? fileName.split("/")[0] : null;
        if (topFolder && conflictResults[topFolder]) continue;
        if (conflictMap[fileName]) {
          let action: string | null = applyToAllAction;
          if (!applyToAllAction) {
            currentConflictIndex = conflicts.findIndex(c => c.name === fileName && c.type === 'file');
            showConflictDialog = false; await tick(); showConflictDialog = true;
            conflictList = conflicts;
            let dialogResult = await new Promise<{ action: string; applyToAll: boolean }>((resolve) => {
              conflictDialogPromise = resolve;
              conflictList = conflicts;
            });
            action = dialogResult.action;
            if (dialogResult.applyToAll) {
              applyToAllAction = action;
            }
          }
          if (action === "cancel" || action === "skip") {
            conflictResults[fileName] = "cancel";
            continue;
          }
          conflictResults[fileName] = action || "";
        }
      }

      // Now upload folders and files, respecting user choices
      // Map for folder renames if 'keep both' is chosen
      const folderRenameMap: Record<string, string> = {};
      for (const folderName of topFolders) {
        const action = conflictResults[folderName] || applyToAllAction;
        if (action === "cancel") {
          console.log(
            `[handleDrop] Skipping folder due to user choice: ${folderName}`
          );
          continue;
        }
        let finalFolderName = folderName;
        if (action === "keep-both") {
          // Find a new folder name
          let n = 1;
          let newName = `${folderName} (${n})`;
          const dirListing = await SFTPAPI.listDirectory(
            activeSession.hostId,
            remotePath
          );
          while (
            dirListing &&
            dirListing.some((f) => f.name === newName && f.is_dir)
          ) {
            n++;
            newName = `${folderName} (${n})`;
          }
          finalFolderName = newName;
          folderRenameMap[folderName] = finalFolderName;
        }
        // Only create the folder if not skipping
        try {
          await ensureRemoteDirExists(
            activeSession.hostId,
            remotePath + (remotePath.endsWith("/") ? "" : "/") + finalFolderName
          );
        } catch (err) {
          console.error(
            `[handleDrop] Failed to create folder ${finalFolderName}:`,
            err
          );
          addNotification({
            type: "error",
            title: `Failed to create folder ${finalFolderName}`,
            message: String(err),
          });
          continue;
        }
      }

      // Now upload files, skipping those in skipped folders
      for (const file of uploadQueue) {
        // If file is in a skipped folder, skip
        const topFolder = file.name.includes("/")
          ? file.name.split("/")[0]
          : null;
        if (topFolder && conflictResults[topFolder] === "cancel") {
          console.log(
            `[handleDrop] Skipping file in skipped folder: ${file.name}`
          );
          continue;
        }
        let remoteFilePathOrig = remotePath.endsWith("/")
          ? remotePath + file.name
          : remotePath + "/" + file.name;
        // If folder was renamed, update file path
        if (topFolder && folderRenameMap[topFolder]) {
          remoteFilePathOrig = remotePath.endsWith("/")
            ? remotePath +
              folderRenameMap[topFolder] +
              file.name.slice(topFolder.length)
            : remotePath +
              "/" +
              folderRenameMap[topFolder] +
              file.name.slice(topFolder.length);
        }
        let remoteFilePath = remoteFilePathOrig;
        const parentDir = remoteFilePath.substring(
          0,
          remoteFilePath.lastIndexOf("/")
        );
        if (conflictMap[file.name]) {
          const action = conflictResults[file.name] || applyToAllAction;
          if (action === "cancel") {
            console.log(
              `[handleDrop] Skipping file due to user choice: ${file.name}`
            );
            continue;
          }
          if (action === "keep-both") {
            const baseName = file.name.split("/").pop() || file.name;
            const extIdx = baseName.lastIndexOf(".");
            const nameOnly = extIdx > 0 ? baseName.slice(0, extIdx) : baseName;
            const ext = extIdx > 0 ? baseName.slice(extIdx) : "";
            let n = 1;
            let newName = `${nameOnly} (${n})${ext}`;
            const dirListing = await SFTPAPI.listDirectory(
              activeSession.hostId,
              parentDir
            );
            while (dirListing && dirListing.some((f) => f.name === newName)) {
              n++;
              newName = `${nameOnly} (${n})${ext}`;
            }
            remoteFilePath = parentDir + "/" + newName;
          }
        }
        console.log("[handleDrop] Attempting uploadFileFromBlob:", {
          hostId: activeSession.hostId,
          file,
          remoteFilePath,
        });
        try {
          await SFTPAPI.uploadFileFromBlob(
            activeSession.hostId,
            file,
            remoteFilePath
          );
          console.log(
            "[handleDrop] uploadFileFromBlob succeeded for",
            file.name
          );
          loadRemoteFiles();
        } catch (err) {
          console.error("[handleDrop] uploadFileFromBlob error:", err);
          addNotification({
            type: "error",
            title: `Failed to upload ${file.name}`,
            message: String(err),
          });
        }
      }
      return;
    }

    if (
      target === "remote" &&
      dragType === "local-to-remote" &&
      fileNames.length > 0
    ) {
      // Internal drag: Upload selected local files/folders to remote (unified conflict logic)
      if (!activeSession?.hostId) {
        addNotification({
          type: "error",
          title: "No active session",
          message: "No active session or hostId for upload.",
        });
        return;
      }
      // Helper to recursively collect all files from a local folder
      async function collectLocalFilesRecursively(
        basePath: string,
        relPath: string
      ): Promise<{ abs: string; rel: string }[]> {
        const results: { abs: string; rel: string }[] = [];
        const sep = basePath.includes("\\") ? "\\" : "/";
        const fullPath =
          basePath + (basePath.endsWith(sep) ? "" : sep) + relPath;
        // Always list the directory contents at this level
        const dirFiles = await SFTPAPI.listLocalDirectory(fullPath);
        if (dirFiles.length === 0) {
          if (relPath && relPath.length > 0) {
            results.push({ abs: fullPath, rel: relPath });
          }
        } else {
          for (const f of dirFiles) {
            if (f.is_dir) {
              const subRel = relPath + sep + f.name;
              const subResults = await collectLocalFilesRecursively(
                basePath,
                subRel
              );
              results.push(...subResults);
            } else {
              const fileRel = relPath ? relPath + sep + f.name : f.name;
              const fileAbs =
                fullPath + (fullPath.endsWith(sep) ? "" : sep) + f.name;
              results.push({ abs: fileAbs, rel: fileRel });
            }
          }
        }
        return results;
      }

      (async () => {
        // Build uploadQueue: File-like objects with .name and .abs
        const uploadQueue: { name: string; abs: string }[] = [];
        const topFolders = new Set<string>();
        for (const fileName of fileNames) {
          const entry = localFiles.find((f) => f.name === fileName);
          if (!entry) continue;
          if (entry.is_dir) {
            topFolders.add(fileName);
            const filesToUpload = await collectLocalFilesRecursively(localPath, fileName);
            for (const file of filesToUpload) {
              uploadQueue.push({ name: file.rel.replace(/\\/g, "/"), abs: file.abs });
            }
          } else {
            let localPathFull = localPath;
            if (!localPathFull.endsWith("/") && !localPathFull.endsWith("\\")) {
              localPathFull += localPathFull.includes("\\") ? "\\" : "/";
            }
            localPathFull += fileName;
            uploadQueue.push({ name: fileName, abs: localPathFull });
          }
        }

        // Conflict detection (same as external drag)
        const conflicts: Array<{ name: string; type: "file" | "folder" }> = [];
        const conflictMap: Record<string, { name: string; type: "file" | "folder" }> = {};
        // Check folder conflicts first
        for (const folderName of topFolders) {
          const dirListing = await SFTPAPI.listDirectory(activeSession.hostId, remotePath);
          if (dirListing && dirListing.some((f) => f.name === folderName && f.is_dir)) {
            const conflict = { name: folderName, type: "folder" as const };
            conflicts.push(conflict);
            conflictMap[folderName] = conflict;
          }
        }
        // Check file conflicts
        for (const file of uploadQueue) {
          const topFolder = file.name.includes("/") ? file.name.split("/")[0] : null;
          if (topFolder && conflictMap[topFolder] && conflictMap[topFolder].type === "folder") {
            continue;
          }
          const remoteFilePath = remotePath.endsWith("/") ? remotePath + file.name : remotePath + "/" + file.name;
          const parentDir = remoteFilePath.substring(0, remoteFilePath.lastIndexOf("/"));
          await ensureRemoteDirExists(activeSession.hostId, parentDir);
          const baseName = file.name.split("/").pop() || file.name;
          const dirListing = await SFTPAPI.listDirectory(activeSession.hostId, parentDir);
          if (dirListing && dirListing.some((f) => f.name === baseName && !f.is_dir)) {
            const conflict = { name: file.name, type: "file" as const };
            conflicts.push(conflict);
            conflictMap[file.name] = conflict;
          }
        }

        // Track user choices for each conflict
        let applyToAllAction: string | null = null;
        const conflictResults: Record<string, string> = {};

        // Handle folder conflicts first
        const folderConflicts = conflicts.filter(c => c.type === 'folder');
        for (let i = 0; i < folderConflicts.length; i++) {
          const folderName = folderConflicts[i].name;
          if (conflictMap[folderName]) {
            let action: string | null = applyToAllAction;
            if (!applyToAllAction) {
              currentConflictIndex = conflicts.findIndex(c => c.name === folderName && c.type === 'folder');
              showConflictDialog = false; await tick(); showConflictDialog = true;
              conflictList = conflicts;
              let dialogResult = await new Promise<{
                action: string;
                applyToAll: boolean;
              }>((resolve) => {
                conflictDialogPromise = resolve;
                conflictList = conflicts;
              });
              action = dialogResult.action;
              if (dialogResult.applyToAll) {
                applyToAllAction = action;
              }
            }
            if (action === "cancel" || action === "skip") {
              conflictResults[folderName] = "cancel";
              continue;
            }
            conflictResults[folderName] = action || "";
          }
        }

        // Handle file conflicts (not inside a conflicting folder)
        const fileConflicts = conflicts.filter(c => c.type === 'file');
        for (let i = 0; i < fileConflicts.length; i++) {
          const fileName = fileConflicts[i].name;
          const topFolder = fileName.includes("/") ? fileName.split("/")[0] : null;
          if (topFolder && conflictResults[topFolder]) continue;
          if (conflictMap[fileName]) {
            let action: string | null = applyToAllAction;
            if (!applyToAllAction) {
              currentConflictIndex = conflicts.findIndex(c => c.name === fileName && c.type === 'file');
              showConflictDialog = false; await tick(); showConflictDialog = true;
              conflictList = conflicts;
              let dialogResult = await new Promise<{ action: string; applyToAll: boolean }>((resolve) => {
                conflictDialogPromise = resolve;
                conflictList = conflicts;
              });
              action = dialogResult.action;
              if (dialogResult.applyToAll) {
                applyToAllAction = action;
              }
            }
            if (action === "cancel" || action === "skip") {
              conflictResults[fileName] = "cancel";
              continue;
            }
            conflictResults[fileName] = action || "";
          }
        }

        // Now upload folders and files, respecting user choices
        // Map for folder renames if 'keep both' is chosen
        const folderRenameMap: Record<string, string> = {};
        for (const folderName of topFolders) {
          const action = conflictResults[folderName] || applyToAllAction;
          if (action === "cancel") {
            console.log(
              `[handleDrop] Skipping folder due to user choice: ${folderName}`
            );
            continue;
          }
          let finalFolderName = folderName;
          if (action === "keep-both") {
            // Find a new folder name
            let n = 1;
            let newName = `${folderName} (${n})`;
            const dirListing = await SFTPAPI.listDirectory(
              activeSession.hostId,
              remotePath
            );
            while (
              dirListing &&
              dirListing.some((f) => f.name === newName && f.is_dir)
            ) {
              n++;
              newName = `${folderName} (${n})`;
            }
            finalFolderName = newName;
            folderRenameMap[folderName] = finalFolderName;
          }
          // Only create the folder if not skipping
          try {
            await ensureRemoteDirExists(
              activeSession.hostId,
              remotePath + (remotePath.endsWith("/") ? "" : "/") + finalFolderName
            );
          } catch (err) {
            console.error(
              `[handleDrop] Failed to create folder ${finalFolderName}:`,
              err
            );
            addNotification({
              type: "error",
              title: `Failed to create folder ${finalFolderName}`,
              message: String(err),
            });
            continue;
          }
        }

        // Now upload files, skipping those in skipped folders
        for (const file of uploadQueue) {
          // If file is in a skipped folder, skip
          const topFolder = file.name.includes("/")
            ? file.name.split("/")[0]
            : null;
          if (topFolder && conflictResults[topFolder] === "cancel") {
            console.log(
              `[handleDrop] Skipping file in skipped folder: ${file.name}`
            );
            continue;
          }
          let remoteFilePathOrig = remotePath.endsWith("/")
            ? remotePath + file.name
            : remotePath + "/" + file.name;
          // If folder was renamed, update file path
          if (topFolder && folderRenameMap[topFolder]) {
            remoteFilePathOrig = remotePath.endsWith("/")
              ? remotePath +
                folderRenameMap[topFolder] +
                file.name.slice(topFolder.length)
              : remotePath +
                "/" +
                folderRenameMap[topFolder] +
                file.name.slice(topFolder.length);
          }
          let remoteFilePath = remoteFilePathOrig;
          const parentDir = remoteFilePath.substring(
            0,
            remoteFilePath.lastIndexOf("/")
          );
          if (conflictMap[file.name]) {
            const action = conflictResults[file.name] || applyToAllAction;
            if (action === "cancel") {
              console.log(
                `[handleDrop] Skipping file due to user choice: ${file.name}`
              );
              continue;
            }
            if (action === "keep-both") {
              const baseName = file.name.split("/").pop() || file.name;
              const extIdx = baseName.lastIndexOf(".");
              const nameOnly = extIdx > 0 ? baseName.slice(0, extIdx) : baseName;
              const ext = extIdx > 0 ? baseName.slice(extIdx) : "";
              let n = 1;
              let newName = `${nameOnly} (${n})${ext}`;
              const dirListing = await SFTPAPI.listDirectory(
                activeSession.hostId,
                parentDir
              );
              while (dirListing && dirListing.some((f) => f.name === newName)) {
                n++;
                newName = `${nameOnly} (${n})${ext}`;
              }
              remoteFilePath = parentDir + "/" + newName;
            }
          }
          console.log("[handleDrop] Attempting uploadFileFromBlob:", {
            hostId: activeSession.hostId,
            file,
            remoteFilePath,
          });
          try {
            let blob = await SFTPAPI.readLocalFileAsBlob(file.abs);
            await SFTPAPI.uploadFileFromBlob(
              activeSession.hostId,
              blob,
              remoteFilePath
            );
            console.log(
              "[handleDrop] uploadFileFromBlob succeeded for",
              file.name
            );
            loadRemoteFiles();
          } catch (err) {
            console.error("[handleDrop] uploadFileFromBlob error:", err);
            addNotification({
              type: "error",
              title: `Failed to upload ${file.name}`,
              message: String(err),
            });
          }
        }
      })();
    } else if (
      target === "local" &&
      dragType === "remote-to-local" &&
      fileNames.length > 0
    ) {
      // Internal drag: Download selected remote files to local
      if (!activeSession?.hostId) {
        addNotification({
          type: "error",
          title: "No active session",
          message: "No active session or hostId for download.",
        });
        return;
      }
      fileNames.forEach((fileName) => {
        const remoteFile = remoteFiles.find((f) => f.name === fileName);
        if (!remoteFile) return;
        let localPathFull = localPath;
        if (!localPathFull.endsWith("/") && !localPathFull.endsWith("\\")) {
          localPathFull += localPathFull.includes("\\") ? "\\" : "/";
        }
        localPathFull += fileName;
        console.debug("Downloading (internal drag) to path:", localPathFull);
        const remoteFilePath = remotePath.endsWith("/")
          ? remotePath + fileName
          : remotePath + "/" + fileName;
        SFTPAPI.downloadFile(
          activeSession.hostId,
          remoteFilePath,
          localPathFull
        )
          .then(() => loadLocalFiles())
          .catch((err) =>
            addNotification({
              type: "error",
              title: `Failed to download ${fileName}`,
              message: String(err),
            })
          );
      });
    }
  }

  // Helper: Recursively collect files from DataTransferItemList (for folder upload)
  async function collectFilesFromDataTransferItems(
    items: DataTransferItemList | undefined,
    outFiles: File[]
  ) {
    console.log(
      "[collectFilesFromDataTransferItems] called with items:",
      items
    );
    if (!items) return;
    const promises: Promise<void>[] = [];
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.kind === "file") {
        const entry = (item as any).webkitGetAsEntry
          ? (item as any).webkitGetAsEntry()
          : null;
        if (entry && entry.isDirectory) {
          console.log(
            "[collectFilesFromDataTransferItems] Directory entry:",
            entry
          );
          promises.push(collectFilesFromDirectoryEntry(entry, "", outFiles));
        } else {
          const file = item.getAsFile();
          if (file) {
            console.log("[collectFilesFromDataTransferItems] File:", file);
            outFiles.push(file);
          }
        }
      }
    }
    await Promise.all(promises);
  }

  // Helper: Recursively collect files from a directory entry (for folder upload)
  async function collectFilesFromDirectoryEntry(
    entry: any,
    pathPrefix: string,
    outFiles: File[]
  ) {
    return new Promise<void>((resolve, reject) => {
      if (entry.isDirectory) {
        const reader = entry.createReader();
        reader.readEntries(async (entries: any[]) => {
          const subPromises = entries.map((subEntry: any) =>
            collectFilesFromDirectoryEntry(
              subEntry,
              pathPrefix + entry.name + "/",
              outFiles
            )
          );
          await Promise.all(subPromises);
          resolve();
        }, reject);
      } else if (entry.isFile) {
        entry.file((file: File) => {
          // Patch file name to include relative path
          Object.defineProperty(file, "name", {
            value: pathPrefix + file.name,
            writable: true,
          });
          outFiles.push(file);
          resolve();
        }, reject);
      } else {
        resolve();
      }
    });
  }

  // Transfer operations - TODO: Implement
  // async function uploadFiles(files: FileItem[]) {
  //   console.log('Uploading files:', files);
  // }

  // async function downloadFiles(files: FileItem[]) {
  //   console.log('Downloading files:', files);
  // }

  // Layout class computation
  $: layoutClasses = {
    hidden: "", // Don't hide here - App.svelte handles conditional rendering
    bottom: "border-t border-slate-700",
    top: "border-b border-slate-700",
    fullscreen: "absolute inset-0 z-50",
  };

  // Load files when component mounts or session changes
  onMount(async () => {
    console.log(
      "SFTPPanel mounted, layout:",
      layout,
      "activeSession:",
      activeSession?.id
    );

    // Only initialize if we don't have saved state
    if (!activeSession || !restoreSessionState(activeSession.id)) {
      await initializeLocalPath();
      loadLocalFiles();
    }
  });

  onDestroy(() => {
    console.log("SFTPPanel destroyed");
  });

  // Reactive loading when activeSession changes
  $: if (activeSession && layout !== "hidden") {
    // Only reload if session actually changed, NOT on layout changes
    const sessionChanged = activeSession?.id !== previousSessionId;
    const firstTimeVisible =
      previousLayout === "hidden" && previousSessionId === activeSession?.id;

    if (sessionChanged) {
      console.log(
        "SFTPPanel: NEW SESSION detected, switching from",
        previousSessionId,
        "to",
        activeSession?.id
      );

      // Save current session state before switching
      if (previousSessionId) {
        saveSessionState(previousSessionId);
      }

      // Try to restore state for new session
      const stateRestored = restoreSessionState(activeSession.id);

      if (!stateRestored) {
        console.log(
          "SFTPPanel: No saved state, loading fresh data for session",
          activeSession?.id
        );
        // Initialize paths for new session
        initializeLocalPath();
        remotePath = "/";

        // Load files for new session
        loadRemoteFiles().catch((error) => {
          console.error("loadRemoteFiles() failed:", error);
        });
        loadLocalFiles();
      } else {
        console.log(
          "SFTPPanel: Restored saved state for session",
          activeSession?.id
        );
      }
    } else if (firstTimeVisible) {
      console.log(
        "SFTPPanel: Panel became visible for existing session, refreshing remote files"
      );
      // Panel became visible for existing session - just refresh remote files
      loadRemoteFiles().catch((error) => {
        console.error("loadRemoteFiles() failed:", error);
      });
    } else {
      console.log("SFTPPanel: Layout change only, no reload needed");
    }

    // Update tracking variables
    previousSessionId = activeSession?.id;
    previousLayout = layout;
  } else if (layout === "hidden") {
    // Save state when panel is hidden
    if (activeSession?.id) {
      saveSessionState(activeSession.id);
    }
    previousLayout = layout;
  }

  // Debug layout changes
  $: {
    console.log("SFTPPanel layout changed to:", layout);
  }

  // Helper functions
  function formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  }
</script>

<!-- Conflict Dialog -->
<ConflictDialog
  show={showConflictDialog}
  conflicts={conflictList}
  currentIndex={currentConflictIndex}
  on:resolve={(e) => handleConflictDialogResolve(e.detail)}
/>

<div
  class="sftp-panel {layoutClasses[
    layout
  ]} bg-slate-900 flex flex-col h-full w-full overflow-hidden select-none"
>
  <!-- Header -->
  <div
    class="flex items-center justify-between p-3 border-b border-slate-700 bg-slate-800 flex-shrink-0"
  >
    <div class="flex items-center gap-2">
      <FolderOpen size={20} class="text-blue-400" />
      <h3 class="text-lg font-semibold">SFTP File Manager</h3>
      <span class="text-sm text-slate-400"
        >({activeSession?.host?.label || "No host"})</span
      >
    </div>

    <div class="flex items-center gap-2">
      <!-- Layout controls -->
      <div class="flex items-center border border-slate-600 rounded">
        <button
          class="p-1 hover:bg-slate-700 {layout === 'top'
            ? 'bg-slate-700'
            : ''}"
          on:click={() => changeLayout("top")}
          title="Top Half"
        >
          <Square size={14} class="rotate-180" />
        </button>
        <button
          class="p-1 hover:bg-slate-700 {layout === 'bottom'
            ? 'bg-slate-700'
            : ''}"
          on:click={() => changeLayout("bottom")}
          title="Bottom Half"
        >
          <Square size={14} />
        </button>
        <button
          class="p-1 hover:bg-slate-700 {layout === 'fullscreen'
            ? 'bg-slate-700'
            : ''}"
          on:click={() => changeLayout("fullscreen")}
          title="Fullscreen"
        >
          <Maximize2 size={14} />
        </button>
      </div>

      <!-- Transfer queue toggle -->
      <button
        class="p-2 hover:bg-slate-700 rounded {showTransferQueue
          ? 'bg-slate-700'
          : ''}"
        on:click={() => (showTransferQueue = !showTransferQueue)}
        title="Transfer Queue"
      >
        <Upload size={16} />
        {#if transferQueue.items.length > 0}
          <span
            class="absolute -top-1 -right-1 bg-blue-500 text-xs rounded-full w-5 h-5 flex items-center justify-center"
          >
            {transferQueue.items.length}
          </span>
        {/if}
      </button>

      <!-- Settings -->
      <button class="p-2 hover:bg-slate-700 rounded" title="SFTP Settings">
        <Settings size={16} />
      </button>

      <!-- Close -->
      <button
        class="p-2 hover:bg-slate-700 rounded text-slate-400 hover:text-white"
        on:click={closePanel}
        title="Close SFTP Panel"
      >
        <X size={16} />
      </button>
    </div>
  </div>

  <!-- Main content with responsive layout -->
  <div class="flex-1 flex min-h-0 overflow-hidden">
    <!-- File browsers -->
    <div class="flex-1 grid grid-cols-2 gap-4 p-4 min-h-0">
      <!-- Local Files -->
      <div
        class="flex flex-col bg-slate-800 rounded-lg border border-slate-700 min-h-0"
      >
        <!-- Local header -->
        <div
          class="flex items-center justify-between p-3 border-b border-slate-700 flex-shrink-0"
        >
          <h4 class="font-medium">Local Files</h4>
          <div class="flex items-center gap-2">
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Go up one directory"
              on:click={goUpLocal}
              disabled={!localPath || localPath === "/" || localPath === ""}
            >
              <ChevronLeft size={14} />
            </button>
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Home"
              on:click={() => navigateLocal("C:\\")}
            >
              <Home size={14} />
            </button>
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Refresh"
              on:click={loadLocalFiles}
            >
              <RefreshCw size={14} class={loadingLocal ? "animate-spin" : ""} />
            </button>
            <button class="p-1 hover:bg-slate-700 rounded" title="More">
              <MoreVertical size={14} />
            </button>
          </div>
        </div>

        <!-- Local breadcrumb with clickable path segments -->
        <div
          class="flex items-center gap-1 p-2 text-sm border-b border-slate-700 flex-shrink-0 overflow-hidden"
        >
          <button
            class="p-1 hover:bg-slate-600 rounded flex-shrink-0"
            on:click={() => navigateToLocalPathSegment(-1)}
            title="Home"
          >
            <Home size={12} />
          </button>
          {#if localPath}
            {#each localPath.split(/[\/\\]/).filter((p) => p) as segment, i}
              <ChevronRight size={12} class="text-slate-500 flex-shrink-0" />
              <button
                class="hover:text-white text-slate-300 hover:bg-slate-600 px-1 py-0.5 rounded truncate max-w-24 flex-shrink-0"
                on:click={() => navigateToLocalPathSegment(i)}
                title={segment}
              >
                {segment}
              </button>
            {/each}
          {/if}
        </div>

        <!-- Local search -->
        <div class="p-2 border-b border-slate-700 flex-shrink-0">
          <div class="relative">
            <Search size={14} class="absolute left-2 top-2 text-slate-500" />
            <input
              type="text"
              placeholder="Search files..."
              class="w-full pl-8 pr-3 py-1 bg-slate-700 border border-slate-600 rounded text-sm"
              bind:value={localSearchQuery}
            />
          </div>
        </div>

        <!-- Local file list with scroll -->
        <div
          class="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-thin scrollbar-thumb-slate-600 custom-scrollbar {dragOverLocal
            ? 'bg-blue-900/20 border-2 border-blue-500'
            : ''}"
          role="listbox"
          aria-label="Local files"
          tabindex="0"
          on:dragover={(e) => handleDragOver(e, "local")}
          on:dragleave={() => handleDragLeave("local")}
          on:drop={(e) => handleDrop(e, "local")}
        >
          {#if loadingLocal}
            <div class="text-center text-slate-400 py-8">
              <RefreshCw
                size={32}
                class="mx-auto mb-2 opacity-50 animate-spin"
              />
              <p>Loading files...</p>
            </div>
          {:else if localFiles.length === 0}
            <div class="text-center text-slate-400 py-8">
              <FolderOpen size={32} class="mx-auto mb-2 opacity-50" />
              <p>No files in this directory</p>
            </div>
          {:else}
            {#each localFiles.filter((f) => f.name
                .toLowerCase()
                .includes(localSearchQuery.toLowerCase())) as file}
              <div
                class="flex items-center gap-2 p-2 rounded hover:bg-slate-700 cursor-pointer {selectedLocalFiles.has(
                  file.name
                )
                  ? 'bg-blue-900/30 ring-2 ring-blue-400'
                  : ''}"
                role="option"
                aria-selected={selectedLocalFiles.has(file.name)}
                tabindex="0"
                on:click={(e) => toggleLocalSelection(file.name, e)}
                on:keydown={(e) =>
                  e.key === "Enter" && toggleLocalSelection(file.name)}
                draggable="true"
                on:dragstart={(e) => handleLocalDragStart(e, file.name)}
                on:dblclick={() => {
                  if (file.is_dir) {
                    const newPath =
                      localPath.endsWith("/") || localPath.endsWith("\\")
                        ? localPath + file.name
                        : localPath +
                          (localPath.includes("\\") ? "\\" : "/") +
                          file.name;
                    navigateLocal(newPath);
                  }
                }}
              >
                {#if file.is_dir}
                  <FolderOpen size={16} class="text-blue-400 flex-shrink-0" />
                {:else}
                  <FileIcon size={16} class="text-slate-400 flex-shrink-0" />
                {/if}
                <span class="flex-1 truncate text-sm">{file.name}</span>
                <span class="text-xs text-slate-500 flex-shrink-0"
                  >{file.is_dir ? "" : formatFileSize(file.size)}</span
                >
              </div>
            {/each}
          {/if}
        </div>
      </div>

      <!-- Remote Files -->
      <div
        class="flex flex-col bg-slate-800 rounded-lg border border-slate-700 min-h-0"
      >
        <!-- Remote header -->
        <div
          class="flex items-center justify-between p-3 border-b border-slate-700 flex-shrink-0"
        >
          <h4 class="font-medium">Remote Files</h4>
          <div class="flex items-center gap-2">
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Go up one directory"
              on:click={goUpRemote}
              disabled={remotePath === "/"}
            >
              <ChevronLeft size={14} />
            </button>
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Home"
              on:click={() => navigateRemote("/")}
            >
              <Home size={14} />
            </button>
            <button
              class="p-1 hover:bg-slate-700 rounded"
              title="Refresh"
              on:click={loadRemoteFiles}
            >
              <RefreshCw
                size={14}
                class={loadingRemote ? "animate-spin" : ""}
              />
            </button>
            <button class="p-1 hover:bg-slate-700 rounded" title="More">
              <MoreVertical size={14} />
            </button>
          </div>
        </div>

        <!-- Remote breadcrumb with clickable path segments -->
        <div
          class="flex items-center gap-1 p-2 text-sm border-b border-slate-700 flex-shrink-0 overflow-hidden"
        >
          <button
            class="p-1 hover:bg-slate-600 rounded flex-shrink-0"
            on:click={() => navigateToRemotePathSegment(-1)}
            title="Root"
          >
            <Home size={12} />
          </button>
          {#each remotePath.split("/").filter((p) => p) as segment, i}
            <ChevronRight size={12} class="text-slate-500 flex-shrink-0" />
            <button
              class="hover:text-white text-slate-300 hover:bg-slate-600 px-1 py-0.5 rounded truncate max-w-24 flex-shrink-0"
              on:click={() => navigateToRemotePathSegment(i)}
              title={segment}
            >
              {segment}
            </button>
          {/each}
        </div>

        <!-- Remote search -->
        <div class="p-2 border-b border-slate-700 flex-shrink-0">
          <div class="relative">
            <Search size={14} class="absolute left-2 top-2 text-slate-500" />
            <input
              type="text"
              placeholder="Search files..."
              class="w-full pl-8 pr-3 py-1 bg-slate-700 border border-slate-600 rounded text-sm"
              bind:value={remoteSearchQuery}
            />
          </div>
        </div>

        <!-- Remote file list with scroll -->
        <div
          class="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-thin scrollbar-thumb-slate-600 custom-scrollbar {dragOverRemote
            ? 'bg-blue-900/20 border-2 border-blue-500'
            : ''}"
          role="listbox"
          aria-label="Remote files"
          tabindex="0"
          on:dragover={(e) => handleDragOver(e, "remote")}
          on:dragleave={() => handleDragLeave("remote")}
          on:drop={(e) => handleDrop(e, "remote")}
        >
          {#if loadingRemote}
            <div class="text-center text-slate-400 py-8">
              <RefreshCw
                size={32}
                class="mx-auto mb-2 opacity-50 animate-spin"
              />
              <p>Loading files...</p>
            </div>
          {:else if remoteFiles.length === 0}
            <div class="text-center text-slate-400 py-8">
              <FolderOpen size={32} class="mx-auto mb-2 opacity-50" />
              <p>No files in this directory</p>
            </div>
          {:else}
            {#each remoteFiles.filter((f) => f.name
                .toLowerCase()
                .includes(remoteSearchQuery.toLowerCase())) as file}
              <div
                class="flex items-center gap-2 p-2 rounded hover:bg-slate-700 cursor-pointer {selectedRemoteFiles.has(
                  file.name
                )
                  ? 'bg-blue-900/30 ring-2 ring-blue-400'
                  : ''}"
                role="option"
                aria-selected={selectedRemoteFiles.has(file.name)}
                tabindex="0"
                on:click={(e) => toggleRemoteSelection(file.name, e)}
                on:keydown={(e) =>
                  e.key === "Enter" && toggleRemoteSelection(file.name)}
                draggable="true"
                on:dragstart={(e) => handleRemoteDragStart(e, file.name)}
                on:dblclick={() => {
                  if (file.is_dir) {
                    const newPath =
                      remotePath === "/"
                        ? "/" + file.name
                        : remotePath + "/" + file.name;
                    navigateRemote(newPath);
                  }
                }}
              >
                {#if file.is_dir}
                  <FolderOpen size={16} class="text-blue-400 flex-shrink-0" />
                {:else}
                  <FileIcon size={16} class="text-slate-400 flex-shrink-0" />
                {/if}
                <span class="flex-1 truncate text-sm">{file.name}</span>
                <span class="text-xs text-slate-500 flex-shrink-0"
                  >{file.is_dir ? "" : formatFileSize(file.size)}</span
                >
              </div>
            {/each}
          {/if}
        </div>
      </div>
    </div>

    <!-- Transfer Queue Sidebar -->
    {#if showTransferQueue}
      <div
        class="flex-1 max-w-xs border-l border-slate-700 bg-slate-800 flex flex-col min-h-0"
      >
        <div class="p-3 border-b border-slate-700 flex-shrink-0">
          <h4 class="font-medium">Transfer Queue</h4>
          <div class="text-xs text-slate-400 mt-1">
            {transferQueue.activeTransfers}/{transferQueue.maxConcurrency} active
          </div>
        </div>

        <div
          class="flex-1 overflow-y-auto p-2 space-y-2 scrollbar-thin scrollbar-thumb-slate-600 scrollbar-track-slate-800"
        >
          {#if transferQueue.items.length === 0}
            <div class="text-center text-slate-400 py-8">
              <Upload size={24} class="mx-auto mb-2 opacity-50" />
              <p>No transfers</p>
            </div>
          {:else}
            {#each transferQueue.items as transfer}
              <div class="bg-slate-700 rounded p-3">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-sm font-medium truncate"
                    >{transfer.sourceFile.name}</span
                  >
                  <div class="flex items-center gap-1">
                    {#if transfer.status === "in-progress"}
                      <button
                        class="p-1 hover:bg-slate-600 rounded"
                        title="Pause"
                      >
                        <Pause size={12} />
                      </button>
                    {:else if transfer.status === "paused"}
                      <button
                        class="p-1 hover:bg-slate-600 rounded"
                        title="Resume"
                      >
                        <Play size={12} />
                      </button>
                    {/if}
                    <button
                      class="p-1 hover:bg-slate-600 rounded"
                      title="Cancel"
                    >
                      <X size={12} />
                    </button>
                  </div>
                </div>

                <div class="text-xs text-slate-400 mb-1">
                  {transfer.direction === "upload" ? "↑" : "↓"}
                  {transfer.progress}%
                </div>

                <div class="w-full bg-slate-600 rounded-full h-1">
                  <div
                    class="bg-blue-500 h-1 rounded-full transition-all"
                    style="width: {transfer.progress}%"
                  ></div>
                </div>
              </div>
            {/each}
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <!-- Action bar -->
  <div class="border-t border-slate-700 p-3 bg-slate-800 flex-shrink-0">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span class="text-sm text-slate-400">
          {selectedLocalFiles.size} local, {selectedRemoteFiles.size} remote selected
        </span>
      </div>

      <div class="flex items-center gap-2">
        <button
          class="px-3 py-1 bg-blue-600 hover:bg-blue-700 rounded text-sm"
          disabled={selectedLocalFiles.size === 0}
        >
          Upload →
        </button>
        <button
          class="px-3 py-1 bg-green-600 hover:bg-green-700 rounded text-sm"
          disabled={selectedRemoteFiles.size === 0}
        >
          ← Download
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .sftp-panel {
    position: relative;
  }

  /* Custom scrollbar styles */
  .scrollbar-thin {
    scrollbar-width: thin;
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 8px;
  }

  .scrollbar-thumb-slate-600::-webkit-scrollbar-thumb {
    background-color: rgb(71 85 105);
    border-radius: 6px;
  }

  .scrollbar-track-slate-800::-webkit-scrollbar-track {
    background-color: rgb(30 41 59);
  }
</style>

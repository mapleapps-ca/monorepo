// FileSyncService for handling file synchronization with version control and tombstone management
// Supports the new fields: version, state, tombstone_version, tombstone_expiry

class FileSyncService {
  constructor() {
    this._apiClient = null;
    this._fileService = null;
    this.isInitialized = false;
    this.isSyncing = false;
    this.lastSyncTime = null;
    this.syncCursor = null;

    // Sync statistics
    this.syncStats = {
      totalFiles: 0,
      newFiles: 0,
      updatedFiles: 0,
      deletedFiles: 0,
      conflictedFiles: 0,
      errorFiles: 0,
      lastSyncDuration: 0,
    };

    // Local sync state storage
    this.SYNC_STORAGE_KEY = "mapleapps_file_sync_state";
    this.CONFLICT_STORAGE_KEY = "mapleapps_file_conflicts";
  }

  // Import dependencies
  async getServices() {
    if (!this._apiClient || !this._fileService) {
      const { default: ApiClient } = await import("./ApiClient.js");
      const { default: FileService } = await import("./FileService.js");
      this._apiClient = ApiClient;
      this._fileService = FileService;
    }
    return { apiClient: this._apiClient, fileService: this._fileService };
  }

  // Initialize sync service
  async initialize() {
    if (this.isInitialized) return;

    try {
      await this.getServices();
      await this.loadSyncState();
      this.isInitialized = true;
      console.log("[FileSyncService] Initialized successfully");
    } catch (error) {
      console.error("[FileSyncService] Failed to initialize:", error);
      throw error;
    }
  }

  // Load sync state from localStorage
  loadSyncState() {
    try {
      const stored = localStorage.getItem(this.SYNC_STORAGE_KEY);
      if (stored) {
        const state = JSON.parse(stored);
        this.lastSyncTime = state.lastSyncTime;
        this.syncCursor = state.syncCursor;
        this.syncStats = { ...this.syncStats, ...state.syncStats };
        console.log("[FileSyncService] Loaded sync state:", state);
      }
    } catch (error) {
      console.warn("[FileSyncService] Failed to load sync state:", error);
    }
  }

  // Save sync state to localStorage
  saveSyncState() {
    try {
      const state = {
        lastSyncTime: this.lastSyncTime,
        syncCursor: this.syncCursor,
        syncStats: this.syncStats,
        savedAt: new Date().toISOString(),
      };
      localStorage.setItem(this.SYNC_STORAGE_KEY, JSON.stringify(state));
    } catch (error) {
      console.warn("[FileSyncService] Failed to save sync state:", error);
    }
  }

  // Get stored file conflicts
  getStoredConflicts() {
    try {
      const stored = localStorage.getItem(this.CONFLICT_STORAGE_KEY);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.warn("[FileSyncService] Failed to load conflicts:", error);
      return [];
    }
  }

  // Store file conflict for later resolution
  storeConflict(conflict) {
    try {
      const conflicts = this.getStoredConflicts();
      conflicts.push({
        ...conflict,
        detectedAt: new Date().toISOString(),
        id: `${conflict.fileId}_${Date.now()}`,
      });
      localStorage.setItem(
        this.CONFLICT_STORAGE_KEY,
        JSON.stringify(conflicts),
      );
      console.log(
        "[FileSyncService] Stored conflict for file:",
        conflict.fileId,
      );
    } catch (error) {
      console.warn("[FileSyncService] Failed to store conflict:", error);
    }
  }

  // Clear resolved conflicts
  clearResolvedConflicts(fileIds) {
    try {
      const conflicts = this.getStoredConflicts();
      const remaining = conflicts.filter((c) => !fileIds.includes(c.fileId));
      localStorage.setItem(
        this.CONFLICT_STORAGE_KEY,
        JSON.stringify(remaining),
      );
    } catch (error) {
      console.warn("[FileSyncService] Failed to clear conflicts:", error);
    }
  }

  // Parse cursor for debugging
  parseCursor(cursor) {
    if (!cursor) return null;
    try {
      const decoded = atob(cursor);
      return JSON.parse(decoded);
    } catch (error) {
      console.warn("[FileSyncService] Failed to parse cursor:", error);
      return null;
    }
  }

  // Process a single file sync item
  processFileSyncItem(fileSyncItem) {
    const processedFile = {
      // Standard file fields
      id: fileSyncItem.id,
      collection_id: fileSyncItem.collection_id,

      // New sync fields
      version: fileSyncItem.version || 1,
      modified_at: fileSyncItem.modified_at,
      state: fileSyncItem.state || "active",
      tombstone_version: fileSyncItem.tombstone_version || 0,
      tombstone_expiry: fileSyncItem.tombstone_expiry || "0001-01-01T00:00:00Z",

      // Computed fields for easier handling
      _is_deleted: fileSyncItem.state === "deleted",
      _is_archived: fileSyncItem.state === "archived",
      _is_pending: fileSyncItem.state === "pending",
      _is_active: fileSyncItem.state === "active",
      _has_tombstone: (fileSyncItem.tombstone_version || 0) > 0,
      _tombstone_expired: this.isTombstoneExpired(
        fileSyncItem.tombstone_expiry,
      ),

      // Sync metadata
      _synced_at: new Date().toISOString(),
      _sync_action: this.determineSyncAction(fileSyncItem),
    };

    return processedFile;
  }

  // Check if tombstone has expired
  isTombstoneExpired(tombstoneExpiry) {
    if (!tombstoneExpiry || tombstoneExpiry === "0001-01-01T00:00:00Z") {
      return false;
    }
    try {
      return new Date(tombstoneExpiry) < new Date();
    } catch {
      return false;
    }
  }

  // Determine what sync action is needed for a file
  determineSyncAction(fileSyncItem) {
    const { fileService } = this.getServices();
    const localFile = fileService.getCachedFile(fileSyncItem.id);

    if (!localFile) {
      return fileSyncItem.state === "deleted" ? "ignore_deleted" : "download";
    }

    const localVersion = localFile.version || 1;
    const remoteVersion = fileSyncItem.version || 1;

    if (remoteVersion > localVersion) {
      return fileSyncItem.state === "deleted" ? "mark_deleted" : "update";
    } else if (remoteVersion < localVersion) {
      return "conflict_local_newer";
    } else {
      // Same version - check if states differ
      if (localFile.state !== fileSyncItem.state) {
        return "sync_state_change";
      }
      return "no_change";
    }
  }

  // Handle version conflicts
  async handleVersionConflict(localFile, remoteFile) {
    const conflict = {
      fileId: localFile.id,
      type: "version_conflict",
      localVersion: localFile.version,
      remoteVersion: remoteFile.version,
      localState: localFile.state,
      remoteState: remoteFile.state,
      localModified: localFile.modified_at,
      remoteModified: remoteFile.modified_at,
    };

    this.storeConflict(conflict);
    this.syncStats.conflictedFiles++;

    console.warn("[FileSyncService] Version conflict detected:", conflict);

    // For now, remote wins (server authority)
    // TODO: Implement user-choice conflict resolution
    return "remote_wins";
  }

  // Sync files from server
  async syncFromServer(cursor = null, limit = 1000) {
    const { apiClient } = await this.getServices();

    try {
      console.log("[FileSyncService] Syncing from server:", { cursor, limit });

      const params = new URLSearchParams({ limit: limit.toString() });
      if (cursor) {
        params.append("cursor", cursor);
      }

      const response = await apiClient.getMapleFile(`/sync/files?${params}`);

      const syncResults = {
        files: [],
        processedCount: 0,
        newFiles: 0,
        updatedFiles: 0,
        deletedFiles: 0,
        conflictedFiles: 0,
        errorFiles: 0,
        hasMore: response.has_more || false,
        nextCursor: response.next_cursor || null,
      };

      if (response.files && response.files.length > 0) {
        for (const fileSyncItem of response.files) {
          try {
            const processedFile = this.processFileSyncItem(fileSyncItem);

            // Handle based on sync action
            switch (processedFile._sync_action) {
              case "download":
                syncResults.newFiles++;
                break;
              case "update":
              case "sync_state_change":
                syncResults.updatedFiles++;
                break;
              case "mark_deleted":
              case "ignore_deleted":
                syncResults.deletedFiles++;
                break;
              case "conflict_local_newer":
                const { fileService } = await this.getServices();
                const localFile = fileService.getCachedFile(fileSyncItem.id);
                await this.handleVersionConflict(localFile, processedFile);
                syncResults.conflictedFiles++;
                break;
              case "no_change":
                // No action needed
                break;
              default:
                console.warn(
                  "[FileSyncService] Unknown sync action:",
                  processedFile._sync_action,
                );
            }

            syncResults.files.push(processedFile);
            syncResults.processedCount++;
          } catch (fileError) {
            console.error(
              "[FileSyncService] Error processing file:",
              fileSyncItem.id,
              fileError,
            );
            syncResults.errorFiles++;
          }
        }
      }

      console.log("[FileSyncService] Sync batch completed:", syncResults);
      return syncResults;
    } catch (error) {
      console.error("[FileSyncService] Sync from server failed:", error);
      throw error;
    }
  }

  // Perform full synchronization
  async performFullSync(onProgress = null) {
    if (this.isSyncing) {
      throw new Error("Sync already in progress");
    }

    this.isSyncing = true;
    const startTime = Date.now();

    // Reset stats
    this.syncStats = {
      totalFiles: 0,
      newFiles: 0,
      updatedFiles: 0,
      deletedFiles: 0,
      conflictedFiles: 0,
      errorFiles: 0,
      lastSyncDuration: 0,
    };

    try {
      console.log("[FileSyncService] Starting full sync...");

      let cursor = null;
      let hasMore = true;
      let totalProcessed = 0;
      const batchSize = 1000;

      while (hasMore) {
        if (onProgress) {
          onProgress({
            phase: "syncing",
            processedFiles: totalProcessed,
            currentBatch: Math.floor(totalProcessed / batchSize) + 1,
          });
        }

        const batchResults = await this.syncFromServer(cursor, batchSize);

        // Accumulate stats
        this.syncStats.totalFiles += batchResults.processedCount;
        this.syncStats.newFiles += batchResults.newFiles;
        this.syncStats.updatedFiles += batchResults.updatedFiles;
        this.syncStats.deletedFiles += batchResults.deletedFiles;
        this.syncStats.conflictedFiles += batchResults.conflictedFiles;
        this.syncStats.errorFiles += batchResults.errorFiles;

        totalProcessed += batchResults.processedCount;
        cursor = batchResults.nextCursor;
        hasMore = batchResults.hasMore;

        // Update sync cursor for resumable sync
        this.syncCursor = cursor;
        this.saveSyncState();
      }

      this.lastSyncTime = new Date().toISOString();
      this.syncStats.lastSyncDuration = Date.now() - startTime;
      this.syncCursor = null; // Reset cursor after successful full sync
      this.saveSyncState();

      if (onProgress) {
        onProgress({
          phase: "completed",
          processedFiles: totalProcessed,
          duration: this.syncStats.lastSyncDuration,
          stats: this.syncStats,
        });
      }

      console.log("[FileSyncService] Full sync completed:", this.syncStats);
      return this.syncStats;
    } catch (error) {
      console.error("[FileSyncService] Full sync failed:", error);
      this.saveSyncState(); // Save partial progress
      throw error;
    } finally {
      this.isSyncing = false;
    }
  }

  // Perform incremental sync (since last sync)
  async performIncrementalSync(onProgress = null) {
    if (this.isSyncing) {
      throw new Error("Sync already in progress");
    }

    // If no previous sync, perform full sync
    if (!this.lastSyncTime) {
      return this.performFullSync(onProgress);
    }

    console.log(
      "[FileSyncService] Starting incremental sync since:",
      this.lastSyncTime,
    );

    // For incremental sync, we can use the stored cursor or start fresh
    const cursor = this.syncCursor;
    return this.performFullSync(onProgress); // For now, same as full sync
  }

  // Get sync status
  getSyncStatus() {
    return {
      isInitialized: this.isInitialized,
      isSyncing: this.isSyncing,
      lastSyncTime: this.lastSyncTime,
      syncCursor: this.syncCursor,
      stats: this.syncStats,
      conflicts: this.getStoredConflicts(),
      hasConflicts: this.getStoredConflicts().length > 0,
    };
  }

  // Get files that need manual conflict resolution
  getConflictedFiles() {
    return this.getStoredConflicts();
  }

  // Resolve a file conflict
  async resolveConflict(conflictId, resolution) {
    const conflicts = this.getStoredConflicts();
    const conflict = conflicts.find((c) => c.id === conflictId);

    if (!conflict) {
      throw new Error("Conflict not found");
    }

    console.log(
      "[FileSyncService] Resolving conflict:",
      conflictId,
      "with:",
      resolution,
    );

    try {
      const { fileService } = await this.getServices();

      switch (resolution) {
        case "use_local":
          // Keep local version, possibly upload changes
          console.log(
            "[FileSyncService] Using local version for:",
            conflict.fileId,
          );
          break;

        case "use_remote":
          // Download and use remote version
          console.log(
            "[FileSyncService] Using remote version for:",
            conflict.fileId,
          );
          // Force refresh the file from server
          fileService.removeCachedFile(conflict.fileId);
          break;

        case "create_copy":
          // Create a copy with both versions
          console.log("[FileSyncService] Creating copy for:", conflict.fileId);
          break;

        default:
          throw new Error("Invalid resolution type");
      }

      // Remove resolved conflict
      this.clearResolvedConflicts([conflict.fileId]);

      return true;
    } catch (error) {
      console.error("[FileSyncService] Failed to resolve conflict:", error);
      throw error;
    }
  }

  // Clear all sync data
  clearSyncData() {
    this.lastSyncTime = null;
    this.syncCursor = null;
    this.syncStats = {
      totalFiles: 0,
      newFiles: 0,
      updatedFiles: 0,
      deletedFiles: 0,
      conflictedFiles: 0,
      errorFiles: 0,
      lastSyncDuration: 0,
    };

    localStorage.removeItem(this.SYNC_STORAGE_KEY);
    localStorage.removeItem(this.CONFLICT_STORAGE_KEY);

    console.log("[FileSyncService] Sync data cleared");
  }

  // Check if sync is needed (heuristic)
  shouldSync() {
    if (!this.lastSyncTime) return true;

    const lastSync = new Date(this.lastSyncTime);
    const now = new Date();
    const timeDiff = now.getTime() - lastSync.getTime();

    // Sync if more than 5 minutes since last sync
    return timeDiff > 5 * 60 * 1000;
  }

  // Get recommended sync action
  getRecommendedSyncAction() {
    const status = this.getSyncStatus();

    if (!status.isInitialized) {
      return "initialize";
    }

    if (status.isSyncing) {
      return "wait";
    }

    if (status.hasConflicts) {
      return "resolve_conflicts";
    }

    if (!status.lastSyncTime) {
      return "full_sync";
    }

    if (this.shouldSync()) {
      return "incremental_sync";
    }

    return "no_sync_needed";
  }

  // Get human-readable sync summary
  getSyncSummary() {
    const status = this.getSyncStatus();
    const action = this.getRecommendedSyncAction();

    return {
      status: status,
      recommendedAction: action,
      summary: this.formatSyncSummary(status, action),
    };
  }

  // Format sync summary for display
  formatSyncSummary(status, action) {
    if (!status.isInitialized) {
      return "Sync service not initialized";
    }

    if (status.isSyncing) {
      return "Synchronization in progress...";
    }

    if (status.hasConflicts) {
      return `${status.conflicts.length} file conflicts need resolution`;
    }

    if (!status.lastSyncTime) {
      return "No previous sync - full sync recommended";
    }

    const lastSync = new Date(status.lastSyncTime);
    const timeSince = Math.floor(
      (Date.now() - lastSync.getTime()) / (60 * 1000),
    );

    if (timeSince < 1) {
      return "Sync is up to date";
    } else if (timeSince < 60) {
      return `Last synced ${timeSince} minutes ago`;
    } else {
      const hoursSince = Math.floor(timeSince / 60);
      return `Last synced ${hoursSince} hours ago - sync recommended`;
    }
  }
}

// Export singleton instance
export default new FileSyncService();

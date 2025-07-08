// File: monorepo/web/maplefile-frontend/src/services/Manager/File/DeleteFileManager.js
// Delete File Manager - Orchestrates API, Storage, and Crypto services for file deletion operations

import DeleteFileAPIService from "../../API/File/DeleteFileAPIService.js";
import DeleteFileStorageService from "../../Storage/File/DeleteFileStorageService.js";

class DeleteFileManager {
  constructor(
    authManager,
    getCollectionManager = null,
    listCollectionManager = null,
  ) {
    // DeleteFileManager depends on AuthManager and collection managers
    this.authManager = authManager;
    this.getCollectionManager = getCollectionManager;
    this.listCollectionManager = listCollectionManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new DeleteFileAPIService(authManager);
    this.storageService = new DeleteFileStorageService();

    // Event listeners for deletion events
    this.fileDeletionListeners = new Set();

    console.log(
      "[DeleteFileManager] File deletion manager initialized with AuthManager and collection managers",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[DeleteFileManager] Initializing file deletion manager...");

      // Initialize crypto services
      const { default: FileCryptoService } = await import(
        "../../Crypto/FileCryptoService.js"
      );
      await FileCryptoService.initialize();
      this.fileCryptoService = FileCryptoService;

      // Initialize collection crypto service for collection keys
      const { default: CollectionCryptoService } = await import(
        "../../Crypto/CollectionCryptoService.js"
      );
      await CollectionCryptoService.initialize();
      this.collectionCryptoService = CollectionCryptoService;

      console.log(
        "[DeleteFileManager] File deletion manager initialized successfully",
      );
      console.log(
        "[DeleteFileManager] FileCryptoService available:",
        !!this.fileCryptoService,
      );
      console.log(
        "[DeleteFileManager] CollectionCryptoService available:",
        !!this.collectionCryptoService,
      );
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to initialize file deletion manager:",
        error,
      );
    }
  }

  // === File States Constants ===

  get FILE_STATES() {
    return {
      PENDING: "pending",
      ACTIVE: "active",
      DELETED: "deleted",
      ARCHIVED: "archived",
    };
  }

  // === Core File Deletion Methods ===

  // Soft delete a single file
  async deleteFile(fileId, reason = null, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteFileManager] === Starting File Deletion Workflow ===",
      );
      console.log("[DeleteFileManager] File ID:", fileId);
      console.log("[DeleteFileManager] Reason:", reason);
      console.log("[DeleteFileManager] Force refresh:", forceRefresh);

      // STEP 1: Check if file can be deleted
      if (!forceRefresh) {
        const cachedCheck = this.storageService.getDeletionOperation(
          `check_${fileId}`,
        );
        if (cachedCheck && !cachedCheck.operationData.can_delete) {
          throw new Error("File cannot be deleted based on cached check");
        }
      }

      // STEP 2: Call API to delete file
      console.log("[DeleteFileManager] Calling API to delete file");
      const response = await this.apiService.deleteFile(fileId, reason);

      console.log("[DeleteFileManager] API response:", response);

      // Handle the actual API response format: {success: true, message: "..."}
      if (!response || !response.success) {
        throw new Error(
          "Deletion failed: " + (response?.message || "Unknown error"),
        );
      }

      console.log(
        "[DeleteFileManager] File deletion API call successful:",
        fileId,
      );

      // Since the API doesn't return the updated file, we need to create a mock response
      // In a real scenario, you might want to call getFile() to get the updated state
      // For now, we'll create a normalized response based on the successful deletion
      const deletedFile = {
        id: fileId,
        state: this.FILE_STATES.DELETED,
        version: 1, // We don't know the actual version from this response
        tombstone_version: 1,
        tombstone_expiry: new Date(
          Date.now() + 30 * 24 * 60 * 60 * 1000,
        ).toISOString(), // 30 days from now
        modified_at: new Date().toISOString(),
        name: "[Deleted File]", // We don't have the name from the API response
      };
      console.log(
        `[DeleteFileManager] File deleted successfully: ${fileId} (v${deletedFile.version}, ${deletedFile.state})`,
      );

      // STEP 3: Normalize file with computed properties
      const normalizedFile = this.fileCryptoService.normalizeFile(deletedFile);

      // STEP 4: Store deletion history
      this.storageService.storeDeletionHistory(
        fileId,
        {
          ...response,
          reason,
          deleted_at: new Date().toISOString(),
        },
        {
          operation: "delete",
          operation_time: new Date().toISOString(),
        },
      );

      // STEP 5: Store tombstone data if present
      if (normalizedFile.tombstone_version && normalizedFile.tombstone_expiry) {
        this.storageService.storeTombstoneData(
          fileId,
          {
            tombstone_version: normalizedFile.tombstone_version,
            tombstone_expiry: normalizedFile.tombstone_expiry,
            state: normalizedFile.state,
            version: normalizedFile.version,
            reason,
          },
          {
            created_at: new Date().toISOString(),
          },
        );
      }

      // STEP 6: Update file state in related caches
      this.storageService.updateFileStateInCaches(
        fileId,
        normalizedFile.state,
        normalizedFile.version,
        normalizedFile.tombstone_version
          ? {
              tombstone_version: normalizedFile.tombstone_version,
              tombstone_expiry: normalizedFile.tombstone_expiry,
            }
          : null,
      );

      // STEP 7: Clear any cached file details to force refresh
      try {
        const { default: GetFileStorageService } = await import(
          "../../Storage/File/GetFileStorageService.js"
        );
        GetFileStorageService.clearFileCache(fileId);
      } catch (storageError) {
        console.warn(
          "[DeleteFileManager] Could not clear file cache:",
          storageError,
        );
      }

      this.notifyFileDeletionListeners("file_deleted", {
        fileId,
        newState: normalizedFile.state,
        newVersion: normalizedFile.version,
        hasTombstone: !!normalizedFile.tombstone_version,
        tombstoneExpiry: normalizedFile.tombstone_expiry,
        reason,
      });

      console.log(
        "[DeleteFileManager] File deletion workflow completed successfully:",
        fileId,
      );

      return normalizedFile;
    } catch (error) {
      console.error("[DeleteFileManager] Failed to delete file:", error);

      this.notifyFileDeletionListeners("file_deletion_failed", {
        fileId,
        error: error.message,
        reason,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Restore a soft-deleted file
  async restoreFile(fileId, reason = null, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteFileManager] === Starting File Restoration Workflow ===",
      );
      console.log("[DeleteFileManager] File ID:", fileId);
      console.log("[DeleteFileManager] Reason:", reason);

      // STEP 1: Check if file can be restored
      const tombstoneData = this.storageService.getTombstoneData(fileId);
      if (tombstoneData && !this.canRestoreFile(tombstoneData.tombstoneData)) {
        throw new Error("File cannot be restored - tombstone may have expired");
      }

      // STEP 2: Call API to restore file
      console.log("[DeleteFileManager] Calling API to restore file");
      const response = await this.apiService.restoreFile(fileId, reason);

      if (!response || !response.file) {
        throw new Error("Invalid restoration response from API");
      }

      const restoredFile = response.file;
      console.log(
        `[DeleteFileManager] File restored successfully: ${fileId} (v${restoredFile.version}, ${restoredFile.state})`,
      );

      // STEP 3: Normalize file with computed properties
      const normalizedFile = this.fileCryptoService.normalizeFile(restoredFile);

      // STEP 4: Store restoration history
      this.storageService.storeDeletionHistory(
        fileId,
        {
          ...response,
          reason,
          restored_at: new Date().toISOString(),
        },
        {
          operation: "restore",
          operation_time: new Date().toISOString(),
        },
      );

      // STEP 5: Clear tombstone data since file is restored
      this.storageService.removeTombstoneData(fileId);

      // STEP 6: Update file state in related caches
      this.storageService.updateFileStateInCaches(
        fileId,
        normalizedFile.state,
        normalizedFile.version,
        null, // No tombstone after restoration
      );

      // STEP 7: Clear any cached file details to force refresh
      try {
        const { default: GetFileStorageService } = await import(
          "../../Storage/File/GetFileStorageService.js"
        );
        GetFileStorageService.clearFileCache(fileId);
      } catch (storageError) {
        console.warn(
          "[DeleteFileManager] Could not clear file cache:",
          storageError,
        );
      }

      this.notifyFileDeletionListeners("file_restored", {
        fileId,
        newState: normalizedFile.state,
        newVersion: normalizedFile.version,
        reason,
      });

      console.log(
        "[DeleteFileManager] File restoration workflow completed successfully:",
        fileId,
      );

      return normalizedFile;
    } catch (error) {
      console.error("[DeleteFileManager] Failed to restore file:", error);

      this.notifyFileDeletionListeners("file_restoration_failed", {
        fileId,
        error: error.message,
        reason,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Archive a file (different from delete)
  async archiveFile(fileId, reason = null) {
    try {
      this.isLoading = true;
      console.log("[DeleteFileManager] === Starting File Archive Workflow ===");
      console.log("[DeleteFileManager] File ID:", fileId);

      // Call API to archive file
      const response = await this.apiService.archiveFile(fileId, reason);

      // Handle the actual API response format: {success: true, message: "..."}
      if (!response || !response.success) {
        throw new Error(
          "Archive failed: " + (response?.message || "Unknown error"),
        );
      }

      console.log(
        "[DeleteFileManager] File archive API call successful:",
        fileId,
      );

      // Check if API returned the file object, otherwise create a mock response
      let archivedFile;

      if (response.file) {
        // API returned the updated file object
        archivedFile = response.file;
        console.log("[DeleteFileManager] Using file object from API response");
      } else {
        // API didn't return file object, create a mock response
        console.log(
          "[DeleteFileManager] API didn't return file object, creating mock response",
        );
        archivedFile = {
          id: fileId,
          state: this.FILE_STATES.ARCHIVED,
          version: 1, // We don't know the actual version from this response
          modified_at: new Date().toISOString(),
          name: "[Archived File]", // We don't have the name from the API response
        };
      }

      // Normalize and update caches
      const normalizedFile = this.fileCryptoService.normalizeFile(archivedFile);

      this.storageService.updateFileStateInCaches(
        fileId,
        normalizedFile.state,
        normalizedFile.version,
      );

      this.notifyFileDeletionListeners("file_archived", {
        fileId,
        newState: normalizedFile.state,
        newVersion: normalizedFile.version,
        reason,
      });

      console.log("[DeleteFileManager] File archived successfully:", fileId);
      return normalizedFile;
    } catch (error) {
      console.error("[DeleteFileManager] Failed to archive file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Unarchive a file
  async unarchiveFile(fileId, reason = null) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteFileManager] === Starting File Unarchive Workflow ===",
      );

      const response = await this.apiService.unarchiveFile(fileId, reason);

      // Handle the actual API response format: {success: true, message: "..."}
      if (!response || !response.success) {
        throw new Error(
          "Unarchive failed: " + (response?.message || "Unknown error"),
        );
      }

      console.log(
        "[DeleteFileManager] File unarchive API call successful:",
        fileId,
      );

      // Check if API returned the file object, otherwise create a mock response
      let unarchivedFile;

      if (response.file) {
        // API returned the updated file object
        unarchivedFile = response.file;
        console.log("[DeleteFileManager] Using file object from API response");
      } else {
        // API didn't return file object, create a mock response
        console.log(
          "[DeleteFileManager] API didn't return file object, creating mock response",
        );
        unarchivedFile = {
          id: fileId,
          state: this.FILE_STATES.ACTIVE,
          version: 1, // We don't know the actual version from this response
          modified_at: new Date().toISOString(),
          name: "[Unarchived File]", // We don't have the name from the API response
        };
      }

      const normalizedFile =
        this.fileCryptoService.normalizeFile(unarchivedFile);

      this.storageService.updateFileStateInCaches(
        fileId,
        normalizedFile.state,
        normalizedFile.version,
      );

      this.notifyFileDeletionListeners("file_unarchived", {
        fileId,
        newState: normalizedFile.state,
        newVersion: normalizedFile.version,
        reason,
      });

      console.log("[DeleteFileManager] File unarchived successfully:", fileId);
      return normalizedFile;
    } catch (error) {
      console.error("[DeleteFileManager] Failed to unarchive file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Batch Operations ===

  // Delete multiple files
  async deleteMultipleFiles(fileIds, reason = null) {
    try {
      this.isLoading = true;
      console.log("[DeleteFileManager] === Starting Batch Delete Workflow ===");
      console.log("[DeleteFileManager] File count:", fileIds.length);

      const batchId = `batch_delete_${Date.now()}`;
      const response = await this.apiService.deleteMultipleFiles(
        fileIds,
        reason,
      );

      // Store batch operation result
      this.storageService.storeBatchOperation(
        batchId,
        {
          operation_type: "batch_delete",
          file_ids: fileIds,
          reason,
          ...response,
        },
        {
          operation_time: new Date().toISOString(),
        },
      );

      // Process individual results
      const results = response.results || [];
      results.forEach((result) => {
        if (result.success && result.file) {
          const normalizedFile = this.fileCryptoService.normalizeFile(
            result.file,
          );

          // Update individual file caches
          this.storageService.updateFileStateInCaches(
            result.file.id,
            normalizedFile.state,
            normalizedFile.version,
            normalizedFile.tombstone_version
              ? {
                  tombstone_version: normalizedFile.tombstone_version,
                  tombstone_expiry: normalizedFile.tombstone_expiry,
                }
              : null,
          );

          // Store individual tombstone data
          if (normalizedFile.tombstone_version) {
            this.storageService.storeTombstoneData(result.file.id, {
              tombstone_version: normalizedFile.tombstone_version,
              tombstone_expiry: normalizedFile.tombstone_expiry,
              state: normalizedFile.state,
              version: normalizedFile.version,
              reason,
              batch_id: batchId,
            });
          }
        }
      });

      const successCount = results.filter((r) => r.success).length;
      const errorCount = results.filter((r) => !r.success).length;

      this.notifyFileDeletionListeners("batch_delete_completed", {
        batchId,
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
        results,
        reason,
      });

      console.log("[DeleteFileManager] Batch delete completed:", {
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
      });

      return {
        batchId,
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
        results,
      };
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to delete multiple files:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Restore multiple files
  async restoreMultipleFiles(fileIds, reason = null) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteFileManager] === Starting Batch Restore Workflow ===",
      );

      const batchId = `batch_restore_${Date.now()}`;
      const response = await this.apiService.restoreMultipleFiles(
        fileIds,
        reason,
      );

      // Store batch operation result
      this.storageService.storeBatchOperation(
        batchId,
        {
          operation_type: "batch_restore",
          file_ids: fileIds,
          reason,
          ...response,
        },
        {
          operation_time: new Date().toISOString(),
        },
      );

      // Process individual results
      const results = response.results || [];
      results.forEach((result) => {
        if (result.success && result.file) {
          const normalizedFile = this.fileCryptoService.normalizeFile(
            result.file,
          );

          // Update individual file caches
          this.storageService.updateFileStateInCaches(
            result.file.id,
            normalizedFile.state,
            normalizedFile.version,
          );

          // Clear tombstone data since file is restored
          this.storageService.removeTombstoneData(result.file.id);
        }
      });

      const successCount = results.filter((r) => r.success).length;
      const errorCount = results.filter((r) => !r.success).length;

      this.notifyFileDeletionListeners("batch_restore_completed", {
        batchId,
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
        results,
        reason,
      });

      console.log("[DeleteFileManager] Batch restore completed:", {
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
      });

      return {
        batchId,
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
        results,
      };
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to restore multiple files:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Permanent Deletion ===

  // Permanently delete a file (only if tombstone expired or special permissions)
  async permanentlyDeleteFile(fileId, reason = null) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteFileManager] === Starting Permanent Deletion Workflow ===",
      );

      // Check if file can be permanently deleted
      const tombstoneData = this.storageService.getTombstoneData(fileId);
      if (
        tombstoneData &&
        !this.canPermanentlyDeleteFile(tombstoneData.tombstoneData)
      ) {
        throw new Error(
          "File cannot be permanently deleted - tombstone has not expired",
        );
      }

      const response = await this.apiService.permanentlyDeleteFile(
        fileId,
        reason,
      );

      // Remove all cached data for this file
      this.storageService.clearFileCache(fileId);

      // Clear from main file cache too
      try {
        const { default: GetFileStorageService } = await import(
          "../../Storage/File/GetFileStorageService.js"
        );
        GetFileStorageService.clearFileCache(fileId);
      } catch (storageError) {
        console.warn(
          "[DeleteFileManager] Could not clear file cache:",
          storageError,
        );
      }

      this.notifyFileDeletionListeners("file_permanently_deleted", {
        fileId,
        reason,
      });

      console.log("[DeleteFileManager] File permanently deleted:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to permanently delete file:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === File State Checks ===

  // Check if a file can be deleted
  canDeleteFile(file) {
    return file._is_active || file._is_archived;
  }

  // Check if a file can be restored
  canRestoreFile(file) {
    if (typeof file.tombstone_expiry === "string") {
      // Handle file object
      return (
        file._has_tombstone && !file._tombstone_expired && file._is_deleted
      );
    } else {
      // Handle tombstone data object
      const now = new Date();
      const expiryDate = new Date(file.tombstone_expiry);
      return (
        file.tombstone_version > 0 &&
        (expiryDate > now || file.tombstone_expiry === "0001-01-01T00:00:00Z")
      );
    }
  }

  // Check if a file can be permanently deleted
  canPermanentlyDeleteFile(file) {
    if (typeof file.tombstone_expiry === "string") {
      // Handle file object
      return (
        file._tombstone_expired || (file._has_tombstone && file._is_deleted)
      );
    } else {
      // Handle tombstone data object
      const now = new Date();
      const expiryDate = new Date(file.tombstone_expiry);
      return (
        file.tombstone_version > 0 &&
        expiryDate < now &&
        file.tombstone_expiry !== "0001-01-01T00:00:00Z"
      );
    }
  }

  // Check if a file can be archived
  canArchiveFile(file) {
    return file._is_active;
  }

  // Check if a file can be unarchived
  canUnarchiveFile(file) {
    return file._is_archived;
  }

  // Get file deletion information
  getFileDeletionInfo(file) {
    return {
      canDelete: this.canDeleteFile(file),
      canRestore: this.canRestoreFile(file),
      canPermanentlyDelete: this.canPermanentlyDeleteFile(file),
      canArchive: this.canArchiveFile(file),
      canUnarchive: this.canUnarchiveFile(file),
      hasTombstone: file._has_tombstone,
      tombstoneVersion: file.tombstone_version || 0,
      tombstoneExpiry: file.tombstone_expiry,
      isExpired: file._tombstone_expired,
      state: file.state,
      version: file.version,
    };
  }

  // === Tombstone Management ===

  // Get expired tombstones
  getExpiredTombstones() {
    return this.storageService.getExpiredTombstones();
  }

  // Get restorable files
  getRestorableFiles() {
    return this.storageService.getRestorableFiles();
  }

  // Extend tombstone expiry
  async extendTombstoneExpiry(fileId, extensionDays = 30) {
    try {
      console.log("[DeleteFileManager] Extending tombstone expiry:", fileId);

      const response = await this.apiService.extendTombstoneExpiry(
        fileId,
        extensionDays,
      );

      // Update cached tombstone data
      if (response.file) {
        const normalizedFile = this.fileCryptoService.normalizeFile(
          response.file,
        );
        this.storageService.storeTombstoneData(fileId, {
          tombstone_version: normalizedFile.tombstone_version,
          tombstone_expiry: normalizedFile.tombstone_expiry,
          state: normalizedFile.state,
          version: normalizedFile.version,
          extended_at: new Date().toISOString(),
          extension_days: extensionDays,
        });
      }

      this.notifyFileDeletionListeners("tombstone_extended", {
        fileId,
        extensionDays,
        newExpiry: response.file?.tombstone_expiry,
      });

      return response;
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to extend tombstone expiry:",
        error,
      );
      throw error;
    }
  }

  // === Deletion History ===

  // Get file deletion history
  async getFileDeletionHistory(fileId, forceRefresh = false) {
    try {
      console.log("[DeleteFileManager] Getting deletion history for:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedHistory = this.storageService.getDeletionHistory(fileId);
        if (cachedHistory) {
          console.log(
            "[DeleteFileManager] Using cached deletion history:",
            fileId,
          );
          return cachedHistory.deletionData;
        }
      }

      // Fetch from API
      const response = await this.apiService.getFileDeletionHistory(fileId);

      // Store in cache
      this.storageService.storeDeletionHistory(fileId, response, {
        fetched_at: new Date().toISOString(),
      });

      console.log("[DeleteFileManager] Deletion history retrieved:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[DeleteFileManager] Failed to get deletion history:",
        error,
      );
      throw error;
    }
  }

  // === Cache Management ===

  // Clear cache for specific file
  clearFileCache(fileId) {
    this.storageService.clearFileCache(fileId);
    console.log("[DeleteFileManager] Cache cleared for file:", fileId);
  }

  // Clear all deletion caches
  clearAllCaches() {
    this.storageService.clearAllDeletionCaches();
    console.log("[DeleteFileManager] All deletion caches cleared");
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // === Event Management ===

  // Add file deletion listener
  addFileDeletionListener(callback) {
    if (typeof callback === "function") {
      this.fileDeletionListeners.add(callback);
      console.log(
        "[DeleteFileManager] File deletion listener added. Total listeners:",
        this.fileDeletionListeners.size,
      );
    }
  }

  // Remove file deletion listener
  removeFileDeletionListener(callback) {
    this.fileDeletionListeners.delete(callback);
    console.log(
      "[DeleteFileManager] File deletion listener removed. Total listeners:",
      this.fileDeletionListeners.size,
    );
  }

  // Notify file deletion listeners
  notifyFileDeletionListeners(eventType, eventData) {
    console.log(
      `[DeleteFileManager] Notifying ${this.fileDeletionListeners.size} listeners of ${eventType}`,
    );

    this.fileDeletionListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[DeleteFileManager] Error in file deletion listener:",
          error,
        );
      }
    });
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canDeleteFiles: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.fileDeletionListeners.size,
      expiredTombstones: this.getExpiredTombstones().length,
      restorableFiles: this.getRestorableFiles().length,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "DeleteFileManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      fileCryptoService: this.fileCryptoService?.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default DeleteFileManager;

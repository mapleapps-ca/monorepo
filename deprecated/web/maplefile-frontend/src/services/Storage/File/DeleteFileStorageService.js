// File: monorepo/web/maplefile-frontend/src/services/Storage/File/DeleteFileStorageService.js
// Delete File Storage Service - Handles localStorage operations for file deletion operations and tombstone management

class DeleteFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      DELETION_HISTORY: "mapleapps_file_deletion_history",
      TOMBSTONE_DATA: "mapleapps_file_tombstones",
      DELETION_OPERATIONS: "mapleapps_deletion_operations",
      BATCH_OPERATIONS: "mapleapps_batch_deletions",
    };

    // Cache configuration
    this.CACHE_DURATION = 15 * 60 * 1000; // 15 minutes for deletion operations
    this.TOMBSTONE_CACHE_DURATION = 24 * 60 * 60 * 1000; // 24 hours for tombstone data

    console.log("[DeleteFileStorageService] Storage service initialized");
  }

  // === Deletion History Storage ===

  // Store file deletion history
  storeDeletionHistory(fileId, deletionData, metadata = {}) {
    try {
      const historyEntry = {
        fileId,
        deletionData: this.sanitizeDeletionDataForStorage(deletionData),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingHistory = this.getAllDeletionHistory();
      existingHistory[fileId] = historyEntry;

      localStorage.setItem(
        this.STORAGE_KEYS.DELETION_HISTORY,
        JSON.stringify(existingHistory),
      );

      console.log(
        "[DeleteFileStorageService] Deletion history stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to store deletion history:",
        error,
      );
    }
  }

  // Get file deletion history from cache
  getDeletionHistory(fileId) {
    try {
      const allHistory = this.getAllDeletionHistory();
      const historyEntry = allHistory[fileId];

      if (!historyEntry) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(historyEntry.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[DeleteFileStorageService] Deletion history cache expired for:",
          fileId,
        );
        this.removeDeletionHistory(fileId);
        return null;
      }

      console.log(
        "[DeleteFileStorageService] Deletion history retrieved from cache:",
        fileId,
      );

      return historyEntry;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get deletion history:",
        error,
      );
      return null;
    }
  }

  // Get all stored deletion histories
  getAllDeletionHistory() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DELETION_HISTORY);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get all deletion histories:",
        error,
      );
      return {};
    }
  }

  // Remove deletion history from cache
  removeDeletionHistory(fileId) {
    try {
      const allHistory = this.getAllDeletionHistory();
      delete allHistory[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.DELETION_HISTORY,
        JSON.stringify(allHistory),
      );

      console.log(
        "[DeleteFileStorageService] Deletion history removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to remove deletion history:",
        error,
      );
      return false;
    }
  }

  // === Tombstone Data Storage ===

  // Store tombstone information
  storeTombstoneData(fileId, tombstoneData, metadata = {}) {
    try {
      const tombstoneEntry = {
        fileId,
        tombstoneData: {
          ...tombstoneData,
          cached_at: new Date().toISOString(),
        },
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(
            Date.now() + this.TOMBSTONE_CACHE_DURATION,
          ).toISOString(),
        },
      };

      const existingTombstones = this.getAllTombstoneData();
      existingTombstones[fileId] = tombstoneEntry;

      localStorage.setItem(
        this.STORAGE_KEYS.TOMBSTONE_DATA,
        JSON.stringify(existingTombstones),
      );

      console.log(
        "[DeleteFileStorageService] Tombstone data stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to store tombstone data:",
        error,
      );
    }
  }

  // Get tombstone data from cache
  getTombstoneData(fileId) {
    try {
      const allTombstones = this.getAllTombstoneData();
      const tombstoneEntry = allTombstones[fileId];

      if (!tombstoneEntry) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(tombstoneEntry.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[DeleteFileStorageService] Tombstone data cache expired for:",
          fileId,
        );
        this.removeTombstoneData(fileId);
        return null;
      }

      console.log(
        "[DeleteFileStorageService] Tombstone data retrieved from cache:",
        fileId,
      );

      return tombstoneEntry;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get tombstone data:",
        error,
      );
      return null;
    }
  }

  // Get all stored tombstone data
  getAllTombstoneData() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.TOMBSTONE_DATA);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get all tombstone data:",
        error,
      );
      return {};
    }
  }

  // Remove tombstone data from cache
  removeTombstoneData(fileId) {
    try {
      const allTombstones = this.getAllTombstoneData();
      delete allTombstones[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.TOMBSTONE_DATA,
        JSON.stringify(allTombstones),
      );

      console.log(
        "[DeleteFileStorageService] Tombstone data removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to remove tombstone data:",
        error,
      );
      return false;
    }
  }

  // === Deletion Operations Storage ===

  // Store deletion operation result
  storeDeletionOperation(operationId, operationData, metadata = {}) {
    try {
      const operationEntry = {
        operationId,
        operationData: this.sanitizeDeletionDataForStorage(operationData),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingOperations = this.getAllDeletionOperations();
      existingOperations[operationId] = operationEntry;

      localStorage.setItem(
        this.STORAGE_KEYS.DELETION_OPERATIONS,
        JSON.stringify(existingOperations),
      );

      console.log(
        "[DeleteFileStorageService] Deletion operation stored:",
        operationId,
      );
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to store deletion operation:",
        error,
      );
    }
  }

  // Get deletion operation from cache
  getDeletionOperation(operationId) {
    try {
      const allOperations = this.getAllDeletionOperations();
      const operationEntry = allOperations[operationId];

      if (!operationEntry) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(operationEntry.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[DeleteFileStorageService] Deletion operation cache expired:",
          operationId,
        );
        this.removeDeletionOperation(operationId);
        return null;
      }

      console.log(
        "[DeleteFileStorageService] Deletion operation retrieved from cache:",
        operationId,
      );

      return operationEntry;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get deletion operation:",
        error,
      );
      return null;
    }
  }

  // Get all stored deletion operations
  getAllDeletionOperations() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.DELETION_OPERATIONS,
      );
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get all deletion operations:",
        error,
      );
      return {};
    }
  }

  // Remove deletion operation from cache
  removeDeletionOperation(operationId) {
    try {
      const allOperations = this.getAllDeletionOperations();
      delete allOperations[operationId];

      localStorage.setItem(
        this.STORAGE_KEYS.DELETION_OPERATIONS,
        JSON.stringify(allOperations),
      );

      console.log(
        "[DeleteFileStorageService] Deletion operation removed:",
        operationId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to remove deletion operation:",
        error,
      );
      return false;
    }
  }

  // === Batch Operations Storage ===

  // Store batch deletion operation result
  storeBatchOperation(batchId, batchData, metadata = {}) {
    try {
      const batchEntry = {
        batchId,
        batchData: {
          ...batchData,
          results: batchData.results?.map((result) =>
            this.sanitizeDeletionDataForStorage(result),
          ),
        },
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingBatches = this.getAllBatchOperations();
      existingBatches[batchId] = batchEntry;

      localStorage.setItem(
        this.STORAGE_KEYS.BATCH_OPERATIONS,
        JSON.stringify(existingBatches),
      );

      console.log(
        "[DeleteFileStorageService] Batch operation stored:",
        batchId,
      );
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to store batch operation:",
        error,
      );
    }
  }

  // Get batch operation from cache
  getBatchOperation(batchId) {
    try {
      const allBatches = this.getAllBatchOperations();
      const batchEntry = allBatches[batchId];

      if (!batchEntry) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(batchEntry.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[DeleteFileStorageService] Batch operation cache expired:",
          batchId,
        );
        this.removeBatchOperation(batchId);
        return null;
      }

      console.log(
        "[DeleteFileStorageService] Batch operation retrieved from cache:",
        batchId,
      );

      return batchEntry;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get batch operation:",
        error,
      );
      return null;
    }
  }

  // Get all stored batch operations
  getAllBatchOperations() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.BATCH_OPERATIONS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get all batch operations:",
        error,
      );
      return {};
    }
  }

  // Remove batch operation from cache
  removeBatchOperation(batchId) {
    try {
      const allBatches = this.getAllBatchOperations();
      delete allBatches[batchId];

      localStorage.setItem(
        this.STORAGE_KEYS.BATCH_OPERATIONS,
        JSON.stringify(allBatches),
      );

      console.log(
        "[DeleteFileStorageService] Batch operation removed:",
        batchId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to remove batch operation:",
        error,
      );
      return false;
    }
  }

  // === Tombstone Management ===

  // Get expired tombstones
  getExpiredTombstones() {
    try {
      const allTombstones = this.getAllTombstoneData();
      const now = new Date();
      const expiredTombstones = [];

      Object.values(allTombstones).forEach((tombstoneEntry) => {
        const tombstoneData = tombstoneEntry.tombstoneData;
        if (tombstoneData.tombstone_expiry) {
          const expiryDate = new Date(tombstoneData.tombstone_expiry);
          if (
            expiryDate < now &&
            tombstoneData.tombstone_expiry !== "0001-01-01T00:00:00Z"
          ) {
            expiredTombstones.push(tombstoneEntry);
          }
        }
      });

      console.log(
        "[DeleteFileStorageService] Found expired tombstones:",
        expiredTombstones.length,
      );

      return expiredTombstones;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get expired tombstones:",
        error,
      );
      return [];
    }
  }

  // Get restorable files (not expired tombstones)
  getRestorableFiles() {
    try {
      const allTombstones = this.getAllTombstoneData();
      const now = new Date();
      const restorableFiles = [];

      Object.values(allTombstones).forEach((tombstoneEntry) => {
        const tombstoneData = tombstoneEntry.tombstoneData;
        if (tombstoneData.tombstone_expiry) {
          const expiryDate = new Date(tombstoneData.tombstone_expiry);
          if (
            expiryDate > now ||
            tombstoneData.tombstone_expiry === "0001-01-01T00:00:00Z"
          ) {
            restorableFiles.push(tombstoneEntry);
          }
        }
      });

      console.log(
        "[DeleteFileStorageService] Found restorable files:",
        restorableFiles.length,
      );

      return restorableFiles;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to get restorable files:",
        error,
      );
      return [];
    }
  }

  // Update file state in related caches
  updateFileStateInCaches(fileId, newState, newVersion, tombstoneData = null) {
    try {
      // Update in the main file cache (from GetFileStorageService)
      const { default: GetFileStorageService } = import(
        "./GetFileStorageService.js"
      )
        .then((module) => {
          const getFileStorage = module.default;
          if (getFileStorage) {
            const fileDetails = getFileStorage.getFileDetails(fileId);
            if (fileDetails) {
              const updatedFile = {
                ...fileDetails.fileDetails,
                state: newState,
                version: newVersion,
                ...(tombstoneData && {
                  tombstone_version: tombstoneData.tombstone_version,
                  tombstone_expiry: tombstoneData.tombstone_expiry,
                }),
                updated_at: new Date().toISOString(),
              };
              getFileStorage.updateFileDetails(fileId, updatedFile);
            }
          }
        })
        .catch((error) => {
          console.warn(
            "[DeleteFileStorageService] Could not update file cache:",
            error,
          );
        });

      // Store tombstone data if provided
      if (tombstoneData) {
        this.storeTombstoneData(fileId, tombstoneData, {
          operation: "state_update",
          new_state: newState,
          new_version: newVersion,
        });
      }

      console.log("[DeleteFileStorageService] File state updated in caches:", {
        fileId,
        newState,
        newVersion,
        hasTombstone: !!tombstoneData,
      });
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to update file state in caches:",
        error,
      );
    }
  }

  // === Cache Management ===

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired deletion history
      const allHistory = this.getAllDeletionHistory();
      Object.keys(allHistory).forEach((fileId) => {
        const history = allHistory[fileId];
        const expiresAt = new Date(history.metadata.expires_at);
        if (now > expiresAt) {
          delete allHistory[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.DELETION_HISTORY,
          JSON.stringify(allHistory),
        );
      }

      // Clear expired tombstone data
      const allTombstones = this.getAllTombstoneData();
      Object.keys(allTombstones).forEach((fileId) => {
        const tombstone = allTombstones[fileId];
        const expiresAt = new Date(tombstone.metadata.expires_at);
        if (now > expiresAt) {
          delete allTombstones[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.TOMBSTONE_DATA,
          JSON.stringify(allTombstones),
        );
      }

      // Clear expired deletion operations
      const allOperations = this.getAllDeletionOperations();
      Object.keys(allOperations).forEach((operationId) => {
        const operation = allOperations[operationId];
        const expiresAt = new Date(operation.metadata.expires_at);
        if (now > expiresAt) {
          delete allOperations[operationId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.DELETION_OPERATIONS,
          JSON.stringify(allOperations),
        );
      }

      // Clear expired batch operations
      const allBatches = this.getAllBatchOperations();
      Object.keys(allBatches).forEach((batchId) => {
        const batch = allBatches[batchId];
        const expiresAt = new Date(batch.metadata.expires_at);
        if (now > expiresAt) {
          delete allBatches[batchId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.BATCH_OPERATIONS,
          JSON.stringify(allBatches),
        );
      }

      console.log(
        "[DeleteFileStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all deletion-related caches
  clearAllDeletionCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DELETION_HISTORY);
      localStorage.removeItem(this.STORAGE_KEYS.TOMBSTONE_DATA);
      localStorage.removeItem(this.STORAGE_KEYS.DELETION_OPERATIONS);
      localStorage.removeItem(this.STORAGE_KEYS.BATCH_OPERATIONS);
      console.log("[DeleteFileStorageService] All deletion caches cleared");
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to clear deletion caches:",
        error,
      );
    }
  }

  // Clear cache for specific file
  clearFileCache(fileId) {
    try {
      this.removeDeletionHistory(fileId);
      this.removeTombstoneData(fileId);

      console.log("[DeleteFileStorageService] Cache cleared for file:", fileId);
    } catch (error) {
      console.error(
        "[DeleteFileStorageService] Failed to clear file cache:",
        error,
      );
    }
  }

  // === Data Sanitization ===

  // Sanitize deletion data for storage (remove sensitive data)
  sanitizeDeletionDataForStorage(deletionData) {
    const sanitized = { ...deletionData };

    // Remove sensitive data that shouldn't be stored
    delete sanitized._file_key; // Never store file keys
    delete sanitized._collection_key; // Never store collection keys
    delete sanitized._decrypted_content; // Never store decrypted content

    return sanitized;
  }

  // === Configuration ===

  // Set cache duration
  setCacheDuration(duration) {
    this.CACHE_DURATION = duration;
    console.log(
      "[DeleteFileStorageService] Cache duration set to:",
      duration,
      "ms",
    );
  }

  // Set tombstone cache duration
  setTombstoneCacheDuration(duration) {
    this.TOMBSTONE_CACHE_DURATION = duration;
    console.log(
      "[DeleteFileStorageService] Tombstone cache duration set to:",
      duration,
      "ms",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const allHistory = this.getAllDeletionHistory();
    const allTombstones = this.getAllTombstoneData();
    const allOperations = this.getAllDeletionOperations();
    const allBatches = this.getAllBatchOperations();

    return {
      deletionHistoryCount: Object.keys(allHistory).length,
      tombstoneDataCount: Object.keys(allTombstones).length,
      deletionOperationCount: Object.keys(allOperations).length,
      batchOperationCount: Object.keys(allBatches).length,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      cacheDuration: this.CACHE_DURATION,
      tombstoneCacheDuration: this.TOMBSTONE_CACHE_DURATION,
    };
  }

  // Get debug information
  getDebugInfo() {
    const allHistory = this.getAllDeletionHistory();
    const allTombstones = this.getAllTombstoneData();
    const allOperations = this.getAllDeletionOperations();

    return {
      serviceName: "DeleteFileStorageService",
      storageInfo: this.getStorageInfo(),
      cachedDeletionHistoryIds: Object.keys(allHistory),
      cachedTombstoneIds: Object.keys(allTombstones),
      recentDeletions: Object.values(allHistory)
        .sort(
          (a, b) =>
            new Date(b.metadata.cached_at) - new Date(a.metadata.cached_at),
        )
        .slice(0, 5)
        .map((history) => ({
          fileId: history.fileId,
          cached_at: history.metadata.cached_at,
        })),
      expiredTombstonesCount: this.getExpiredTombstones().length,
      restorableFilesCount: this.getRestorableFiles().length,
    };
  }
}

export default DeleteFileStorageService;

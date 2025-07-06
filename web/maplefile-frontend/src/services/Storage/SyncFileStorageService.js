// File: monorepo/web/maplefile-frontend/src/services/Storage/SyncFileStorageService.js
// Simple service to save/get sync files from localStorage

class SyncFileStorageService {
  constructor() {
    this.STORAGE_KEY = "maplefile_sync_files";
    this.METADATA_KEY = "maplefile_sync_files_metadata";
    console.log("[SyncFileStorageService] Service initialized");
  }

  /**
   * Save the entire sync file array to localStorage
   * @param {Array} syncFiles - Array of sync file objects
   * @returns {boolean} - Success status
   */
  saveSyncFiles(syncFiles) {
    try {
      if (!Array.isArray(syncFiles)) {
        throw new Error("Sync files must be an array");
      }

      console.log(
        `[SyncFileStorageService] Saving ${syncFiles.length} sync files to localStorage`,
      );

      // Save the sync files
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(syncFiles));

      // Save metadata
      const metadata = {
        count: syncFiles.length,
        savedAt: new Date().toISOString(),
        version: 1,
        collectionBreakdown: this.getCollectionBreakdown(syncFiles),
      };
      localStorage.setItem(this.METADATA_KEY, JSON.stringify(metadata));

      console.log(
        `[SyncFileStorageService] ✅ Successfully saved ${syncFiles.length} sync files`,
      );
      return true;
    } catch (error) {
      console.error(
        "[SyncFileStorageService] ❌ Failed to save sync files:",
        error,
      );
      return false;
    }
  }

  /**
   * Get the entire sync file array from localStorage
   * @returns {Array} - Array of sync file objects (empty array if none found)
   */
  getSyncFiles() {
    try {
      console.log(
        "[SyncFileStorageService] Loading sync files from localStorage",
      );

      const storedData = localStorage.getItem(this.STORAGE_KEY);

      if (!storedData) {
        console.log(
          "[SyncFileStorageService] No sync files found in localStorage",
        );
        return [];
      }

      const syncFiles = JSON.parse(storedData);

      if (!Array.isArray(syncFiles)) {
        console.warn(
          "[SyncFileStorageService] Invalid data format in localStorage, returning empty array",
        );
        return [];
      }

      console.log(
        `[SyncFileStorageService] ✅ Successfully loaded ${syncFiles.length} sync files`,
      );
      return syncFiles;
    } catch (error) {
      console.error(
        "[SyncFileStorageService] ❌ Failed to load sync files:",
        error,
      );
      return [];
    }
  }

  /**
   * Get sync files by collection ID
   * @param {string} collectionId - Collection ID to filter by
   * @returns {Array} - Array of sync files for the collection
   */
  getSyncFilesByCollection(collectionId) {
    if (!collectionId) {
      console.warn("[SyncFileStorageService] No collection ID provided");
      return [];
    }

    const allFiles = this.getSyncFiles();
    return allFiles.filter((file) => file.collection_id === collectionId);
  }

  /**
   * Get a single sync file by ID
   * @param {string} fileId - File ID
   * @returns {Object|null} - Sync file object or null if not found
   */
  getSyncFileById(fileId) {
    if (!fileId) {
      console.warn("[SyncFileStorageService] No file ID provided");
      return null;
    }

    const allFiles = this.getSyncFiles();
    return allFiles.find((file) => file.id === fileId) || null;
  }

  /**
   * Get metadata about stored sync files
   * @returns {Object|null} - Metadata object or null if no data
   */
  getMetadata() {
    try {
      const metadataStr = localStorage.getItem(this.METADATA_KEY);
      if (!metadataStr) return null;

      return JSON.parse(metadataStr);
    } catch (error) {
      console.error("[SyncFileStorageService] Failed to load metadata:", error);
      return null;
    }
  }

  /**
   * Get collection breakdown of sync files
   * @param {Array} syncFiles - Array of sync files
   * @returns {Object} - Breakdown by collection
   */
  getCollectionBreakdown(syncFiles) {
    const breakdown = {};

    syncFiles.forEach((file) => {
      const collectionId = file.collection_id || "no_collection";
      if (!breakdown[collectionId]) {
        breakdown[collectionId] = {
          count: 0,
          totalSize: 0,
          states: { active: 0, deleted: 0, archived: 0 },
        };
      }

      breakdown[collectionId].count++;
      breakdown[collectionId].totalSize += file.file_size || 0;

      if (
        file.state &&
        breakdown[collectionId].states[file.state] !== undefined
      ) {
        breakdown[collectionId].states[file.state]++;
      }
    });

    return breakdown;
  }

  /**
   * Check if sync files are stored in localStorage
   * @returns {boolean} - True if sync files exist
   */
  hasStoredSyncFiles() {
    const syncFiles = this.getSyncFiles();
    return syncFiles.length > 0;
  }

  /**
   * Clear all stored sync files
   * @returns {boolean} - Success status
   */
  clearSyncFiles() {
    try {
      localStorage.removeItem(this.STORAGE_KEY);
      localStorage.removeItem(this.METADATA_KEY);
      console.log(
        "[SyncFileStorageService] ✅ Sync files cleared from localStorage",
      );
      return true;
    } catch (error) {
      console.error(
        "[SyncFileStorageService] ❌ Failed to clear sync files:",
        error,
      );
      return false;
    }
  }

  /**
   * Get storage info for debugging
   * @returns {Object} - Storage information
   */
  getStorageInfo() {
    const syncFiles = this.getSyncFiles();
    const metadata = this.getMetadata();

    return {
      hasSyncFiles: syncFiles.length > 0,
      syncFilesCount: syncFiles.length,
      metadata,
      storageKeys: {
        syncFiles: !!localStorage.getItem(this.STORAGE_KEY),
        metadata: !!localStorage.getItem(this.METADATA_KEY),
      },
      collectionBreakdown:
        metadata?.collectionBreakdown || this.getCollectionBreakdown(syncFiles),
    };
  }
}

export default SyncFileStorageService;

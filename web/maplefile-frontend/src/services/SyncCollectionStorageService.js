// SyncCollectionStorageService.js
// Simple service to save/get sync collections from localStorage

class SyncCollectionStorageService {
  constructor() {
    this.STORAGE_KEY = "maplefile_sync_collections";
    this.METADATA_KEY = "maplefile_sync_collections_metadata";
    console.log("[SyncCollectionStorageService] Service initialized");
  }

  /**
   * Save the entire sync collection array to localStorage
   * @param {Array} syncCollections - Array of sync collection objects
   * @returns {boolean} - Success status
   */
  saveSyncCollections(syncCollections) {
    try {
      if (!Array.isArray(syncCollections)) {
        throw new Error("Sync collections must be an array");
      }

      console.log(
        `[SyncCollectionStorageService] Saving ${syncCollections.length} sync collections to localStorage`,
      );

      // Save the sync collections
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(syncCollections));

      // Save metadata
      const metadata = {
        count: syncCollections.length,
        savedAt: new Date().toISOString(),
        version: 1,
      };
      localStorage.setItem(this.METADATA_KEY, JSON.stringify(metadata));

      console.log(
        `[SyncCollectionStorageService] ✅ Successfully saved ${syncCollections.length} sync collections`,
      );
      return true;
    } catch (error) {
      console.error(
        "[SyncCollectionStorageService] ❌ Failed to save sync collections:",
        error,
      );
      return false;
    }
  }

  /**
   * Get the entire sync collection array from localStorage
   * @returns {Array} - Array of sync collection objects (empty array if none found)
   */
  getSyncCollections() {
    try {
      console.log(
        "[SyncCollectionStorageService] Loading sync collections from localStorage",
      );

      const storedData = localStorage.getItem(this.STORAGE_KEY);

      if (!storedData) {
        console.log(
          "[SyncCollectionStorageService] No sync collections found in localStorage",
        );
        return [];
      }

      const syncCollections = JSON.parse(storedData);

      if (!Array.isArray(syncCollections)) {
        console.warn(
          "[SyncCollectionStorageService] Invalid data format in localStorage, returning empty array",
        );
        return [];
      }

      console.log(
        `[SyncCollectionStorageService] ✅ Successfully loaded ${syncCollections.length} sync collections`,
      );
      return syncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionStorageService] ❌ Failed to load sync collections:",
        error,
      );
      return [];
    }
  }

  /**
   * Get metadata about stored sync collections
   * @returns {Object|null} - Metadata object or null if no data
   */
  getMetadata() {
    try {
      const metadataStr = localStorage.getItem(this.METADATA_KEY);
      if (!metadataStr) return null;

      return JSON.parse(metadataStr);
    } catch (error) {
      console.error(
        "[SyncCollectionStorageService] Failed to load metadata:",
        error,
      );
      return null;
    }
  }

  /**
   * Check if sync collections are stored in localStorage
   * @returns {boolean} - True if sync collections exist
   */
  hasStoredSyncCollections() {
    const syncCollections = this.getSyncCollections();
    return syncCollections.length > 0;
  }

  /**
   * Clear all stored sync collections
   * @returns {boolean} - Success status
   */
  clearSyncCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEY);
      localStorage.removeItem(this.METADATA_KEY);
      console.log(
        "[SyncCollectionStorageService] ✅ Sync collections cleared from localStorage",
      );
      return true;
    } catch (error) {
      console.error(
        "[SyncCollectionStorageService] ❌ Failed to clear sync collections:",
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
    const syncCollections = this.getSyncCollections();
    const metadata = this.getMetadata();

    return {
      hasSyncCollections: syncCollections.length > 0,
      syncCollectionsCount: syncCollections.length,
      metadata,
      storageKeys: {
        syncCollections: !!localStorage.getItem(this.STORAGE_KEY),
        metadata: !!localStorage.getItem(this.METADATA_KEY),
      },
    };
  }
}

export default SyncCollectionStorageService;

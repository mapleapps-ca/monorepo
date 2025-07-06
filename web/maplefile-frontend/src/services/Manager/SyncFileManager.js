// File: monorepo/web/maplefile-frontend/src/services/Manager/SyncFileManager.js
// Sync File Manager - Orchestrates API and Storage services for sync file management

import SyncFileAPIService from "../API/SyncFileAPIService.js";
import SyncFileStorageService from "../Storage/SyncFileStorageService.js";

class SyncFileManager {
  constructor(authManager) {
    // SyncFileManager depends on AuthManager and orchestrates API and Storage services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new SyncFileAPIService(authManager);
    this.storageService = new SyncFileStorageService();

    console.log(
      "[SyncFileManager] Sync file manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[SyncFileManager] Initializing sync file manager...");
      // Any initialization logic for storage service if needed
      console.log(
        "[SyncFileManager] Sync file manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[SyncFileManager] Failed to initialize sync file manager:",
        error,
      );
    }
  }

  // === Sync File Management ===

  /**
   * Get sync files with smart caching logic
   * @param {Object} options - Options for retrieval
   * @param {boolean} options.forceRefresh - Force refresh from API
   * @param {boolean} options.useCache - Use cached data if available (default: true)
   * @param {string} options.collectionId - Filter by collection ID
   * @returns {Promise<Array>} - Array of sync files
   */
  async getSyncFiles(options = {}) {
    const {
      forceRefresh = false,
      useCache = true,
      collectionId = null,
    } = options;

    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log("[SyncFileManager] Orchestrating sync file retrieval");

      // If force refresh is requested, skip cache
      if (forceRefresh) {
        console.log(
          "[SyncFileManager] Force refresh requested, fetching from API",
        );
        return await this.refreshSyncFiles(options);
      }

      // Try to get sync files from local storage first
      if (useCache) {
        let cachedFiles;

        if (collectionId) {
          cachedFiles =
            this.storageService.getSyncFilesByCollection(collectionId);
        } else {
          cachedFiles = this.storageService.getSyncFiles();
        }

        if (cachedFiles && cachedFiles.length > 0) {
          console.log(
            `[SyncFileManager] Retrieved ${cachedFiles.length} sync files from cache`,
          );
          return cachedFiles;
        }
      }

      // If not in local storage or cache disabled, fetch from API and save
      console.log("[SyncFileManager] No cached data found, fetching from API");
      return await this.refreshSyncFiles(options);
    } catch (error) {
      console.error("[SyncFileManager] Error getting sync files:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Get sync files for a specific collection
   * @param {string} collectionId - Collection ID
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of sync files
   */
  async getSyncFilesByCollection(collectionId, options = {}) {
    return await this.getSyncFiles({ ...options, collectionId });
  }

  /**
   * Get a single sync file by ID
   * @param {string} fileId - File ID
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Object>} - Sync file object
   */
  async getSyncFile(fileId, options = {}) {
    const { forceRefresh = false, useCache = true } = options;

    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    if (!fileId) {
      throw new Error("File ID is required");
    }

    try {
      // Try cache first unless force refresh
      if (!forceRefresh && useCache) {
        const cachedFile = this.storageService.getSyncFileById(fileId);
        if (cachedFile) {
          console.log(`[SyncFileManager] Found sync file ${fileId} in cache`);
          return cachedFile;
        }
      }

      // Fetch from API
      console.log(`[SyncFileManager] Fetching sync file ${fileId} from API`);
      const syncFile = await this.apiService.getSyncFile(fileId);

      // Update cache with this file
      const allFiles = this.storageService.getSyncFiles();
      const index = allFiles.findIndex((f) => f.id === fileId);
      if (index >= 0) {
        allFiles[index] = syncFile;
      } else {
        allFiles.push(syncFile);
      }
      this.storageService.saveSyncFiles(allFiles);

      return syncFile;
    } catch (error) {
      console.error(
        `[SyncFileManager] Error getting sync file ${fileId}:`,
        error,
      );
      throw error;
    }
  }

  /**
   * Refresh sync files from API and update cache
   * @param {Object} options - Options for refresh
   * @returns {Promise<Array>} - Array of sync files
   */
  async refreshSyncFiles(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log("[SyncFileManager] Orchestrating sync file refresh from API");

      // 1. Fetch sync files from API via the API service
      const syncFiles = await this.apiService.syncAllFiles(options);

      // 2. Save sync files to local storage via the storage service
      const saveSuccess = this.storageService.saveSyncFiles(syncFiles);
      if (!saveSuccess) {
        console.warn("[SyncFileManager] Failed to save sync files to storage");
      }

      console.log(
        `[SyncFileManager] Refreshed and cached ${syncFiles.length} sync files from API`,
      );
      return syncFiles;
    } catch (error) {
      console.error("[SyncFileManager] Error refreshing sync files:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Force refresh sync files (alias for refreshSyncFiles with force option)
   * @param {Object} options - Options for refresh
   * @returns {Promise<Array>} - Array of sync files
   */
  async forceRefreshSyncFiles(options = {}) {
    console.log("[SyncFileManager] Force refresh sync files requested");
    return await this.refreshSyncFiles(options);
  }

  // === Storage Management ===

  /**
   * Get sync files from storage only (no API call)
   * @returns {Array} - Array of sync files from storage
   */
  getSyncFilesFromStorage() {
    console.log("[SyncFileManager] Getting sync files from storage only");
    return this.storageService.getSyncFiles();
  }

  /**
   * Save sync files to storage
   * @param {Array} syncFiles - Array of sync files to save
   * @returns {boolean} - Success status
   */
  saveSyncFilesToStorage(syncFiles) {
    console.log(
      `[SyncFileManager] Saving ${syncFiles.length} sync files to storage`,
    );
    return this.storageService.saveSyncFiles(syncFiles);
  }

  /**
   * Clear all stored sync files
   * @returns {boolean} - Success status
   */
  clearSyncFiles() {
    console.log("[SyncFileManager] Clearing all sync files from storage");
    return this.storageService.clearSyncFiles();
  }

  /**
   * Check if sync files are stored locally
   * @returns {boolean} - True if sync files exist in storage
   */
  hasSyncFiles() {
    return this.storageService.hasStoredSyncFiles();
  }

  // === API Management ===

  /**
   * Sync sync files from API only (no storage interaction)
   * @param {Object} options - Options for API sync
   * @returns {Promise<Array>} - Array of sync files from API
   */
  async syncFromAPI(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log("[SyncFileManager] Syncing sync files from API only");
      return await this.apiService.syncAllFiles(options);
    } catch (error) {
      console.error("[SyncFileManager] Error syncing from API:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Filtering and Utilities ===

  /**
   * Filter sync files by state
   * @param {Array} syncFiles - Array of sync files
   * @param {string} state - State to filter by ('active', 'deleted', 'archived')
   * @returns {Array} - Filtered sync files
   */
  filterSyncFilesByState(syncFiles, state) {
    if (!Array.isArray(syncFiles)) {
      console.warn(
        "[SyncFileManager] filterSyncFilesByState: syncFiles is not an array",
      );
      return [];
    }

    return syncFiles.filter((file) => file.state === state);
  }

  /**
   * Filter sync files by collection
   * @param {Array} syncFiles - Array of sync files
   * @param {string} collectionId - Collection ID to filter by
   * @returns {Array} - Filtered sync files
   */
  filterSyncFilesByCollection(syncFiles, collectionId) {
    if (!Array.isArray(syncFiles)) {
      console.warn(
        "[SyncFileManager] filterSyncFilesByCollection: syncFiles is not an array",
      );
      return [];
    }

    return syncFiles.filter((file) => file.collection_id === collectionId);
  }

  /**
   * Get sync files grouped by state
   * @param {Array} syncFiles - Array of sync files
   * @returns {Object} - Object with files grouped by state
   */
  groupSyncFilesByState(syncFiles) {
    if (!Array.isArray(syncFiles)) {
      console.warn(
        "[SyncFileManager] groupSyncFilesByState: syncFiles is not an array",
      );
      return { active: [], deleted: [], archived: [] };
    }

    return {
      active: this.filterSyncFilesByState(syncFiles, "active"),
      deleted: this.filterSyncFilesByState(syncFiles, "deleted"),
      archived: this.filterSyncFilesByState(syncFiles, "archived"),
    };
  }

  /**
   * Get sync file statistics
   * @param {Array} syncFiles - Array of sync files
   * @returns {Object} - Statistics object
   */
  getSyncFileStats(syncFiles) {
    if (!Array.isArray(syncFiles)) {
      return {
        total: 0,
        active: 0,
        deleted: 0,
        archived: 0,
        totalSize: 0,
        collections: 0,
      };
    }

    const grouped = this.groupSyncFilesByState(syncFiles);
    const collections = new Set(syncFiles.map((f) => f.collection_id)).size;
    const totalSize = syncFiles.reduce(
      (sum, file) => sum + (file.file_size || 0),
      0,
    );

    return {
      total: syncFiles.length,
      active: grouped.active.length,
      deleted: grouped.deleted.length,
      archived: grouped.archived.length,
      totalSize,
      collections,
    };
  }

  // === State Management ===

  /**
   * Get loading state
   * @returns {boolean} - True if currently loading
   */
  getIsLoading() {
    return this.isLoading;
  }

  /**
   * Check if API service is loading
   * @returns {boolean} - True if API service is loading
   */
  isAPILoading() {
    return this.apiService.isLoadingSync();
  }

  // === Storage Information ===

  /**
   * Get storage information
   * @returns {Object} - Storage information
   */
  getStorageInfo() {
    return this.storageService.getStorageInfo();
  }

  /**
   * Get storage metadata
   * @returns {Object|null} - Metadata object or null
   */
  getStorageMetadata() {
    return this.storageService.getMetadata();
  }

  // === Manager Status ===

  /**
   * Get manager status and information
   * @returns {Object} - Manager status
   */
  getManagerStatus() {
    const storageInfo = this.getStorageInfo();
    const stats = this.getSyncFileStats(this.getSyncFilesFromStorage());

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      isAPILoading: this.isAPILoading(),
      canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      statistics: stats,
      lastSync: storageInfo.metadata?.savedAt || null,
    };
  }

  // === Debug Information ===

  /**
   * Get comprehensive debug information
   * @returns {Object} - Debug information
   */
  getDebugInfo() {
    return {
      serviceName: "SyncFileManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getStorageInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }

  // === Convenience Methods ===

  /**
   * Get active sync files
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of active sync files
   */
  async getActiveSyncFiles(options = {}) {
    const allFiles = await this.getSyncFiles(options);
    return this.filterSyncFilesByState(allFiles, "active");
  }

  /**
   * Get deleted sync files
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of deleted sync files
   */
  async getDeletedSyncFiles(options = {}) {
    const allFiles = await this.getSyncFiles(options);
    return this.filterSyncFilesByState(allFiles, "deleted");
  }

  /**
   * Get archived sync files
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of archived sync files
   */
  async getArchivedSyncFiles(options = {}) {
    const allFiles = await this.getSyncFiles(options);
    return this.filterSyncFilesByState(allFiles, "archived");
  }

  /**
   * Sync and save sync files (convenience method)
   * @param {Object} options - Options for sync
   * @returns {Promise<Array>} - Array of sync files
   */
  async syncAndSave(options = {}) {
    console.log("[SyncFileManager] Sync and save sync files");
    return await this.refreshSyncFiles(options);
  }

  /**
   * Get file size statistics
   * @param {Array} syncFiles - Array of sync files
   * @returns {Object} - Size statistics
   */
  getFileSizeStats(syncFiles) {
    if (!Array.isArray(syncFiles) || syncFiles.length === 0) {
      return {
        totalSize: 0,
        averageSize: 0,
        largestFile: null,
        smallestFile: null,
      };
    }

    const sizes = syncFiles.map((f) => f.file_size || 0);
    const totalSize = sizes.reduce((sum, size) => sum + size, 0);
    const averageSize = totalSize / syncFiles.length;

    const sorted = [...syncFiles].sort(
      (a, b) => (b.file_size || 0) - (a.file_size || 0),
    );

    return {
      totalSize,
      averageSize,
      largestFile: sorted[0] || null,
      smallestFile: sorted[sorted.length - 1] || null,
    };
  }
}

export default SyncFileManager;

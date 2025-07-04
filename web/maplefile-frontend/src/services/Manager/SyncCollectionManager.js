// File: monorepo/web/maplefile-frontend/src/services/Manager/SyncCollectionManager.js
// Sync Collection Manager - Orchestrates API and Storage services for sync collection management

import SyncCollectionAPIService from "../API/SyncCollectionAPIService.js";
import SyncCollectionStorageService from "../Storage/SyncCollectionStorageService.js";

class SyncCollectionManager {
  constructor(authManager) {
    // SyncCollectionManager depends on AuthManager and orchestrates API and Storage services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new SyncCollectionAPIService(authManager);
    this.storageService = new SyncCollectionStorageService();

    console.log(
      "[SyncCollectionManager] Sync collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[SyncCollectionManager] Initializing sync collection manager...",
      );
      // Any initialization logic for storage service if needed
      console.log(
        "[SyncCollectionManager] Sync collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[SyncCollectionManager] Failed to initialize sync collection manager:",
        error,
      );
    }
  }

  // === Sync Collection Management ===

  /**
   * Get sync collections with smart caching logic
   * @param {Object} options - Options for retrieval
   * @param {boolean} options.forceRefresh - Force refresh from API
   * @param {boolean} options.useCache - Use cached data if available (default: true)
   * @returns {Promise<Array>} - Array of sync collections
   */
  async getSyncCollections(options = {}) {
    const { forceRefresh = false, useCache = true } = options;

    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncCollectionManager] Orchestrating sync collection retrieval",
      );

      // If force refresh is requested, skip cache
      if (forceRefresh) {
        console.log(
          "[SyncCollectionManager] Force refresh requested, fetching from API",
        );
        return await this.refreshSyncCollections();
      }

      // Try to get sync collections from local storage first
      if (useCache) {
        const cachedCollections = this.storageService.getSyncCollections();
        if (cachedCollections && cachedCollections.length > 0) {
          console.log(
            `[SyncCollectionManager] Retrieved ${cachedCollections.length} sync collections from cache`,
          );
          return cachedCollections;
        }
      }

      // If not in local storage or cache disabled, fetch from API and save
      console.log(
        "[SyncCollectionManager] No cached data found, fetching from API",
      );
      return await this.refreshSyncCollections();
    } catch (error) {
      console.error(
        "[SyncCollectionManager] Error getting sync collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Refresh sync collections from API and update cache
   * @param {Object} options - Options for refresh
   * @returns {Promise<Array>} - Array of sync collections
   */
  async refreshSyncCollections(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncCollectionManager] Orchestrating sync collection refresh from API",
      );

      // 1. Fetch sync collections from API via the API service
      const syncCollections = await this.apiService.syncAllCollections(options);

      // 2. Save sync collections to local storage via the storage service
      const saveSuccess =
        this.storageService.saveSyncCollections(syncCollections);
      if (!saveSuccess) {
        console.warn(
          "[SyncCollectionManager] Failed to save sync collections to storage",
        );
      }

      console.log(
        `[SyncCollectionManager] Refreshed and cached ${syncCollections.length} sync collections from API`,
      );
      return syncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionManager] Error refreshing sync collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Force refresh sync collections (alias for refreshSyncCollections with force option)
   * @param {Object} options - Options for refresh
   * @returns {Promise<Array>} - Array of sync collections
   */
  async forceRefreshSyncCollections(options = {}) {
    console.log(
      "[SyncCollectionManager] Force refresh sync collections requested",
    );
    return await this.refreshSyncCollections(options);
  }

  // === Storage Management ===

  /**
   * Get sync collections from storage only (no API call)
   * @returns {Array} - Array of sync collections from storage
   */
  getSyncCollectionsFromStorage() {
    console.log(
      "[SyncCollectionManager] Getting sync collections from storage only",
    );
    return this.storageService.getSyncCollections();
  }

  /**
   * Save sync collections to storage
   * @param {Array} syncCollections - Array of sync collections to save
   * @returns {boolean} - Success status
   */
  saveSyncCollectionsToStorage(syncCollections) {
    console.log(
      `[SyncCollectionManager] Saving ${syncCollections.length} sync collections to storage`,
    );
    return this.storageService.saveSyncCollections(syncCollections);
  }

  /**
   * Clear all stored sync collections
   * @returns {boolean} - Success status
   */
  clearSyncCollections() {
    console.log(
      "[SyncCollectionManager] Clearing all sync collections from storage",
    );
    return this.storageService.clearSyncCollections();
  }

  /**
   * Check if sync collections are stored locally
   * @returns {boolean} - True if sync collections exist in storage
   */
  hasSyncCollections() {
    return this.storageService.hasStoredSyncCollections();
  }

  // === API Management ===

  /**
   * Sync sync collections from API only (no storage interaction)
   * @param {Object} options - Options for API sync
   * @returns {Promise<Array>} - Array of sync collections from API
   */
  async syncFromAPI(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncCollectionManager] Syncing sync collections from API only",
      );
      return await this.apiService.syncAllCollections(options);
    } catch (error) {
      console.error("[SyncCollectionManager] Error syncing from API:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Filtering and Utilities ===

  /**
   * Filter sync collections by state
   * @param {Array} syncCollections - Array of sync collections
   * @param {string} state - State to filter by ('active', 'deleted', 'archived')
   * @returns {Array} - Filtered sync collections
   */
  filterSyncCollectionsByState(syncCollections, state) {
    if (!Array.isArray(syncCollections)) {
      console.warn(
        "[SyncCollectionManager] filterSyncCollectionsByState: syncCollections is not an array",
      );
      return [];
    }

    return syncCollections.filter((collection) => collection.state === state);
  }

  /**
   * Get sync collections grouped by state
   * @param {Array} syncCollections - Array of sync collections
   * @returns {Object} - Object with collections grouped by state
   */
  groupSyncCollectionsByState(syncCollections) {
    if (!Array.isArray(syncCollections)) {
      console.warn(
        "[SyncCollectionManager] groupSyncCollectionsByState: syncCollections is not an array",
      );
      return { active: [], deleted: [], archived: [] };
    }

    return {
      active: this.filterSyncCollectionsByState(syncCollections, "active"),
      deleted: this.filterSyncCollectionsByState(syncCollections, "deleted"),
      archived: this.filterSyncCollectionsByState(syncCollections, "archived"),
    };
  }

  /**
   * Get sync collection statistics
   * @param {Array} syncCollections - Array of sync collections
   * @returns {Object} - Statistics object
   */
  getSyncCollectionStats(syncCollections) {
    if (!Array.isArray(syncCollections)) {
      return { total: 0, active: 0, deleted: 0, archived: 0 };
    }

    const grouped = this.groupSyncCollectionsByState(syncCollections);
    return {
      total: syncCollections.length,
      active: grouped.active.length,
      deleted: grouped.deleted.length,
      archived: grouped.archived.length,
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
    const stats = this.getSyncCollectionStats(
      this.getSyncCollectionsFromStorage(),
    );

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
      serviceName: "SyncCollectionManager",
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
   * Get active sync collections
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of active sync collections
   */
  async getActiveSyncCollections(options = {}) {
    const allCollections = await this.getSyncCollections(options);
    return this.filterSyncCollectionsByState(allCollections, "active");
  }

  /**
   * Get deleted sync collections
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of deleted sync collections
   */
  async getDeletedSyncCollections(options = {}) {
    const allCollections = await this.getSyncCollections(options);
    return this.filterSyncCollectionsByState(allCollections, "deleted");
  }

  /**
   * Get archived sync collections
   * @param {Object} options - Options for retrieval
   * @returns {Promise<Array>} - Array of archived sync collections
   */
  async getArchivedSyncCollections(options = {}) {
    const allCollections = await this.getSyncCollections(options);
    return this.filterSyncCollectionsByState(allCollections, "archived");
  }

  /**
   * Sync and save sync collections (convenience method)
   * @param {Object} options - Options for sync
   * @returns {Promise<Array>} - Array of sync collections
   */
  async syncAndSave(options = {}) {
    console.log("[SyncCollectionManager] Sync and save sync collections");
    return await this.refreshSyncCollections(options);
  }
}

export default SyncCollectionManager;

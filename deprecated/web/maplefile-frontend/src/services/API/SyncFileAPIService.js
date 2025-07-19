// File: monorepo/web/maplefile-frontend/src/services/API/SyncFileAPIService.js
// Service for making API calls to sync files (paginated requests)

class SyncFileAPIService {
  constructor(authManager) {
    this.authManager = authManager;
    this.isLoading = false;
    this._apiClient = null;
    console.log(
      "[SyncFileAPIService] API service initialized with AuthManager",
    );
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("./ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  /**
   * Sync ALL files by automatically paginating through all API pages
   * This is the ONLY function you need to call!
   * @param {Object} options - Optional settings
   * @param {number} options.pageSize - Files per page (default: 1000, max: 5000)
   * @param {string} options.collectionId - Filter by collection ID (optional)
   * @returns {Promise<Array>} - Complete array of all sync files from API
   */
  async syncAllFiles(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncFileAPIService] Starting complete API sync via AuthManager...",
      );

      const allSyncFiles = [];
      let cursor = null;
      let hasMore = true;
      let pageCount = 0;
      const pageSize = Math.min(options.pageSize || 1000, 5000);
      const collectionId = options.collectionId || null;

      while (hasMore) {
        pageCount++;
        console.log(
          `[SyncFileAPIService] Fetching API page ${pageCount} via AuthManager...`,
        );

        // Make single page request
        const response = await this.fetchSinglePage(
          pageSize,
          cursor,
          collectionId,
        );

        // Add sync files to the result
        if (response.files && response.files.length > 0) {
          allSyncFiles.push(...response.files);
          console.log(
            `[SyncFileAPIService] Page ${pageCount} via AuthManager: ${response.files.length} sync files (Total: ${allSyncFiles.length})`,
          );
        }

        // Check if there are more pages
        hasMore = response.has_more;
        cursor = response.next_cursor;
      }

      console.log(
        `[SyncFileAPIService] ✅ API sync complete via AuthManager! ${allSyncFiles.length} sync files across ${pageCount} pages`,
      );
      return allSyncFiles;
    } catch (error) {
      console.error(
        "[SyncFileAPIService] ❌ API sync failed via AuthManager:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Get sync files for a specific collection
   * @param {string} collectionId - Collection ID to filter by
   * @param {Object} options - Optional settings
   * @returns {Promise<Array>} - Array of sync files for the collection
   */
  async syncFilesByCollection(collectionId, options = {}) {
    if (!collectionId) {
      throw new Error("Collection ID is required");
    }

    return await this.syncAllFiles({ ...options, collectionId });
  }

  /**
   * Internal method to fetch a single page from API
   * @private
   */
  async fetchSinglePage(limit, cursor, collectionId = null) {
    const apiClient = await this.getApiClient();

    // Build query parameters
    const queryParams = new URLSearchParams();
    queryParams.append("limit", limit.toString());

    if (cursor) {
      queryParams.append("cursor", cursor);
    }

    if (collectionId) {
      queryParams.append("collection_id", collectionId);
    }

    // Build endpoint URL
    const endpoint = `/sync/files?${queryParams.toString()}`;

    console.log(`[SyncFileAPIService] → GET ${endpoint} via AuthManager`);

    // Make the API request
    const response = await apiClient.getMapleFile(endpoint);

    console.log(
      `[SyncFileAPIService] ← API Response via AuthManager: ${response.files?.length || 0} files, hasMore: ${response.has_more}`,
    );

    return response;
  }

  /**
   * Get a single sync file by ID
   * @param {string} fileId - File ID
   * @returns {Promise<Object>} - Sync file object
   */
  async getSyncFile(fileId) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    if (!fileId) {
      throw new Error("File ID is required");
    }

    try {
      console.log(
        `[SyncFileAPIService] Fetching sync file ${fileId} via AuthManager`,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/sync/files/${fileId}`);

      console.log(
        `[SyncFileAPIService] Sync file ${fileId} retrieved successfully`,
      );
      return response;
    } catch (error) {
      console.error(
        `[SyncFileAPIService] Failed to get sync file ${fileId}:`,
        error,
      );
      throw error;
    }
  }

  /**
   * Get loading state
   * @returns {boolean} - True if currently syncing from API
   */
  isLoadingSync() {
    return this.isLoading;
  }

  /**
   * Get debug information
   * @returns {Object} - Debug information
   */
  getDebugInfo() {
    return {
      serviceName: "SyncFileAPIService",
      managedBy: "AuthManager",
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default SyncFileAPIService;

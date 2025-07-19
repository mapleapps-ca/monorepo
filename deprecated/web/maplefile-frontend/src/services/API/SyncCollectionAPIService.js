// File: monorepo/web/maplefile-frontend/src/services/API/SyncCollectionAPIService.js
// Service for making API calls to sync collections (paginated requests)

class SyncCollectionAPIService {
  constructor(authManager) {
    this.authManager = authManager;
    this.isLoading = false;
    this._apiClient = null;
    console.log(
      "[SyncCollectionAPIService] API service initialized with AuthManager",
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
   * Sync ALL collections by automatically paginating through all API pages
   * This is the ONLY function you need to call!
   * @param {Object} options - Optional settings
   * @param {number} options.pageSize - Collections per page (default: 1000, max: 5000)
   * @returns {Promise<Array>} - Complete array of all sync collections from API
   */
  async syncAllCollections(options = {}) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncCollectionAPIService] Starting complete API sync via AuthManager...",
      );

      const allSyncCollections = [];
      let cursor = null;
      let hasMore = true;
      let pageCount = 0;
      const pageSize = Math.min(options.pageSize || 1000, 5000);

      while (hasMore) {
        pageCount++;
        console.log(
          `[SyncCollectionAPIService] Fetching API page ${pageCount} via AuthManager...`,
        );

        // Make single page request
        const response = await this.fetchSinglePage(pageSize, cursor);

        // Add sync collections to the result
        if (response.collections && response.collections.length > 0) {
          allSyncCollections.push(...response.collections);
          console.log(
            `[SyncCollectionAPIService] Page ${pageCount} via AuthManager: ${response.collections.length} sync collections (Total: ${allSyncCollections.length})`,
          );
        }

        // Check if there are more pages
        hasMore = response.has_more;
        cursor = response.next_cursor;
      }

      console.log(
        `[SyncCollectionAPIService] ✅ API sync complete via AuthManager! ${allSyncCollections.length} sync collections across ${pageCount} pages`,
      );
      return allSyncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionAPIService] ❌ API sync failed via AuthManager:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Internal method to fetch a single page from API
   * @private
   */
  async fetchSinglePage(limit, cursor) {
    const apiClient = await this.getApiClient();

    // Build query parameters
    const queryParams = new URLSearchParams();
    queryParams.append("limit", limit.toString());

    if (cursor) {
      queryParams.append("cursor", cursor);
    }

    // Build endpoint URL
    const endpoint = `/sync/collections?${queryParams.toString()}`;

    console.log(`[SyncCollectionAPIService] → GET ${endpoint} via AuthManager`);

    // Make the API request
    const response = await apiClient.getMapleFile(endpoint);

    console.log(
      `[SyncCollectionAPIService] ← API Response via AuthManager: ${response.collections?.length || 0} collections, hasMore: ${response.has_more}`,
    );

    return response;
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
      serviceName: "SyncCollectionAPIService",
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

export default SyncCollectionAPIService;

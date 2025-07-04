// SyncCollectionsService.js - SIMPLIFIED VERSION
// Just one main function: syncAllCollections() that returns everything

class SyncCollectionsService {
  constructor(authService) {
    this.authService = authService;
    this.isLoading = false;
    this._apiClient = null;
    console.log("[SyncCollectionsService] Simple service initialized");
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
   * Sync ALL collections by automatically paginating through all pages
   * This is the ONLY function you need to call!
   * @param {Object} options - Optional settings
   * @param {number} options.pageSize - Collections per page (default: 1000, max: 5000)
   * @returns {Promise<Array>} - Complete array of all collections
   */
  async syncAllCollections(options = {}) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      this.isLoading = true;
      console.log("[SyncCollectionsService] Starting complete sync...");

      const allCollections = [];
      let cursor = null;
      let hasMore = true;
      let pageCount = 0;
      const pageSize = Math.min(options.pageSize || 1000, 5000);

      while (hasMore) {
        pageCount++;
        console.log(`[SyncCollectionsService] Fetching page ${pageCount}...`);

        // Make single page request
        const response = await this.fetchSinglePage(pageSize, cursor);

        // Add collections to the result
        if (response.collections && response.collections.length > 0) {
          allCollections.push(...response.collections);
          console.log(
            `[SyncCollectionsService] Page ${pageCount}: ${response.collections.length} collections (Total: ${allCollections.length})`,
          );
        }

        // Check if there are more pages
        hasMore = response.has_more;
        cursor = response.next_cursor;
      }

      console.log(
        `[SyncCollectionsService] ✅ Complete! ${allCollections.length} collections across ${pageCount} pages`,
      );
      return allCollections;
    } catch (error) {
      console.error("[SyncCollectionsService] ❌ Sync failed:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Internal method to fetch a single page
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

    console.log(`[SyncCollectionsService] → GET ${endpoint}`);

    // Make the API request
    const response = await apiClient.getMapleFile(endpoint);

    console.log(
      `[SyncCollectionsService] ← Response: ${response.collections?.length || 0} collections, hasMore: ${response.has_more}`,
    );

    return response;
  }

  /**
   * Get loading state
   * @returns {boolean} - True if currently syncing
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
      isAuthenticated: this.authService.isAuthenticated(),
      isLoading: this.isLoading,
      canMakeRequests: this.authService.canMakeAuthenticatedRequests(),
      serviceName: "SyncCollectionsService (Simplified)",
    };
  }
}

export default SyncCollectionsService;

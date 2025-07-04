// SyncCollectionsService.js
// Service for syncing collections with paginated API calls
// Save this file to: src/services/SyncCollectionsService.js

class SyncCollectionsService {
  constructor(authService) {
    this.authService = authService;
    this.isLoading = false;
    this._apiClient = null;
    console.log("[SyncCollectionsService] Service initialized");
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
   * Sync collections with pagination support
   * @param {Object} options - Sync options
   * @param {number} options.limit - Maximum number of results (default: 1000, max: 5000)
   * @param {string} options.cursor - Base64 encoded cursor for pagination
   * @returns {Promise<Object>} - Sync response with collections, next_cursor, and has_more
   */
  async syncCollections(options = {}) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      this.isLoading = true;
      console.log(
        "[SyncCollectionsService] Syncing collections with options:",
        options,
      );

      const apiClient = await this.getApiClient();

      // Build query parameters
      const queryParams = new URLSearchParams();

      // Add limit parameter (default: 1000, max: 5000)
      if (options.limit !== undefined) {
        const limit = Math.min(Math.max(1, parseInt(options.limit)), 5000);
        queryParams.append("limit", limit.toString());
      }

      // Add cursor parameter if provided
      if (options.cursor) {
        queryParams.append("cursor", options.cursor);
      }

      // Build endpoint URL with query parameters
      const endpoint = `/sync/collections${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

      console.log(
        "[SyncCollectionsService] Requesting sync endpoint:",
        endpoint,
      );

      // Make the API request
      const response = await apiClient.getMapleFile(endpoint);

      console.log("[SyncCollectionsService] Sync response received:", {
        collectionsCount: response.collections?.length || 0,
        hasMore: response.has_more,
        hasNextCursor: !!response.next_cursor,
      });

      return response;
    } catch (error) {
      console.error(
        "[SyncCollectionsService] Failed to sync collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Sync all collections by automatically paginating through all pages
   * @param {Object} options - Sync options
   * @param {number} options.limit - Maximum number of results per page (default: 1000)
   * @param {Function} options.onPageReceived - Callback function called for each page of results
   * @returns {Promise<Array>} - Array of all collections
   */
  async syncAllCollections(options = {}) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      this.isLoading = true;
      console.log("[SyncCollectionsService] Syncing all collections");

      const allCollections = [];
      let cursor = null;
      let hasMore = true;
      let pageCount = 0;

      while (hasMore) {
        pageCount++;
        console.log(`[SyncCollectionsService] Fetching page ${pageCount}...`);

        const syncOptions = {
          limit: options.limit || 1000,
          ...(cursor && { cursor }),
        };

        const response = await this.syncCollections(syncOptions);

        // Add collections to the result
        if (response.collections && response.collections.length > 0) {
          allCollections.push(...response.collections);

          // Call the callback if provided
          if (
            options.onPageReceived &&
            typeof options.onPageReceived === "function"
          ) {
            try {
              await options.onPageReceived(
                response.collections,
                pageCount,
                response,
              );
            } catch (callbackError) {
              console.error(
                "[SyncCollectionsService] Error in page callback:",
                callbackError,
              );
            }
          }
        }

        // Check if there are more pages
        hasMore = response.has_more;
        cursor = response.next_cursor;

        console.log(`[SyncCollectionsService] Page ${pageCount} completed:`, {
          collectionsInPage: response.collections?.length || 0,
          totalCollections: allCollections.length,
          hasMore,
          nextCursor: cursor ? `${cursor.substring(0, 20)}...` : null,
        });
      }

      console.log(
        `[SyncCollectionsService] Sync complete: ${allCollections.length} collections across ${pageCount} pages`,
      );
      return allCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionsService] Failed to sync all collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * Sync collections since a specific modification time
   * @param {string} sinceModified - ISO timestamp to sync from
   * @param {Object} options - Additional sync options
   * @returns {Promise<Array>} - Array of collections modified since the specified time
   */
  async syncCollectionsSince(sinceModified, options = {}) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      console.log(
        "[SyncCollectionsService] Syncing collections since:",
        sinceModified,
      );

      const allCollections = [];
      let cursor = null;
      let hasMore = true;
      let foundOlderData = false;

      while (hasMore && !foundOlderData) {
        const syncOptions = {
          limit: options.limit || 1000,
          ...(cursor && { cursor }),
        };

        const response = await this.syncCollections(syncOptions);

        if (response.collections && response.collections.length > 0) {
          // Filter collections that are newer than the specified time
          const filteredCollections = response.collections.filter(
            (collection) => {
              const modifiedAt = new Date(collection.modified_at);
              const sinceDate = new Date(sinceModified);
              return modifiedAt >= sinceDate;
            },
          );

          // If we got fewer filtered results than total results, we've reached older data
          if (filteredCollections.length < response.collections.length) {
            foundOlderData = true;
          }

          allCollections.push(...filteredCollections);
        }

        hasMore = response.has_more;
        cursor = response.next_cursor;
      }

      console.log(
        `[SyncCollectionsService] Sync since ${sinceModified} complete: ${allCollections.length} collections`,
      );
      return allCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionsService] Failed to sync collections since:",
        error,
      );
      throw error;
    }
  }

  /**
   * Get loading state
   * @returns {boolean} - True if currently loading
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
      serviceName: "SyncCollectionsService",
    };
  }
}

export default SyncCollectionsService;

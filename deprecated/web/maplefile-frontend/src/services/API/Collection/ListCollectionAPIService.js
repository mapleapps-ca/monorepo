// File: monorepo/web/maplefile-frontend/src/services/API/Collection/ListCollectionAPIService.js
// List Collection API Service - Handles collection listing API calls

class ListCollectionAPIService {
  constructor(authManager) {
    // ListCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[ListCollectionAPIService] API service initialized with AuthManager dependency",
    );
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("../ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // List all collections owned by the authenticated user
  async listCollections() {
    try {
      console.log(
        "[ListCollectionAPIService] Listing user collections via API",
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections");

      console.log(
        "[ListCollectionAPIService] Collections retrieved successfully:",
        {
          totalCount: response.collections?.length || 0,
          hasCollections: !!(response.collections?.length > 0),
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[ListCollectionAPIService] Collection listing failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("401")) {
        throw new Error("Authentication required");
      } else if (error.message.includes("403")) {
        throw new Error("You don't have permission to list collections");
      }

      throw error;
    }
  }

  // List shared collections (for future use)
  async listSharedCollections() {
    try {
      console.log(
        "[ListCollectionAPIService] Listing shared collections via API",
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/shared");

      console.log(
        "[ListCollectionAPIService] Shared collections retrieved successfully:",
        {
          totalCount: response.collections?.length || 0,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[ListCollectionAPIService] Shared collection listing failed:",
        error,
      );
      throw error;
    }
  }

  // List collections with filters
  async listFilteredCollections(includeOwned = true, includeShared = false) {
    try {
      console.log("[ListCollectionAPIService] Listing filtered collections:", {
        includeOwned,
        includeShared,
      });

      const params = new URLSearchParams();
      if (includeOwned) params.append("include_owned", "true");
      if (includeShared) params.append("include_shared", "true");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections/filtered?${params.toString()}`,
      );

      console.log(
        "[ListCollectionAPIService] Filtered collections retrieved successfully:",
        {
          ownedCount: response.owned_collections?.length || 0,
          sharedCount: response.shared_collections?.length || 0,
          totalCount: response.total_count || 0,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[ListCollectionAPIService] Filtered collection listing failed:",
        error,
      );
      throw error;
    }
  }

  // List root collections (no parent)
  async listRootCollections() {
    try {
      console.log(
        "[ListCollectionAPIService] Listing root collections via API",
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/root");

      console.log(
        "[ListCollectionAPIService] Root collections retrieved successfully:",
        {
          totalCount: response.collections?.length || 0,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[ListCollectionAPIService] Root collection listing failed:",
        error,
      );
      throw error;
    }
  }

  // List collections by parent
  async listCollectionsByParent(parentId) {
    try {
      console.log(
        "[ListCollectionAPIService] Listing collections by parent:",
        parentId,
      );

      if (!parentId) {
        throw new Error("Parent ID is required");
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections-by-parent/${parentId}`,
      );

      console.log(
        "[ListCollectionAPIService] Collections by parent retrieved successfully:",
        {
          parentId,
          totalCount: response.collections?.length || 0,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[ListCollectionAPIService] Collections by parent listing failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("404")) {
        throw new Error("Parent collection not found");
      }

      throw error;
    }
  }

  // Validate collections response structure
  validateCollectionsResponse(response) {
    if (!response || typeof response !== "object") {
      throw new Error("Invalid response format");
    }

    if (!Array.isArray(response.collections)) {
      throw new Error("Response must contain collections array");
    }

    // Validate each collection
    response.collections.forEach((collection, index) => {
      const requiredFields = [
        "id",
        "collection_type",
        "encrypted_name",
        "created_at",
      ];
      const errors = [];

      requiredFields.forEach((field) => {
        if (!collection[field]) {
          errors.push(`Missing required field: ${field}`);
        }
      });

      // Validate collection type
      const validTypes = ["folder", "album"];
      if (
        collection.collection_type &&
        !validTypes.includes(collection.collection_type)
      ) {
        errors.push(
          `Invalid collection_type: ${collection.collection_type}. Must be 'folder' or 'album'`,
        );
      }

      if (errors.length > 0) {
        throw new Error(
          `Collection at index ${index} validation failed: ${errors.join(", ")}`,
        );
      }
    });

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "ListCollectionAPIService",
      managedBy: "AuthManager",
      isAuthenticated: this.authManager.isAuthenticated(),
      canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default ListCollectionAPIService;

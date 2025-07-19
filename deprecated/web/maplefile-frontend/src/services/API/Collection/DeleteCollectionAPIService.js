// File: monorepo/web/maplefile-frontend/src/services/API/Collection/DeleteCollectionAPIService.js
// Delete Collection API Service - Handles collection deletion API calls (soft delete)

class DeleteCollectionAPIService {
  constructor(authManager) {
    // DeleteCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[DeleteCollectionAPIService] API service initialized with AuthManager dependency",
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

  // Delete collection via API (soft delete)
  async deleteCollection(collectionId) {
    try {
      console.log("[DeleteCollectionAPIService] Deleting collection via API");
      console.log("[DeleteCollectionAPIService] Collection ID:", collectionId);

      // Validate collection ID
      this.validateCollectionId(collectionId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.deleteMapleFile(
        `/collections/${collectionId}`,
      );

      console.log(
        "[DeleteCollectionAPIService] Collection deleted successfully",
      );
      return response;
    } catch (error) {
      console.error(
        "[DeleteCollectionAPIService] Collection deletion failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("403")) {
        throw new Error("You don't have permission to delete this collection");
      } else if (error.message.includes("404")) {
        throw new Error("Collection not found");
      } else if (error.message.includes("409")) {
        throw new Error(
          "Collection cannot be deleted - it may contain files or subcollections",
        );
      } else if (error.message.includes("401")) {
        throw new Error("Authentication required");
      }

      throw error;
    }
  }

  // Batch delete multiple collections
  async deleteCollections(collectionIds) {
    try {
      console.log(
        "[DeleteCollectionAPIService] Deleting multiple collections:",
        collectionIds.length,
      );

      if (!Array.isArray(collectionIds) || collectionIds.length === 0) {
        throw new Error("Collection IDs array is required");
      }

      // Delete collections one by one (API doesn't support batch delete)
      const results = [];
      const errors = [];

      for (const collectionId of collectionIds) {
        try {
          const result = await this.deleteCollection(collectionId);
          results.push({ collectionId, result });
        } catch (error) {
          errors.push({ collectionId, error: error.message });
          console.warn(
            `[DeleteCollectionAPIService] Failed to delete collection ${collectionId}:`,
            error.message,
          );
        }
      }

      console.log(
        `[DeleteCollectionAPIService] Batch deletion completed: ${results.length} successful, ${errors.length} failed`,
      );

      return {
        successful: results,
        errors: errors,
        successCount: results.length,
        errorCount: errors.length,
      };
    } catch (error) {
      console.error(
        "[DeleteCollectionAPIService] Batch collection deletion failed:",
        error,
      );
      throw error;
    }
  }

  // Restore collection from soft delete (if supported by API)
  async restoreCollection(collectionId) {
    try {
      console.log(
        "[DeleteCollectionAPIService] Restoring collection:",
        collectionId,
      );

      this.validateCollectionId(collectionId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.putMapleFile(
        `/collections/${collectionId}/restore`,
        {},
      );

      console.log(
        "[DeleteCollectionAPIService] Collection restored successfully",
      );
      return response;
    } catch (error) {
      console.error(
        "[DeleteCollectionAPIService] Collection restoration failed:",
        error,
      );

      if (error.message.includes("404")) {
        throw new Error("Collection not found or already restored");
      }

      throw error;
    }
  }

  // Validate collection ID before API call
  validateCollectionId(collectionId) {
    if (!collectionId) {
      throw new Error("Collection ID is required");
    }

    // Validate UUID format
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(collectionId)) {
      throw new Error("Invalid collection ID format");
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "DeleteCollectionAPIService",
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

export default DeleteCollectionAPIService;

// File: monorepo/web/maplefile-frontend/src/services/API/Collection/GetCollectionAPIService.js
// Get Collection API Service - Handles collection retrieval API calls

class GetCollectionAPIService {
  constructor(authManager) {
    // GetCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[GetCollectionAPIService] API service initialized with AuthManager dependency",
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

  // Get collection by ID via API
  async getCollection(collectionId) {
    try {
      console.log(
        "[GetCollectionAPIService] Getting collection via API:",
        collectionId,
      );

      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      // Validate UUID format
      const uuidRegex =
        /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
      if (!uuidRegex.test(collectionId)) {
        throw new Error("Invalid collection ID format");
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections/${collectionId}`,
      );

      console.log(
        "[GetCollectionAPIService] Collection retrieved successfully:",
        {
          id: response.id,
          collection_type: response.collection_type,
          hasEncryptedName: !!response.encrypted_name,
          hasEncryptedKey: !!response.encrypted_collection_key,
          membersCount: response.members?.length || 0,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[GetCollectionAPIService] Collection retrieval failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("403")) {
        throw new Error("You don't have permission to access this collection");
      } else if (error.message.includes("404")) {
        throw new Error("Collection not found");
      } else if (error.message.includes("401")) {
        throw new Error("Authentication required");
      }

      throw error;
    }
  }

  // Get multiple collections by IDs (batch operation)
  async getCollections(collectionIds) {
    try {
      console.log(
        "[GetCollectionAPIService] Getting multiple collections:",
        collectionIds.length,
      );

      if (!Array.isArray(collectionIds) || collectionIds.length === 0) {
        throw new Error("Collection IDs array is required");
      }

      // Get collections one by one (API doesn't support batch get)
      const results = [];
      const errors = [];

      for (const collectionId of collectionIds) {
        try {
          const collection = await this.getCollection(collectionId);
          results.push(collection);
        } catch (error) {
          errors.push({ collectionId, error: error.message });
          console.warn(
            `[GetCollectionAPIService] Failed to get collection ${collectionId}:`,
            error.message,
          );
        }
      }

      console.log(
        `[GetCollectionAPIService] Batch retrieval completed: ${results.length} successful, ${errors.length} failed`,
      );

      return {
        collections: results,
        errors: errors,
        successCount: results.length,
        errorCount: errors.length,
      };
    } catch (error) {
      console.error(
        "[GetCollectionAPIService] Batch collection retrieval failed:",
        error,
      );
      throw error;
    }
  }

  // Check if collection exists (lightweight check)
  async collectionExists(collectionId) {
    try {
      console.log(
        "[GetCollectionAPIService] Checking if collection exists:",
        collectionId,
      );

      await this.getCollection(collectionId);
      return true;
    } catch (error) {
      if (error.message.includes("404")) {
        return false;
      }
      // Re-throw other errors (403, 401, etc.)
      throw error;
    }
  }

  // Validate collection data structure
  validateCollectionResponse(collection) {
    const requiredFields = [
      "id",
      "collection_type",
      "encrypted_name",
      "encrypted_collection_key",
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

    // Validate encrypted_collection_key structure
    if (collection.encrypted_collection_key) {
      const keyData = collection.encrypted_collection_key;
      if (!keyData.ciphertext || !keyData.nonce) {
        errors.push("encrypted_collection_key must have ciphertext and nonce");
      }
    }

    if (errors.length > 0) {
      throw new Error(`Collection validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "GetCollectionAPIService",
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

export default GetCollectionAPIService;

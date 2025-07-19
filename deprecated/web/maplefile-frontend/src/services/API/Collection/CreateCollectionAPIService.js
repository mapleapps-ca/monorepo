// File: monorepo/web/maplefile-frontend/src/services/API/Collection/CreateCollectionAPIService.js
// Create Collection API Service - Handles collection creation API calls (FIXED)

class CreateCollectionAPIService {
  constructor(authManager) {
    // CreateCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[CreateCollectionAPIService] API service initialized with AuthManager dependency",
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

  // Create collection via API
  async createCollection(collectionData) {
    try {
      console.log("[CreateCollectionAPIService] Creating collection via API");
      console.log("[CreateCollectionAPIService] Collection data:", {
        id: collectionData.id,
        collection_type: collectionData.collection_type,
        hasEncryptedName: !!collectionData.encrypted_name,
        hasEncryptedKey: !!collectionData.encrypted_collection_key,
        parent_id: collectionData.parent_id,
        membersCount: collectionData.members?.length || 0,
      });

      // Validate collection data before sending to API
      this.validateCollectionData(collectionData);

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        "/collections",
        collectionData,
      );

      console.log(
        "[CreateCollectionAPIService] Collection created successfully",
      );
      return response;
    } catch (error) {
      console.error(
        "[CreateCollectionAPIService] Collection creation failed:",
        error,
      );
      // Just throw the error - don't try to notify listeners (that's the manager's job)
      throw error;
    }
  }

  // Validate collection data before API call
  validateCollectionData(collectionData) {
    const requiredFields = [
      "id",
      "encrypted_name",
      "collection_type",
      "encrypted_collection_key",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (!collectionData[field]) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    // Validate collection type
    const validTypes = ["folder", "album"];
    if (
      collectionData.collection_type &&
      !validTypes.includes(collectionData.collection_type)
    ) {
      errors.push(
        `Invalid collection_type: ${collectionData.collection_type}. Must be 'folder' or 'album'`,
      );
    }

    // Validate encrypted_collection_key structure
    if (collectionData.encrypted_collection_key) {
      const keyData = collectionData.encrypted_collection_key;
      if (!keyData.ciphertext || !keyData.nonce) {
        errors.push("encrypted_collection_key must have ciphertext and nonce");
      }
    }

    if (errors.length > 0) {
      throw new Error(`Validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CreateCollectionAPIService",
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

export default CreateCollectionAPIService;

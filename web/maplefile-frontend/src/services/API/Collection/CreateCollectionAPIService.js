// File: monorepo/web/maplefile-frontend/src/services/API/Collection/CreateCollectionAPIService.js
// Create Collection API Service - Handles collection creation API calls

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
      this.isLoading = true;
      console.log("[CreateCollectionManager] Starting collection creation");
      console.log("[CreateCollectionManager] Collection data:", {
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
        hasParent: !!collectionData.parent_id,
      });

      // Validate input
      if (!collectionData.name) {
        throw new Error("Collection name is required");
      }

      console.log(
        "[CreateCollectionManager] Encrypting collection data using PasswordStorageService",
      );

      // Use CollectionCryptoService to encrypt all collection data for API (it uses PasswordStorageService automatically)
      const { apiData, collectionKey, collectionId } =
        await this.cryptoService.encryptCollectionForAPI(collectionData);

      console.log(
        "[CreateCollectionManager] Collection data encrypted, sending to API",
      );

      // Create collection via API
      const createdCollection = await this.apiService.createCollection(apiData);

      // Store collection key in memory for future use
      this.storageService.storeCollectionKey(collectionId, collectionKey);

      // Prepare decrypted collection for local storage
      const decryptedCollection = {
        ...createdCollection,
        name: collectionData.name, // Store decrypted name
        _originalEncryptedName: apiData.encrypted_name,
        _hasCollectionKey: true,
      };

      // Store in local storage
      this.storageService.storeCreatedCollection(decryptedCollection);

      // Notify listeners
      this.notifyCollectionCreationListeners("collection_created", {
        collectionId,
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
      });

      console.log(
        "[CreateCollectionManager] Collection created successfully:",
        collectionId,
      );

      return {
        collection: decryptedCollection,
        collectionId,
        success: true,
      };
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Collection creation failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionCreationListeners("collection_creation_failed", {
        error: error.message,
        collectionData,
      });

      throw error;
    } finally {
      this.isLoading = false;
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

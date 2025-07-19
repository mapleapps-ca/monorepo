// File: monorepo/web/maplefile-frontend/src/services/API/Collection/UpdateCollectionAPIService.js
// Update Collection API Service - Handles collection update API calls

class UpdateCollectionAPIService {
  constructor(authManager) {
    // UpdateCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[UpdateCollectionAPIService] API service initialized with AuthManager dependency",
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

  // Update collection via API
  async updateCollection(collectionId, updateData) {
    try {
      console.log("[UpdateCollectionAPIService] Updating collection via API");
      console.log("[UpdateCollectionAPIService] Collection ID:", collectionId);
      console.log("[UpdateCollectionAPIService] Update data:", {
        id: updateData.id,
        hasEncryptedName: !!updateData.encrypted_name,
        hasEncryptedKey: !!updateData.encrypted_collection_key,
        collection_type: updateData.collection_type,
        version: updateData.version,
      });

      // Validate collection data before sending to API
      this.validateUpdateData(collectionId, updateData);

      const apiClient = await this.getApiClient();
      const response = await apiClient.putMapleFile(
        `/collections/${collectionId}`,
        updateData,
      );

      console.log(
        "[UpdateCollectionAPIService] Collection updated successfully",
      );
      return response;
    } catch (error) {
      console.error(
        "[UpdateCollectionAPIService] Collection update failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("403")) {
        throw new Error("You don't have admin permission for this collection");
      } else if (error.message.includes("404")) {
        throw new Error("Collection not found");
      } else if (
        error.message.includes("409") ||
        error.message.includes("conflict")
      ) {
        throw new Error(
          "Version conflict - collection was modified by someone else. Please refresh and try again.",
        );
      } else if (
        error.message.includes("400") &&
        error.message.includes("version")
      ) {
        throw new Error("Invalid version number provided");
      } else if (error.message.includes("401")) {
        throw new Error("Authentication required");
      }

      throw error;
    }
  }

  // Validate update data before API call
  validateUpdateData(collectionId, updateData) {
    const requiredFields = ["id", "version"];
    const errors = [];

    requiredFields.forEach((field) => {
      if (updateData[field] === undefined || updateData[field] === null) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    // Validate collection ID matches
    if (updateData.id !== collectionId) {
      errors.push(
        `Collection ID mismatch: URL has ${collectionId}, body has ${updateData.id}`,
      );
    }

    // Validate collection type if provided
    if (updateData.collection_type) {
      const validTypes = ["folder", "album"];
      if (!validTypes.includes(updateData.collection_type)) {
        errors.push(
          `Invalid collection_type: ${updateData.collection_type}. Must be 'folder' or 'album'`,
        );
      }
    }

    // Validate encrypted_collection_key structure if provided
    if (updateData.encrypted_collection_key) {
      const keyData = updateData.encrypted_collection_key;

      // According to architecture docs, encrypted_collection_key should be an object with base64 strings
      if (typeof keyData === "object" && keyData !== null) {
        if (!keyData.ciphertext || typeof keyData.ciphertext !== "string") {
          errors.push(
            "encrypted_collection_key must have a ciphertext field (base64 string)",
          );
        }
        if (!keyData.nonce || typeof keyData.nonce !== "string") {
          errors.push(
            "encrypted_collection_key must have a nonce field (base64 string)",
          );
        }
        if (typeof keyData.key_version !== "number") {
          errors.push(
            "encrypted_collection_key must have a numeric key_version",
          );
        }
        // rotated_at and previous_keys are optional
      } else {
        errors.push(
          "encrypted_collection_key must be an object with ciphertext, nonce, and key_version",
        );
      }
    }

    // Validate version is a number
    if (typeof updateData.version !== "number") {
      errors.push("version must be a number");
    }

    if (errors.length > 0) {
      throw new Error(`Validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "UpdateCollectionAPIService",
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

export default UpdateCollectionAPIService;

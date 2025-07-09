// File: monorepo/web/maplefile-frontend/src/services/API/Collection/ShareCollectionAPIService.js
// Share Collection API Service - Handles collection sharing API calls

class ShareCollectionAPIService {
  constructor(authManager) {
    // ShareCollectionAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[ShareCollectionAPIService] API service initialized with AuthManager dependency",
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

  // Share collection with another user via API
  async shareCollection(collectionId, shareData) {
    try {
      console.log("[ShareCollectionAPIService] Sharing collection via API");
      console.log("[ShareCollectionAPIService] Collection ID:", collectionId);
      console.log("[ShareCollectionAPIService] Share data:", {
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level,
        share_with_descendants: shareData.share_with_descendants,
        hasEncryptedKey: !!shareData.encrypted_collection_key,
      });

      // Validate share data before sending to API
      this.validateShareData(collectionId, shareData);

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/collections/${collectionId}/share`,
        shareData,
      );

      console.log("[ShareCollectionAPIService] Collection shared successfully");
      return response;
    } catch (error) {
      console.error(
        "[ShareCollectionAPIService] Collection sharing failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("403")) {
        throw new Error("You don't have admin permission for this collection");
      } else if (error.message.includes("404")) {
        throw new Error("Collection not found");
      } else if (error.message.includes("400")) {
        if (error.message.includes("recipient")) {
          throw new Error("Invalid recipient user ID or email");
        } else if (error.message.includes("permission")) {
          throw new Error("Invalid permission level");
        } else {
          throw new Error("Invalid sharing data provided");
        }
      } else if (error.message.includes("401")) {
        throw new Error("Authentication required");
      }

      throw error;
    }
  }

  // Remove member from collection via API
  async removeMember(collectionId, removeData) {
    try {
      console.log("[ShareCollectionAPIService] Removing member via API");
      console.log("[ShareCollectionAPIService] Collection ID:", collectionId);
      console.log("[ShareCollectionAPIService] Remove data:", {
        recipient_id: removeData.recipient_id,
        remove_from_descendants: removeData.remove_from_descendants,
      });

      // Validate remove data before sending to API
      this.validateRemoveData(collectionId, removeData);

      const apiClient = await this.getApiClient();
      const response = await apiClient.deleteMapleFile(
        `/collections/${collectionId}/members`,
        {
          body: JSON.stringify(removeData),
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      console.log("[ShareCollectionAPIService] Member removed successfully");
      return response;
    } catch (error) {
      console.error(
        "[ShareCollectionAPIService] Member removal failed:",
        error,
      );

      // Handle specific API errors
      if (error.message.includes("403")) {
        throw new Error("You don't have admin permission for this collection");
      } else if (error.message.includes("404")) {
        throw new Error("Collection or member not found");
      } else if (error.message.includes("400")) {
        throw new Error("Invalid member removal data provided");
      } else if (error.message.includes("401")) {
        throw new Error("Authentication required");
      }

      throw error;
    }
  }

  // Get collection members (if API supports it)
  async getCollectionMembers(collectionId) {
    try {
      console.log(
        "[ShareCollectionAPIService] Getting collection members via API",
      );
      console.log("[ShareCollectionAPIService] Collection ID:", collectionId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections/${collectionId}`,
      );

      console.log(
        "[ShareCollectionAPIService] Collection members retrieved successfully",
      );
      return response.members || [];
    } catch (error) {
      console.error(
        "[ShareCollectionAPIService] Failed to get collection members:",
        error,
      );
      throw error;
    }
  }

  // Validate share data before API call
  validateShareData(collectionId, shareData) {
    const requiredFields = [
      "collection_id",
      "recipient_id",
      "recipient_email",
      "permission_level",
      "encrypted_collection_key",
      "share_with_descendants",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (shareData[field] === undefined || shareData[field] === null) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    // Validate collection ID matches
    if (shareData.collection_id !== collectionId) {
      errors.push(
        `Collection ID mismatch: URL has ${collectionId}, body has ${shareData.collection_id}`,
      );
    }

    // Validate permission level
    const validPermissions = ["read_only", "read_write", "admin"];
    if (!validPermissions.includes(shareData.permission_level)) {
      errors.push(
        `Invalid permission_level: ${shareData.permission_level}. Must be one of: ${validPermissions.join(", ")}`,
      );
    }

    // Validate recipient_id format (should be UUID)
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(shareData.recipient_id)) {
      errors.push("recipient_id must be a valid UUID");
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(shareData.recipient_email)) {
      errors.push("recipient_email must be a valid email address");
    }

    // Validate encrypted_collection_key (should be base64 string or byte array)
    if (shareData.encrypted_collection_key) {
      if (typeof shareData.encrypted_collection_key === "string") {
        // Should be base64 string
        try {
          atob(shareData.encrypted_collection_key);
        } catch (e) {
          errors.push("encrypted_collection_key must be a valid base64 string");
        }
      } else if (Array.isArray(shareData.encrypted_collection_key)) {
        // Should be array of numbers (bytes)
        const isValidByteArray = shareData.encrypted_collection_key.every(
          (byte) => typeof byte === "number" && byte >= 0 && byte <= 255,
        );
        if (!isValidByteArray) {
          errors.push(
            "encrypted_collection_key array must contain only bytes (0-255)",
          );
        }
      } else {
        errors.push(
          "encrypted_collection_key must be a base64 string or byte array",
        );
      }
    }

    // Validate share_with_descendants is boolean
    if (typeof shareData.share_with_descendants !== "boolean") {
      errors.push("share_with_descendants must be a boolean");
    }

    if (errors.length > 0) {
      throw new Error(`Share validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Validate remove data before API call
  validateRemoveData(collectionId, removeData) {
    const requiredFields = [
      "collection_id",
      "recipient_id",
      "remove_from_descendants",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (removeData[field] === undefined || removeData[field] === null) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    // Validate collection ID matches
    if (removeData.collection_id !== collectionId) {
      errors.push(
        `Collection ID mismatch: URL has ${collectionId}, body has ${removeData.collection_id}`,
      );
    }

    // Validate recipient_id format (should be UUID)
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(removeData.recipient_id)) {
      errors.push("recipient_id must be a valid UUID");
    }

    // Validate remove_from_descendants is boolean
    if (typeof removeData.remove_from_descendants !== "boolean") {
      errors.push("remove_from_descendants must be a boolean");
    }

    if (errors.length > 0) {
      throw new Error(`Remove validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "ShareCollectionAPIService",
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

export default ShareCollectionAPIService;

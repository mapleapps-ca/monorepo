// File: monorepo/web/maplefile-frontend/src/services/API/File/CreateFileAPIService.js
// Create File API Service - Handles pending file creation API calls

class CreateFileAPIService {
  constructor(authManager) {
    // CreateFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[CreateFileAPIService] API service initialized with AuthManager dependency",
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

  // Create pending file via API
  async createPendingFile(fileData) {
    try {
      console.log("[CreateFileAPIService] Creating pending file via API");
      console.log("[CreateFileAPIService] File data:", {
        id: fileData.id,
        collection_id: fileData.collection_id,
        hasEncryptedMetadata: !!fileData.encrypted_metadata,
        hasEncryptedFileKey: !!fileData.encrypted_file_key,
        expected_file_size_in_bytes: fileData.expected_file_size_in_bytes,
        expected_thumbnail_size_in_bytes:
          fileData.expected_thumbnail_size_in_bytes,
      });

      // Validate file data before sending to API
      this.validateFileData(fileData);

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        "/files/pending",
        fileData,
      );

      console.log("[CreateFileAPIService] Pending file created successfully");
      return response;
    } catch (error) {
      console.error(
        "[CreateFileAPIService] Pending file creation failed:",
        error,
      );
      throw error;
    }
  }

  // Validate file data before API call
  validateFileData(fileData) {
    const requiredFields = [
      "id",
      "collection_id",
      "encrypted_metadata",
      "encrypted_file_key",
      "encryption_version",
      "encrypted_hash",
      "expected_file_size_in_bytes",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (!fileData[field] && fileData[field] !== 0) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    // Validate encrypted_file_key structure
    if (fileData.encrypted_file_key) {
      const keyData = fileData.encrypted_file_key;
      if (
        !keyData.ciphertext ||
        !keyData.nonce ||
        keyData.key_version === undefined
      ) {
        errors.push(
          "encrypted_file_key must have ciphertext, nonce, and key_version",
        );
      }
    }

    // Validate file size
    if (
      fileData.expected_file_size_in_bytes &&
      fileData.expected_file_size_in_bytes < 0
    ) {
      errors.push("expected_file_size_in_bytes must be non-negative");
    }

    if (errors.length > 0) {
      throw new Error(`Validation failed: ${errors.join(", ")}`);
    }

    return true;
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CreateFileAPIService",
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

export default CreateFileAPIService;

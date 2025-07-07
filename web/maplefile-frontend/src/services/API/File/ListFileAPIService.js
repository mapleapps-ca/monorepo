// File: monorepo/web/maplefile-frontend/src/services/API/File/ListFileAPIService.js
// List File API Service - Handles API calls for listing files

class ListFileAPIService {
  constructor(authManager) {
    // ListFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[ListFileAPIService] API service initialized with AuthManager dependency",
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

  // List files by collection with optional state filtering
  async listFilesByCollection(collectionId, includeStates = null) {
    try {
      console.log(
        "[ListFileAPIService] Listing files for collection:",
        collectionId,
      );
      console.log("[ListFileAPIService] Include states:", includeStates);

      const apiClient = await this.getApiClient();
      let url = `/collections/${collectionId}/files`;

      // Add state filtering if specified
      if (includeStates && Array.isArray(includeStates)) {
        const params = new URLSearchParams();
        includeStates.forEach((state) => params.append("states", state));
        url += `?${params.toString()}`;
      }

      const response = await apiClient.getMapleFile(url);

      console.log("[ListFileAPIService] Files retrieved:", {
        count: response.files?.length || 0,
        collectionId,
      });

      return response;
    } catch (error) {
      console.error("[ListFileAPIService] Failed to list files:", error);
      throw error;
    }
  }

  // Get file by ID
  async getFileById(fileId) {
    try {
      console.log("[ListFileAPIService] Getting file by ID:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/files/${fileId}`);

      console.log("[ListFileAPIService] File retrieved:", fileId);
      return response;
    } catch (error) {
      console.error("[ListFileAPIService] Failed to get file:", error);
      throw error;
    }
  }

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[ListFileAPIService] Getting download URL for:", fileId);

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-url`,
        requestData,
      );

      console.log("[ListFileAPIService] Download URL generated for:", fileId);
      return response;
    } catch (error) {
      console.error("[ListFileAPIService] Failed to get download URL:", error);
      throw error;
    }
  }

  // Sync files for offline support
  async syncFiles(cursor = null, limit = 5000) {
    try {
      console.log("[ListFileAPIService] Syncing files", { cursor, limit });

      const apiClient = await this.getApiClient();
      const params = new URLSearchParams({ limit: limit.toString() });

      if (cursor) {
        params.append("cursor", cursor);
      }

      const response = await apiClient.getMapleFile(`/sync/files?${params}`);

      console.log("[ListFileAPIService] Files synced:", {
        count: response.files?.length || 0,
        hasMore: response.has_more || false,
      });

      return response;
    } catch (error) {
      console.error("[ListFileAPIService] Failed to sync files:", error);
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "ListFileAPIService",
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

export default ListFileAPIService;

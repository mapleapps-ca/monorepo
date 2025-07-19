// File: monorepo/web/maplefile-frontend/src/services/API/File/RecentFileAPIService.js
// Recent File API Service - Handles API calls for listing recent files

class RecentFileAPIService {
  constructor(authManager) {
    // RecentFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[RecentFileAPIService] API service initialized with AuthManager dependency",
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

  // List recent files with pagination
  async listRecentFiles(limit = 30, cursor = null) {
    try {
      console.log("[RecentFileAPIService] Listing recent files:", {
        limit,
        cursor: cursor ? cursor.substring(0, 20) + "..." : null,
      });

      // Validate limit
      if (limit > 100) {
        console.warn(
          "[RecentFileAPIService] Limit exceeds maximum of 100, capping to 100",
        );
        limit = 100;
      }

      const apiClient = await this.getApiClient();
      const params = new URLSearchParams({ limit: limit.toString() });

      if (cursor) {
        params.append("cursor", cursor);
      }

      const url = `/files/recent?${params.toString()}`;
      const response = await apiClient.getMapleFile(url);

      console.log("[RecentFileAPIService] Recent files retrieved:", {
        count: response.files?.length || 0,
        hasMore: response.has_more || false,
        nextCursor: response.next_cursor ? "present" : "none",
      });

      return response;
    } catch (error) {
      console.error(
        "[RecentFileAPIService] Failed to list recent files:",
        error,
      );
      throw error;
    }
  }

  // Get file by ID (for individual file operations)
  async getFileById(fileId) {
    try {
      console.log("[RecentFileAPIService] Getting file by ID:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/files/${fileId}`);

      console.log("[RecentFileAPIService] File retrieved:", fileId);
      return response;
    } catch (error) {
      console.error("[RecentFileAPIService] Failed to get file:", error);
      throw error;
    }
  }

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[RecentFileAPIService] Getting download URL for:", fileId);

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-url`,
        requestData,
      );

      console.log("[RecentFileAPIService] Download URL generated for:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[RecentFileAPIService] Failed to get download URL:",
        error,
      );
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "RecentFileAPIService",
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

export default RecentFileAPIService;

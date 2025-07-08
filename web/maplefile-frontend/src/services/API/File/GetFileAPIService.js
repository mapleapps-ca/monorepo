// File: monorepo/web/maplefile-frontend/src/services/API/File/GetFileAPIService.js
// Get File API Service - Handles API calls for retrieving individual file details

class GetFileAPIService {
  constructor(authManager) {
    // GetFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[GetFileAPIService] API service initialized with AuthManager dependency",
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

  // Get file by ID with full details
  async getFileById(fileId, includeVersionHistory = false) {
    try {
      console.log("[GetFileAPIService] Getting file by ID:", fileId);
      console.log(
        "[GetFileAPIService] Include version history:",
        includeVersionHistory,
      );

      const apiClient = await this.getApiClient();
      let url = `/files/${fileId}`;

      // Add version history parameter if requested
      if (includeVersionHistory) {
        url += "?include_versions=true";
      }

      const response = await apiClient.getMapleFile(url);

      console.log("[GetFileAPIService] File retrieved:", {
        id: fileId,
        hasVersionHistory: !!response.versions,
        versionCount: response.versions?.length || 0,
      });

      return response;
    } catch (error) {
      console.error("[GetFileAPIService] Failed to get file:", error);
      throw error;
    }
  }

  // Get file version history
  async getFileVersionHistory(fileId) {
    try {
      console.log("[GetFileAPIService] Getting version history for:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/versions`,
      );

      console.log("[GetFileAPIService] Version history retrieved:", {
        fileId,
        versionCount: response.versions?.length || 0,
      });

      return response;
    } catch (error) {
      console.error(
        "[GetFileAPIService] Failed to get version history:",
        error,
      );
      throw error;
    }
  }

  // Get file metadata only (lightweight)
  async getFileMetadata(fileId) {
    try {
      console.log("[GetFileAPIService] Getting file metadata:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/metadata`,
      );

      console.log("[GetFileAPIService] File metadata retrieved:", fileId);
      return response;
    } catch (error) {
      console.error("[GetFileAPIService] Failed to get file metadata:", error);
      throw error;
    }
  }

  // Get file permissions and sharing info
  async getFilePermissions(fileId) {
    try {
      console.log("[GetFileAPIService] Getting file permissions:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/permissions`,
      );

      console.log("[GetFileAPIService] File permissions retrieved:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[GetFileAPIService] Failed to get file permissions:",
        error,
      );
      throw error;
    }
  }

  // Get file usage statistics
  async getFileStats(fileId) {
    try {
      console.log("[GetFileAPIService] Getting file statistics:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/files/${fileId}/stats`);

      console.log("[GetFileAPIService] File statistics retrieved:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[GetFileAPIService] Failed to get file statistics:",
        error,
      );
      throw error;
    }
  }

  // Check if file exists and is accessible
  async checkFileExists(fileId) {
    try {
      console.log("[GetFileAPIService] Checking if file exists:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/files/${fileId}/exists`);

      console.log("[GetFileAPIService] File existence check:", {
        fileId,
        exists: response.exists,
        accessible: response.accessible,
      });

      return response;
    } catch (error) {
      console.error(
        "[GetFileAPIService] Failed to check file existence:",
        error,
      );
      throw error;
    }
  }

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[GetFileAPIService] Getting download URL for:", fileId);

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-url`,
        requestData,
      );

      console.log("[GetFileAPIService] Download URL generated for:", fileId);
      return response;
    } catch (error) {
      console.error("[GetFileAPIService] Failed to get download URL:", error);
      throw error;
    }
  }

  // Get file with all related data (comprehensive)
  async getFileComplete(fileId) {
    try {
      console.log("[GetFileAPIService] Getting complete file data:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/complete`,
      );

      console.log("[GetFileAPIService] Complete file data retrieved:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[GetFileAPIService] Failed to get complete file data:",
        error,
      );
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "GetFileAPIService",
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

export default GetFileAPIService;

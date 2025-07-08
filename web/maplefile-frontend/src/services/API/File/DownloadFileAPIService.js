// File: monorepo/web/maplefile-frontend/src/services/API/File/DownloadFileAPIService.js
// Download File API Service - Handles API calls for downloading files

class DownloadFileAPIService {
  constructor(authManager) {
    // DownloadFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[DownloadFileAPIService] API service initialized with AuthManager dependency",
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

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[DownloadFileAPIService] Getting download URL for:", fileId);

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-url`,
        requestData,
      );

      console.log(
        "[DownloadFileAPIService] Download URL generated for:",
        fileId,
      );
      return response;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to get download URL:",
        error,
      );
      throw error;
    }
  }

  // Get presigned thumbnail download URL
  async getPresignedThumbnailUrl(fileId, urlDuration = null) {
    try {
      console.log(
        "[DownloadFileAPIService] Getting thumbnail URL for:",
        fileId,
      );

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/thumbnail-url`,
        requestData,
      );

      console.log(
        "[DownloadFileAPIService] Thumbnail URL generated for:",
        fileId,
      );
      return response;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to get thumbnail URL:",
        error,
      );
      throw error;
    }
  }

  // Download file content from S3 using presigned URL
  async downloadFileFromS3(presignedUrl, onProgress = null) {
    try {
      console.log("[DownloadFileAPIService] Downloading file from S3");

      const response = await fetch(presignedUrl, {
        method: "GET",
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 download failed with status: ${response.status}`);
      }

      // Handle progress if callback provided
      if (onProgress && response.body) {
        const contentLength = response.headers.get("content-length");
        const total = contentLength ? parseInt(contentLength, 10) : 0;
        let received = 0;

        const reader = response.body.getReader();
        const chunks = [];

        while (true) {
          const { done, value } = await reader.read();

          if (done) break;

          chunks.push(value);
          received += value.length;

          if (total > 0) {
            onProgress(received, total, (received / total) * 100);
          }
        }

        const blob = new Blob(chunks);
        console.log(
          "[DownloadFileAPIService] File downloaded from S3 with progress tracking",
        );
        return blob;
      } else {
        const blob = await response.blob();
        console.log(
          "[DownloadFileAPIService] File downloaded from S3 successfully",
        );
        return blob;
      }
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to download file from S3:",
        error,
      );

      // Provide better error messages for common issues
      if (error.name === "TypeError" && error.message === "Failed to fetch") {
        throw new Error(
          "Unable to download file from cloud storage. This may be a network issue or CORS configuration problem.",
        );
      }

      throw error;
    }
  }

  // Download thumbnail from S3 using presigned URL
  async downloadThumbnailFromS3(presignedUrl, onProgress = null) {
    try {
      console.log("[DownloadFileAPIService] Downloading thumbnail from S3");

      const response = await fetch(presignedUrl, {
        method: "GET",
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(
          `S3 thumbnail download failed with status: ${response.status}`,
        );
      }

      const blob = await response.blob();
      console.log(
        "[DownloadFileAPIService] Thumbnail downloaded from S3 successfully",
      );

      return blob;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to download thumbnail from S3:",
        error,
      );
      throw error;
    }
  }

  // Get file metadata for download preparation
  async getFileForDownload(fileId) {
    try {
      console.log(
        "[DownloadFileAPIService] Getting file metadata for download:",
        fileId,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(`/files/${fileId}`);

      console.log(
        "[DownloadFileAPIService] File metadata retrieved for download:",
        fileId,
      );
      return response;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to get file for download:",
        error,
      );
      throw error;
    }
  }

  // Verify file availability for download
  async checkFileDownloadAvailability(fileId) {
    try {
      console.log(
        "[DownloadFileAPIService] Checking file download availability:",
        fileId,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/download-status`,
      );

      console.log(
        "[DownloadFileAPIService] File download availability checked:",
        {
          fileId,
          available: response.available,
          reason: response.reason,
        },
      );

      return response;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to check file download availability:",
        error,
      );
      throw error;
    }
  }

  // Get batch download URLs for multiple files
  async getBatchDownloadUrls(fileIds, urlDuration = null) {
    try {
      console.log(
        "[DownloadFileAPIService] Getting batch download URLs for:",
        fileIds.length,
        "files",
      );

      const requestData = {
        file_ids: fileIds,
      };

      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        "/files/batch/download-urls",
        requestData,
      );

      console.log(
        "[DownloadFileAPIService] Batch download URLs generated for:",
        fileIds.length,
        "files",
      );
      return response;
    } catch (error) {
      console.error(
        "[DownloadFileAPIService] Failed to get batch download URLs:",
        error,
      );
      throw error;
    }
  }

  // Report download completion (for analytics/tracking)
  async reportDownloadCompletion(fileId, downloadMetadata = {}) {
    try {
      console.log(
        "[DownloadFileAPIService] Reporting download completion for:",
        fileId,
      );

      const requestData = {
        file_id: fileId,
        downloaded_at: new Date().toISOString(),
        ...downloadMetadata,
      };

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-completed`,
        requestData,
      );

      console.log(
        "[DownloadFileAPIService] Download completion reported for:",
        fileId,
      );
      return response;
    } catch (error) {
      console.warn(
        "[DownloadFileAPIService] Failed to report download completion:",
        error,
      );
      // Don't throw error as this is not critical
      return null;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "DownloadFileAPIService",
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

export default DownloadFileAPIService;

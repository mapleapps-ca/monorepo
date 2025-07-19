// File: monorepo/web/maplefile-frontend/src/services/API/File/DeleteFileAPIService.js
// Delete File API Service - Handles API calls for deleting, restoring, and managing file lifecycle

class DeleteFileAPIService {
  constructor(authManager) {
    // DeleteFileAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[DeleteFileAPIService] API service initialized with AuthManager dependency",
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

  // Soft delete a single file (creates tombstone)
  async deleteFile(fileId, reason = null) {
    try {
      console.log("[DeleteFileAPIService] Soft deleting file:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {};

      if (reason) {
        requestData.reason = reason;
      }

      // Use DELETE method for file deletion
      const response = await apiClient.deleteMapleFile(`/files/${fileId}`, {
        body: requestData.reason ? JSON.stringify(requestData) : undefined,
        headers: requestData.reason
          ? { "Content-Type": "application/json" }
          : {},
      });

      console.log("[DeleteFileAPIService] File soft deleted:", {
        fileId,
        newState: response?.file?.state,
        tombstoneVersion: response?.file?.tombstone_version,
        tombstoneExpiry: response?.file?.tombstone_expiry,
      });

      return response;
    } catch (error) {
      console.error("[DeleteFileAPIService] Failed to delete file:", error);
      throw error;
    }
  }

  // Restore a soft-deleted file
  async restoreFile(fileId, reason = null) {
    try {
      console.log("[DeleteFileAPIService] Restoring file:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {};

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        `/files/${fileId}/restore`,
        requestData,
      );

      console.log("[DeleteFileAPIService] File restored:", {
        fileId,
        newState: response?.file?.state,
        newVersion: response?.file?.version,
      });

      return response;
    } catch (error) {
      console.error("[DeleteFileAPIService] Failed to restore file:", error);
      throw error;
    }
  }

  // Archive a file (soft archive - different from delete)
  async archiveFile(fileId, reason = null) {
    try {
      console.log("[DeleteFileAPIService] Archiving file:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {};

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        `/files/${fileId}/archive`,
        requestData,
      );

      console.log("[DeleteFileAPIService] File archived:", {
        fileId,
        newState: response?.file?.state,
        newVersion: response?.file?.version,
      });

      return response;
    } catch (error) {
      console.error("[DeleteFileAPIService] Failed to archive file:", error);
      throw error;
    }
  }

  // Unarchive a file (restore from archive)
  async unarchiveFile(fileId, reason = null) {
    try {
      console.log("[DeleteFileAPIService] Unarchiving file:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {};

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        `/files/${fileId}/unarchive`,
        requestData,
      );

      console.log("[DeleteFileAPIService] File unarchived:", {
        fileId,
        newState: response?.file?.state,
        newVersion: response?.file?.version,
      });

      return response;
    } catch (error) {
      console.error("[DeleteFileAPIService] Failed to unarchive file:", error);
      throw error;
    }
  }

  // Delete multiple files in batch
  async deleteMultipleFiles(fileIds, reason = null) {
    try {
      console.log(
        "[DeleteFileAPIService] Batch deleting files:",
        fileIds.length,
      );

      const apiClient = await this.getApiClient();
      const requestData = {
        file_ids: fileIds,
      };

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        "/files/batch/delete",
        requestData,
      );

      const successCount =
        response.results?.filter((r) => r.success)?.length || 0;
      const errorCount =
        response.results?.filter((r) => !r.success)?.length || 0;

      console.log("[DeleteFileAPIService] Batch delete completed:", {
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
      });

      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to batch delete files:",
        error,
      );
      throw error;
    }
  }

  // Restore multiple files in batch
  async restoreMultipleFiles(fileIds, reason = null) {
    try {
      console.log(
        "[DeleteFileAPIService] Batch restoring files:",
        fileIds.length,
      );

      const apiClient = await this.getApiClient();
      const requestData = {
        file_ids: fileIds,
      };

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        "/files/batch/restore",
        requestData,
      );

      const successCount =
        response.results?.filter((r) => r.success)?.length || 0;
      const errorCount =
        response.results?.filter((r) => !r.success)?.length || 0;

      console.log("[DeleteFileAPIService] Batch restore completed:", {
        total: fileIds.length,
        successful: successCount,
        failed: errorCount,
      });

      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to batch restore files:",
        error,
      );
      throw error;
    }
  }

  // Permanently delete a file (only if tombstone expired or special permissions)
  async permanentlyDeleteFile(fileId, reason = null) {
    try {
      console.log("[DeleteFileAPIService] Permanently deleting file:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {};

      if (reason) {
        requestData.reason = reason;
      }

      const response = await apiClient.postMapleFile(
        `/files/${fileId}/permanent-delete`,
        requestData,
      );

      console.log("[DeleteFileAPIService] File permanently deleted:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to permanently delete file:",
        error,
      );
      throw error;
    }
  }

  // Get file deletion history/audit log
  async getFileDeletionHistory(fileId) {
    try {
      console.log(
        "[DeleteFileAPIService] Getting deletion history for:",
        fileId,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/deletion-history`,
      );

      console.log("[DeleteFileAPIService] Deletion history retrieved:", fileId);
      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to get deletion history:",
        error,
      );
      throw error;
    }
  }

  // Check if file can be deleted
  async checkFileDeletable(fileId) {
    try {
      console.log(
        "[DeleteFileAPIService] Checking if file can be deleted:",
        fileId,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/can-delete`,
      );

      console.log("[DeleteFileAPIService] Delete check result:", {
        fileId,
        canDelete: response.can_delete,
        canRestore: response.can_restore,
        canPermanentlyDelete: response.can_permanently_delete,
      });

      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to check file deletability:",
        error,
      );
      throw error;
    }
  }

  // Extend tombstone expiry
  async extendTombstoneExpiry(fileId, extensionDays = 30) {
    try {
      console.log("[DeleteFileAPIService] Extending tombstone expiry:", fileId);

      const apiClient = await this.getApiClient();
      const requestData = {
        extension_days: extensionDays,
      };

      const response = await apiClient.postMapleFile(
        `/files/${fileId}/extend-tombstone`,
        requestData,
      );

      console.log("[DeleteFileAPIService] Tombstone expiry extended:", {
        fileId,
        newExpiry: response?.file?.tombstone_expiry,
      });

      return response;
    } catch (error) {
      console.error(
        "[DeleteFileAPIService] Failed to extend tombstone expiry:",
        error,
      );
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "DeleteFileAPIService",
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

export default DeleteFileAPIService;

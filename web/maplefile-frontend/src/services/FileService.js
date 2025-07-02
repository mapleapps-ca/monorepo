// FileService for managing MapleFile files with end-to-end encryption
// Supports all file operations including upload, download, sharing, and synchronization

class FileService {
  constructor() {
    this._apiClient = null;
    this.cache = new Map(); // Simple cache for file metadata
    this.isLoading = false;
    this.uploadQueue = new Map(); // Track pending uploads
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("./ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // 1. Create Pending File
  async createPendingFile(fileData) {
    try {
      this.isLoading = true;
      console.log("[FileService] Creating pending file");

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        "/files/pending",
        fileData,
      );

      // Cache the pending file metadata
      if (response.file) {
        this.cache.set(response.file.id, response.file);
        // Track in upload queue
        this.uploadQueue.set(response.file.id, {
          file: response.file,
          uploadUrl: response.presigned_upload_url,
          thumbnailUrl: response.presigned_thumbnail_url,
          expirationTime: response.upload_url_expiration_time,
        });
      }

      console.log("[FileService] Pending file created:", response.file?.id);
      return response;
    } catch (error) {
      console.error("[FileService] Failed to create pending file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 2. Complete File Upload
  async completeFileUpload(fileId, completionData = {}) {
    try {
      this.isLoading = true;
      console.log("[FileService] Completing file upload:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/complete`,
        completionData,
      );

      // Update cache with completed file
      if (response.file) {
        this.cache.set(fileId, response.file);
      }

      // Remove from upload queue
      this.uploadQueue.delete(fileId);

      console.log("[FileService] File upload completed:", fileId);
      return response;
    } catch (error) {
      console.error("[FileService] Failed to complete file upload:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 3. Get File by ID
  async getFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Getting file:", fileId);

      // Check cache first
      if (this.cache.has(fileId)) {
        console.log("[FileService] File found in cache");
        return this.cache.get(fileId);
      }

      const apiClient = await this.getApiClient();
      const file = await apiClient.getMapleFile(`/files/${fileId}`);

      // Cache the file metadata
      this.cache.set(fileId, file);
      console.log("[FileService] File retrieved:", fileId);

      return file;
    } catch (error) {
      console.error("[FileService] Failed to get file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 4. Update File
  async updateFile(fileId, updateData) {
    try {
      this.isLoading = true;
      console.log("[FileService] Updating file:", fileId);

      const apiClient = await this.getApiClient();
      const updatedFile = await apiClient.putMapleFile(
        `/files/${fileId}`,
        updateData,
      );

      // Update cache
      this.cache.set(fileId, updatedFile);
      console.log("[FileService] File updated:", fileId);

      return updatedFile;
    } catch (error) {
      console.error("[FileService] Failed to update file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 5. Archive File
  async archiveFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Archiving file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(`/files/${fileId}/archive`);

      // Update cache if we have the file
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        file.state = "archived";
        this.cache.set(fileId, file);
      }

      console.log("[FileService] File archived:", fileId);
      return result;
    } catch (error) {
      console.error("[FileService] Failed to archive file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 6. Restore File
  async restoreFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Restoring file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(`/files/${fileId}/restore`);

      // Update cache if we have the file
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        file.state = "active";
        this.cache.set(fileId, file);
      }

      console.log("[FileService] File restored:", fileId);
      return result;
    } catch (error) {
      console.error("[FileService] Failed to restore file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 7. Soft Delete File
  async deleteFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Soft deleting file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.deleteMapleFile(`/files/${fileId}`);

      // Update cache to mark as deleted
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        file.state = "deleted";
        this.cache.set(fileId, file);
      }

      console.log("[FileService] File soft deleted:", fileId);
      return result;
    } catch (error) {
      console.error("[FileService] Failed to delete file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 8. Delete Multiple Files
  async deleteMultipleFiles(fileIds) {
    try {
      this.isLoading = true;
      console.log("[FileService] Deleting multiple files:", fileIds.length);

      const apiClient = await this.getApiClient();
      const result = await apiClient.deleteMapleFile("/files/multiple", {
        body: JSON.stringify({ file_ids: fileIds }),
      });

      // Update cache for deleted files
      fileIds.forEach((fileId) => {
        if (this.cache.has(fileId)) {
          const file = this.cache.get(fileId);
          file.state = "deleted";
          this.cache.set(fileId, file);
        }
      });

      console.log("[FileService] Multiple files deleted:", result);
      return result;
    } catch (error) {
      console.error("[FileService] Failed to delete multiple files:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 9. Get Presigned Upload URL
  async getPresignedUploadUrl(fileId, urlDuration = null) {
    try {
      this.isLoading = true;
      console.log(
        "[FileService] Getting presigned upload URL for file:",
        fileId,
      );

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/upload-url`,
        requestData,
      );

      // Update cache with file metadata if returned
      if (response.file) {
        this.cache.set(fileId, response.file);
      }

      console.log("[FileService] Presigned upload URL generated");
      return response;
    } catch (error) {
      console.error("[FileService] Failed to get presigned upload URL:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 10. Get Presigned Download URL
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      this.isLoading = true;
      console.log(
        "[FileService] Getting presigned download URL for file:",
        fileId,
      );

      const requestData = {};
      if (urlDuration) {
        requestData.url_duration = urlDuration.toString();
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        `/files/${fileId}/download-url`,
        requestData,
      );

      // Update cache with file metadata if returned
      if (response.file) {
        this.cache.set(fileId, response.file);
      }

      console.log("[FileService] Presigned download URL generated");
      return response;
    } catch (error) {
      console.error(
        "[FileService] Failed to get presigned download URL:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 11. List Files by Collection
  async listFilesByCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Listing files in collection:", collectionId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections/${collectionId}/files`,
      );

      // Cache all files
      if (response.files) {
        response.files.forEach((file) => {
          this.cache.set(file.id, file);
        });
      }

      console.log(
        "[FileService] Files retrieved:",
        response.files?.length || 0,
      );
      return response.files || [];
    } catch (error) {
      console.error("[FileService] Failed to list files by collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 12. List File Sync Data
  async syncFiles(cursor = null, limit = 5000) {
    try {
      this.isLoading = true;
      console.log("[FileService] Syncing files", { cursor, limit });

      const apiClient = await this.getApiClient();
      const params = new URLSearchParams({ limit: limit.toString() });

      if (cursor) {
        params.append("cursor", cursor);
      }

      const response = await apiClient.getMapleFile(`/sync/files?${params}`);

      console.log("[FileService] Files synced:", {
        count: response.files?.length || 0,
        hasMore: response.has_more || false,
      });

      return response;
    } catch (error) {
      console.error("[FileService] Failed to sync files:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Upload file data directly to S3 using presigned URL with CORS handling
  async uploadFileToS3(presignedUrl, fileData, onProgress = null) {
    try {
      console.log("[FileService] Uploading file to S3");

      // For DigitalOcean Spaces, we need to handle CORS differently
      // Check if this is a CORS issue and provide better guidance
      const response = await fetch(presignedUrl, {
        method: "PUT",
        body: fileData,
        headers: {
          "Content-Type": "application/octet-stream",
        },
        // Remove credentials for CORS compatibility
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 upload failed with status: ${response.status}`);
      }

      console.log("[FileService] File uploaded to S3 successfully");
      return true;
    } catch (error) {
      console.error("[FileService] Failed to upload file to S3:", error);

      // Provide better error messages
      if (error.name === "TypeError" && error.message === "Failed to fetch") {
        throw new Error(
          "Unable to upload file to cloud storage. This is likely a CORS configuration issue. " +
            "Please ensure your S3/Spaces bucket has the correct CORS settings to allow uploads from " +
            window.location.origin,
        );
      }

      throw error;
    }
  }

  // Download file data directly from S3 using presigned URL
  async downloadFileFromS3(presignedUrl, onProgress = null) {
    try {
      console.log("[FileService] Downloading file from S3");

      const response = await fetch(presignedUrl, {
        method: "GET",
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 download failed with status: ${response.status}`);
      }

      const blob = await response.blob();
      console.log("[FileService] File downloaded from S3 successfully");

      return blob;
    } catch (error) {
      console.error("[FileService] Failed to download file from S3:", error);
      throw error;
    }
  }

  // Complete file upload workflow with better error handling
  async uploadFile(fileData, encryptedFileContent, encryptedThumbnail = null) {
    let fileId = null;
    let pendingFileCreated = false;

    try {
      // Step 1: Create pending file
      console.log("[FileService] Starting file upload workflow");
      const pendingResponse = await this.createPendingFile(fileData);

      fileId = pendingResponse.file.id;
      pendingFileCreated = true;
      const uploadUrl = pendingResponse.presigned_upload_url;
      const thumbnailUrl = pendingResponse.presigned_thumbnail_url;

      // Step 2: Upload encrypted file content to S3
      console.log("[FileService] Uploading encrypted file content");
      await this.uploadFileToS3(uploadUrl, encryptedFileContent);

      // Step 3: Upload thumbnail if provided
      if (encryptedThumbnail && thumbnailUrl) {
        console.log("[FileService] Uploading encrypted thumbnail");
        await this.uploadFileToS3(thumbnailUrl, encryptedThumbnail);
      }

      // Step 4: Complete the upload
      console.log("[FileService] Completing file upload");
      const completionData = {
        actual_file_size_in_bytes:
          encryptedFileContent.size || encryptedFileContent.length,
        upload_confirmed: true,
      };

      if (encryptedThumbnail) {
        completionData.actual_thumbnail_size_in_bytes =
          encryptedThumbnail.size || encryptedThumbnail.length;
        completionData.thumbnail_upload_confirmed = true;
      }

      const completeResponse = await this.completeFileUpload(
        fileId,
        completionData,
      );

      console.log("[FileService] File upload workflow completed successfully");
      return completeResponse.file;
    } catch (error) {
      console.error("[FileService] File upload workflow failed:", error);

      // Only try to clean up if we created a pending file and it's not a CORS issue
      // The backend doesn't allow deleting pending files, so we skip cleanup
      if (pendingFileCreated && fileId) {
        console.log(
          "[FileService] Pending file created but upload failed. File ID:",
          fileId,
        );
        console.log(
          "[FileService] Note: Pending files cannot be deleted and will be cleaned up by the backend",
        );

        // Remove from local cache and upload queue
        this.cache.delete(fileId);
        this.uploadQueue.delete(fileId);
      }

      throw error;
    }
  }

  // Complete file download workflow
  async downloadFile(fileId) {
    try {
      // Step 1: Get presigned download URL
      console.log("[FileService] Starting file download workflow");
      const downloadResponse = await this.getPresignedDownloadUrl(fileId);

      const fileMetadata = downloadResponse.file;
      const downloadUrl = downloadResponse.presigned_download_url;
      const thumbnailUrl = downloadResponse.presigned_thumbnail_url;

      // Step 2: Download encrypted file content from S3
      console.log("[FileService] Downloading encrypted file content");
      const encryptedContent = await this.downloadFileFromS3(downloadUrl);

      // Step 3: Download thumbnail if available
      let encryptedThumbnail = null;
      if (thumbnailUrl && fileMetadata.encrypted_thumbnail_size_in_bytes > 0) {
        console.log("[FileService] Downloading encrypted thumbnail");
        try {
          encryptedThumbnail = await this.downloadFileFromS3(thumbnailUrl);
        } catch (thumbError) {
          console.warn(
            "[FileService] Failed to download thumbnail:",
            thumbError,
          );
          // Thumbnail download failure is not critical
        }
      }

      console.log(
        "[FileService] File download workflow completed successfully",
      );
      return {
        metadata: fileMetadata,
        encryptedContent,
        encryptedThumbnail,
      };
    } catch (error) {
      console.error("[FileService] File download workflow failed:", error);
      throw error;
    }
  }

  // Sync all files for offline support
  async syncAllFiles() {
    let cursor = null;
    let hasMore = true;
    const allSyncedFiles = [];

    while (hasMore) {
      const syncResult = await this.syncFiles(cursor, 5000);
      allSyncedFiles.push(...(syncResult.files || []));

      cursor = syncResult.next_cursor;
      hasMore = syncResult.has_more;
    }

    console.log(`[FileService] Synced ${allSyncedFiles.length} files`);
    return allSyncedFiles;
  }

  // Batch operations
  async batchArchiveFiles(fileIds) {
    try {
      console.log("[FileService] Batch archiving files:", fileIds.length);

      const results = await Promise.all(
        fileIds.map((fileId) =>
          this.archiveFile(fileId).catch((error) => ({
            fileId,
            error: error.message,
            success: false,
          })),
        ),
      );

      const successful = results.filter((r) => r.success !== false).length;
      const failed = results.filter((r) => r.success === false);

      console.log(
        `[FileService] Batch archive completed: ${successful} successful, ${failed.length} failed`,
      );

      return {
        successful,
        failed,
        total: fileIds.length,
      };
    } catch (error) {
      console.error("[FileService] Batch archive failed:", error);
      throw error;
    }
  }

  async batchRestoreFiles(fileIds) {
    try {
      console.log("[FileService] Batch restoring files:", fileIds.length);

      const results = await Promise.all(
        fileIds.map((fileId) =>
          this.restoreFile(fileId).catch((error) => ({
            fileId,
            error: error.message,
            success: false,
          })),
        ),
      );

      const successful = results.filter((r) => r.success !== false).length;
      const failed = results.filter((r) => r.success === false);

      console.log(
        `[FileService] Batch restore completed: ${successful} successful, ${failed.length} failed`,
      );

      return {
        successful,
        failed,
        total: fileIds.length,
      };
    } catch (error) {
      console.error("[FileService] Batch restore failed:", error);
      throw error;
    }
  }

  // Get files by state
  async getFilesByState(collectionId, state = "active") {
    const files = await this.listFilesByCollection(collectionId);
    return files.filter((file) => file.state === state);
  }

  // Get file from cache without API call
  getCachedFile(fileId) {
    return this.cache.get(fileId) || null;
  }

  // Check if file is in upload queue
  isFileUploading(fileId) {
    return this.uploadQueue.has(fileId);
  }

  // Get upload queue info
  getUploadQueueInfo(fileId) {
    return this.uploadQueue.get(fileId) || null;
  }

  // Clear file cache
  clearCache() {
    this.cache.clear();
    this.uploadQueue.clear();
    console.log("[FileService] Cache cleared");
  }

  // Remove specific file from cache
  removeCachedFile(fileId) {
    this.cache.delete(fileId);
    this.uploadQueue.delete(fileId);
  }

  // Get cache size
  getCacheSize() {
    return {
      files: this.cache.size,
      uploadQueue: this.uploadQueue.size,
    };
  }

  // Check if service is loading
  isLoadingData() {
    return this.isLoading;
  }

  // Validate file metadata before upload
  validateFileMetadata(fileData) {
    const errors = [];

    if (!fileData.id) {
      errors.push("File ID is required");
    }

    if (!fileData.collection_id) {
      errors.push("Collection ID is required");
    }

    if (!fileData.encrypted_metadata) {
      errors.push("Encrypted metadata is required");
    }

    if (
      !fileData.encrypted_file_key ||
      !fileData.encrypted_file_key.ciphertext
    ) {
      errors.push("Encrypted file key is required");
    }

    if (!fileData.encryption_version) {
      errors.push("Encryption version is required");
    }

    if (!fileData.encrypted_hash) {
      errors.push("Encrypted hash is required");
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      cacheSize: this.cache.size,
      cachedFileIds: Array.from(this.cache.keys()),
      uploadQueueSize: this.uploadQueue.size,
      uploadingFileIds: Array.from(this.uploadQueue.keys()),
      isLoading: this.isLoading,
    };
  }
}

// Export singleton instance
export default new FileService();

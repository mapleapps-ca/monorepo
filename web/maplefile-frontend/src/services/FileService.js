// FileService for managing MapleFile files with end-to-end encryption
// Supports all file operations including upload, download, sharing, and synchronization

class FileService {
  async getCryptoServices() {
    if (!this._cryptoService) {
      const { default: CryptoService } = await import("./CryptoService.js");
      const { default: CollectionCryptoService } = await import(
        "./CollectionCryptoService.js"
      );
      this._cryptoService = CryptoService;
      this._collectionCryptoService = CollectionCryptoService;
    }
    return {
      CryptoService: this._cryptoService,
      CollectionCryptoService: this._collectionCryptoService,
    };
  }

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

  // Decrypt file metadata for display
  async decryptFileMetadata(file, collectionKey) {
    try {
      console.log(`[FileService] Decrypting file metadata for ${file.id}`);

      const { CryptoService, CollectionCryptoService } =
        await this.getCryptoServices();

      // Log the encrypted file key structure for debugging
      console.log(
        "[FileService] Encrypted file key structure:",
        file.encrypted_file_key,
      );

      // Handle the file key format - convert base64 strings to proper format
      let encryptedFileKey = file.encrypted_file_key;

      // Check if ciphertext and nonce are base64 strings and convert them
      if (typeof encryptedFileKey.ciphertext === "string") {
        console.log(
          "[FileService] Converting base64 file key components to Uint8Array",
        );

        try {
          const ciphertext = CryptoService.tryDecodeBase64(
            encryptedFileKey.ciphertext,
          );
          const nonce = CryptoService.tryDecodeBase64(encryptedFileKey.nonce);

          encryptedFileKey = {
            ciphertext: ciphertext,
            nonce: nonce,
            key_version: encryptedFileKey.key_version || 1,
          };

          console.log(
            "[FileService] Converted file key - ciphertext length:",
            ciphertext.length,
            "nonce length:",
            nonce.length,
          );
        } catch (conversionError) {
          console.error(
            "[FileService] Failed to convert file key format:",
            conversionError,
          );
          throw new Error(
            `File key conversion failed: ${conversionError.message}`,
          );
        }
      }

      // Decrypt the file key first
      console.log("[FileService] Decrypting file key with collection key...");
      const fileKey = await CryptoService.decryptFileKey(
        encryptedFileKey,
        collectionKey,
      );

      console.log(
        "[FileService] File key decrypted successfully, length:",
        fileKey.length,
      );

      // Decrypt the metadata
      console.log("[FileService] Decrypting file metadata...");
      const decryptedMetadataBytes = await CryptoService.decryptWithKey(
        file.encrypted_metadata,
        fileKey,
      );

      // Parse the metadata JSON
      const metadataString = new TextDecoder().decode(decryptedMetadataBytes);
      console.log("[FileService] Decrypted metadata string:", metadataString);

      const metadata = JSON.parse(metadataString);

      console.log(
        `[FileService] Successfully decrypted file ${file.id}: ${metadata.name}`,
      );

      return {
        ...file,
        name: metadata.name,
        mime_type: metadata.mime_type,
        size: metadata.size,
        _decrypted_metadata: metadata,
        _file_key: fileKey, // Store for potential future use
      };
    } catch (error) {
      console.error(`[FileService] Failed to decrypt file ${file.id}:`, error);
      return {
        ...file,
        name: "[Unable to decrypt]",
        mime_type: "unknown",
        size: file.encrypted_file_size_in_bytes || 0,
        _decrypt_error: error.message,
      };
    }
  }

  // Decrypt multiple files
  async decryptFiles(files, collectionKey) {
    if (!files || files.length === 0) return [];

    const decryptedFiles = await Promise.all(
      files.map((file) => this.decryptFileMetadata(file, collectionKey)),
    );

    return decryptedFiles;
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
  // 11. List Files by Collection
  async listFilesByCollection(collectionId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log(
        "[FileService] Listing files in collection:",
        collectionId,
        "forceRefresh:",
        forceRefresh,
      );

      // If force refresh, clear cache for this collection first
      if (forceRefresh) {
        this.invalidateCollectionFilesCache(collectionId);
      }

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections/${collectionId}/files`,
      );

      let files = response.files || [];

      // Add default state for files that don't have one
      files = files.map((file) => ({
        ...file,
        state: file.state || "active", // Default to active if no state
      }));

      // Try to decrypt files if we have collection crypto service available
      try {
        const { default: CollectionCryptoService } = await import(
          "./CollectionCryptoService.js"
        );

        // First try to get collection key from cache
        let collectionKey =
          CollectionCryptoService.getCachedCollectionKey(collectionId);

        if (!collectionKey) {
          console.log(
            "[FileService] No cached collection key found, trying to load collection...",
          );

          // Try to get the collection to ensure it's loaded and key is cached
          try {
            const { default: CollectionService } = await import(
              "./CollectionService.js"
            );
            const collection =
              await CollectionService.getCollection(collectionId);

            if (collection && collection.collection_key) {
              collectionKey = collection.collection_key;
              console.log(
                "[FileService] Got collection key from loaded collection",
              );
            }
          } catch (collectionError) {
            console.warn(
              "[FileService] Could not load collection:",
              collectionError,
            );
          }
        }

        if (collectionKey) {
          console.log(
            "[FileService] Decrypting",
            files.length,
            "files with collection key",
          );
          files = await this.decryptFiles(files, collectionKey);
          console.log("[FileService] File decryption completed");
        } else {
          console.warn(
            "[FileService] No collection key available - files will show as encrypted",
          );
          console.warn(
            "[FileService] Make sure collection is loaded first with proper password",
          );
        }
      } catch (decryptError) {
        console.error("[FileService] Could not decrypt files:", decryptError);
        // Continue with encrypted files
      }

      // Cache all files
      files.forEach((file) => {
        this.cache.set(file.id, file);
      });

      console.log("[FileService] Files retrieved and processed:", files.length);
      return files;
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

      // NEW: Clear cache for this collection so next listFilesByCollection call fetches fresh data
      this.invalidateCollectionFilesCache(fileData.collection_id);

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

  invalidateCollectionFilesCache(collectionId) {
    // Remove any cached files for this collection
    const filesToRemove = [];
    for (const [fileId, file] of this.cache.entries()) {
      if (file.collection_id === collectionId) {
        filesToRemove.push(fileId);
      }
    }

    filesToRemove.forEach((fileId) => {
      this.cache.delete(fileId);
    });

    console.log(
      `[FileService] Invalidated cache for collection ${collectionId}, removed ${filesToRemove.length} files`,
    );
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

  // Update or add a file to cache
  updateFileInCache(file) {
    this.cache.set(file.id, file);
    console.log(`[FileService] Updated file ${file.id} in cache`);
  }

  // Remove file from cache
  removeFileFromCache(fileId) {
    this.cache.delete(fileId);
    this.uploadQueue.delete(fileId);
    console.log(`[FileService] Removed file ${fileId} from cache`);
  }

  // Get all cached files for a collection
  getCachedFilesForCollection(collectionId) {
    const files = [];
    for (const file of this.cache.values()) {
      if (file.collection_id === collectionId) {
        files.push(file);
      }
    }
    return files;
  }

  // Complete file download workflow with decryption
  async downloadAndDecryptFile(fileId) {
    try {
      console.log(
        "[FileService] Starting complete download and decryption for:",
        fileId,
      );

      // Step 1: Get the file metadata (should be cached and decrypted)
      const fileMetadata = this.getCachedFile(fileId);
      if (!fileMetadata) {
        throw new Error(
          "File metadata not found. Please refresh the file list.",
        );
      }

      if (!fileMetadata._file_key) {
        throw new Error(
          "File key not available. File may not be properly decrypted.",
        );
      }

      console.log("[FileService] File metadata found:", fileMetadata.name);

      // Step 2: Get presigned download URL and download encrypted content
      const downloadResponse = await this.downloadFile(fileId);
      const encryptedContent = downloadResponse.encryptedContent;

      console.log(
        "[FileService] Encrypted content downloaded, size:",
        encryptedContent.size,
      );

      // Step 3: Convert blob to array buffer
      const encryptedArrayBuffer = await encryptedContent.arrayBuffer();
      const encryptedBytes = new Uint8Array(encryptedArrayBuffer);

      console.log(
        "[FileService] Converting to bytes, length:",
        encryptedBytes.length,
      );

      // Step 4: Import crypto service for decryption
      const { default: CryptoService } = await import("./CryptoService.js");
      await CryptoService.initialize();

      // Step 5: Decrypt the file content
      console.log("[FileService] Decrypting file content...");
      const decryptedBytes = await CryptoService.decryptWithKey(
        CryptoService.uint8ArrayToBase64(encryptedBytes), // Convert to base64 for decryption
        fileMetadata._file_key,
      );

      console.log(
        "[FileService] File decrypted successfully, size:",
        decryptedBytes.length,
      );

      // Step 6: Create blob with proper MIME type
      const mimeType = fileMetadata.mime_type || "application/octet-stream";
      const decryptedBlob = new Blob([decryptedBytes], { type: mimeType });

      // Step 7: Get the original filename
      const filename =
        fileMetadata.name || `downloaded_file_${fileId.substring(0, 8)}`;

      console.log(
        "[FileService] Download prepared:",
        filename,
        "size:",
        decryptedBlob.size,
      );

      return {
        blob: decryptedBlob,
        filename: filename,
        mimeType: mimeType,
        size: decryptedBlob.size,
      };
    } catch (error) {
      console.error(
        "[FileService] Failed to download and decrypt file:",
        error,
      );
      throw error;
    }
  }

  // Trigger browser download
  downloadBlobAsFile(blob, filename) {
    try {
      // Create object URL
      const url = URL.createObjectURL(blob);

      // Create temporary download link
      const downloadLink = document.createElement("a");
      downloadLink.href = url;
      downloadLink.download = filename;
      downloadLink.style.display = "none";

      // Add to document, click, and remove
      document.body.appendChild(downloadLink);
      downloadLink.click();
      document.body.removeChild(downloadLink);

      // Clean up object URL
      setTimeout(() => {
        URL.revokeObjectURL(url);
      }, 1000);

      console.log("[FileService] Browser download triggered for:", filename);
    } catch (error) {
      console.error("[FileService] Failed to trigger browser download:", error);
      throw error;
    }
  }

  // Combined download and save function
  async downloadAndSaveFile(fileId) {
    try {
      console.log("[FileService] Starting download and save for file:", fileId);

      const downloadResult = await this.downloadAndDecryptFile(fileId);

      // Trigger browser download
      this.downloadBlobAsFile(downloadResult.blob, downloadResult.filename);

      console.log("[FileService] File download completed successfully");
      return downloadResult;
    } catch (error) {
      console.error("[FileService] Download and save failed:", error);
      throw error;
    }
  }
}

// Export singleton instance
export default new FileService();

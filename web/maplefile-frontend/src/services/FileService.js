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

    // File states constants
    this.FILE_STATES = {
      PENDING: "pending",
      ACTIVE: "active",
      DELETED: "deleted",
      ARCHIVED: "archived",
    };
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("./ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // Normalize file object with new fields
  normalizeFile(file) {
    return {
      ...file,
      // Ensure version is present (default to 1 for new files)
      version: file.version || 1,
      // Ensure state is present (default to active)
      state: file.state || this.FILE_STATES.ACTIVE,
      // Ensure tombstone fields are present
      tombstone_version: file.tombstone_version || 0,
      tombstone_expiry: file.tombstone_expiry || "0001-01-01T00:00:00Z",
      // Add computed properties for easier checking
      _is_deleted: file.state === this.FILE_STATES.DELETED,
      _is_archived: file.state === this.FILE_STATES.ARCHIVED,
      _is_pending: file.state === this.FILE_STATES.PENDING,
      _is_active: file.state === this.FILE_STATES.ACTIVE,
      _has_tombstone: (file.tombstone_version || 0) > 0,
      _tombstone_expired: file.tombstone_expiry
        ? new Date(file.tombstone_expiry) < new Date() &&
          file.tombstone_expiry !== "0001-01-01T00:00:00Z"
        : false,
    };
  }

  // Decrypt file metadata for display
  async decryptFileMetadata(file, collectionKey) {
    try {
      console.log(`[FileService] === Decrypting file ${file.id} ===`);
      console.log("[FileService] Collection key length:", collectionKey.length);
      console.log("[FileService] File version:", file.version);
      console.log("[FileService] File state:", file.state);
      console.log(
        "[FileService] File key structure:",
        JSON.stringify(file.encrypted_file_key),
      );

      const { CryptoService, CollectionCryptoService } =
        await this.getCryptoServices();

      // Step 1: Decrypt the file key
      console.log("[FileService] Step 1: Decrypting file key...");
      const fileKey = await CryptoService.decryptFileKey(
        file.encrypted_file_key,
        collectionKey,
      );

      console.log("[FileService] File key decrypted, length:", fileKey.length);

      // Step 2: Decrypt the metadata
      console.log("[FileService] Step 2: Decrypting metadata...");
      console.log(
        "[FileService] Encrypted metadata length:",
        file.encrypted_metadata.length,
      );

      const decryptedMetadataBytes = await CryptoService.decryptWithKey(
        file.encrypted_metadata,
        fileKey,
      );

      console.log(
        "[FileService] Metadata decrypted, byte length:",
        decryptedMetadataBytes.length,
      );

      // Step 3: Parse the metadata JSON
      const metadataString = new TextDecoder().decode(decryptedMetadataBytes);
      console.log("[FileService] Metadata string:", metadataString);

      const metadata = JSON.parse(metadataString);
      console.log("[FileService] Parsed metadata:", metadata);

      const result = this.normalizeFile({
        ...file,
        name: metadata.name,
        mime_type: metadata.mime_type,
        size: metadata.size,
        _decrypted_metadata: metadata,
        _file_key: fileKey,
        _collection_key_length: collectionKey.length, // Debug info
      });

      console.log(
        `[FileService] ✅ Successfully decrypted file: ${metadata.name} (v${result.version}, ${result.state})`,
      );
      return result;
    } catch (error) {
      console.error(
        `[FileService] ❌ Failed to decrypt file ${file.id}:`,
        error,
      );
      console.error("[FileService] Error details:", {
        message: error.message,
        stack: error.stack,
        fileKeyStructure: file.encrypted_file_key,
        collectionKeyLength: collectionKey?.length,
      });

      return this.normalizeFile({
        ...file,
        name: "[Unable to decrypt]",
        mime_type: "unknown",
        size: file.encrypted_file_size_in_bytes || 0,
        _decrypt_error: error.message,
        _debug_info: {
          collectionKeyAvailable: !!collectionKey,
          fileKeyStructure: JSON.stringify(file.encrypted_file_key),
          errorMessage: error.message,
        },
      });
    }
  }

  // Decrypt multiple files
  async decryptFiles(files, collectionKey) {
    if (!files || files.length === 0) return [];

    const decryptedFiles = [];
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      try {
        console.log(
          `[FileService] Decrypting file ${i + 1}/${files.length}: ${file.id} (v${file.version || "unknown"}, ${file.state || "unknown"})`,
        );
        const decryptedFile = await this.decryptFileMetadata(
          file,
          collectionKey,
        );
        decryptedFiles.push(decryptedFile);
      } catch (fileError) {
        console.error(
          `[FileService] Failed to decrypt file ${file.id}:`,
          fileError.message,
        );
        // Add the file with error info
        decryptedFiles.push(
          this.normalizeFile({
            ...file,
            name: `[Decrypt failed: ${fileError.message.substring(0, 50)}...]`,
            _decrypt_error: fileError.message,
          }),
        );
      }
    }

    return decryptedFiles;
  }

  // 1. Create Pending File
  async createPendingFile(fileData) {
    try {
      this.isLoading = true;
      console.log("[FileService] Creating pending file");

      // Ensure the file data includes the new fields
      const normalizedFileData = {
        ...fileData,
        version: fileData.version || 1,
        state: this.FILE_STATES.PENDING, // Always pending when creating
      };

      const apiClient = await this.getApiClient();
      const response = await apiClient.postMapleFile(
        "/files/pending",
        normalizedFileData,
      );

      // Cache the pending file metadata with normalization
      if (response.file) {
        const normalizedFile = this.normalizeFile(response.file);
        this.cache.set(normalizedFile.id, normalizedFile);

        // Track in upload queue
        this.uploadQueue.set(normalizedFile.id, {
          file: normalizedFile,
          uploadUrl: response.presigned_upload_url,
          thumbnailUrl: response.presigned_thumbnail_url,
          expirationTime: response.upload_url_expiration_time,
        });
      }

      console.log(
        "[FileService] Pending file created:",
        response.file?.id,
        "version:",
        response.file?.version,
      );
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
        const normalizedFile = this.normalizeFile(response.file);
        this.cache.set(fileId, normalizedFile);
      }

      // Remove from upload queue
      this.uploadQueue.delete(fileId);

      console.log(
        "[FileService] File upload completed:",
        fileId,
        "new state:",
        response.file?.state,
        "version:",
        response.file?.version,
      );
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

      // Cache the file metadata with normalization
      const normalizedFile = this.normalizeFile(file);
      this.cache.set(fileId, normalizedFile);
      console.log(
        "[FileService] File retrieved:",
        fileId,
        "version:",
        normalizedFile.version,
        "state:",
        normalizedFile.state,
      );

      return normalizedFile;
    } catch (error) {
      console.error("[FileService] Failed to get file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 4. Update File with version support for optimistic locking
  async updateFile(fileId, updateData) {
    try {
      this.isLoading = true;
      console.log(
        "[FileService] Updating file:",
        fileId,
        "with data:",
        updateData,
      );

      // Get current file to check version
      const currentFile = await this.getFile(fileId);
      if (!currentFile) {
        throw new Error("File not found for update");
      }

      // Include current version for optimistic locking if not provided
      const updatePayload = {
        ...updateData,
        version: updateData.version || currentFile.version,
        id: fileId, // Ensure ID is included
      };

      console.log(
        "[FileService] Update payload includes version:",
        updatePayload.version,
      );

      const apiClient = await this.getApiClient();
      const updatedFile = await apiClient.putMapleFile(
        `/files/${fileId}`,
        updatePayload,
      );

      // Update cache with normalized file
      const normalizedFile = this.normalizeFile(updatedFile);
      this.cache.set(fileId, normalizedFile);
      console.log(
        "[FileService] File updated:",
        fileId,
        "new version:",
        normalizedFile.version,
        "state:",
        normalizedFile.state,
      );

      return normalizedFile;
    } catch (error) {
      console.error("[FileService] Failed to update file:", error);

      // Check for version conflict
      if (
        error.message?.includes("version") ||
        error.message?.includes("conflict")
      ) {
        throw new Error(
          "File has been updated by another process. Please refresh and try again.",
        );
      }

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 5. Archive File (soft archive)
  async archiveFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Archiving file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(`/files/${fileId}/archive`);

      // Update cache if we have the file
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        const updatedFile = this.normalizeFile({
          ...file,
          state: this.FILE_STATES.ARCHIVED,
          version: (file.version || 1) + 1, // Increment version on state change
        });
        this.cache.set(fileId, updatedFile);
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

  // 6. Restore File (unarchive)
  async restoreFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Restoring file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(`/files/${fileId}/restore`);

      // Update cache if we have the file
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        const updatedFile = this.normalizeFile({
          ...file,
          state: this.FILE_STATES.ACTIVE,
          version: (file.version || 1) + 1, // Increment version on state change
          tombstone_version: 0, // Clear tombstone on restore
          tombstone_expiry: "0001-01-01T00:00:00Z",
        });
        this.cache.set(fileId, updatedFile);
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

  // 7. Soft Delete File (creates tombstone)
  async deleteFile(fileId) {
    try {
      this.isLoading = true;
      console.log("[FileService] Soft deleting file:", fileId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.deleteMapleFile(`/files/${fileId}`);

      // Update cache to mark as deleted with tombstone
      if (this.cache.has(fileId)) {
        const file = this.cache.get(fileId);
        const newVersion = (file.version || 1) + 1;
        const updatedFile = this.normalizeFile({
          ...file,
          state: this.FILE_STATES.DELETED,
          version: newVersion,
          tombstone_version: newVersion, // Set tombstone version
          tombstone_expiry:
            result.tombstone_expiry ||
            new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // Default 30 days
        });
        this.cache.set(fileId, updatedFile);
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
          const newVersion = (file.version || 1) + 1;
          const updatedFile = this.normalizeFile({
            ...file,
            state: this.FILE_STATES.DELETED,
            version: newVersion,
            tombstone_version: newVersion,
            tombstone_expiry: new Date(
              Date.now() + 30 * 24 * 60 * 60 * 1000,
            ).toISOString(),
          });
          this.cache.set(fileId, updatedFile);
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
        const normalizedFile = this.normalizeFile(response.file);
        this.cache.set(fileId, normalizedFile);
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
        const normalizedFile = this.normalizeFile(response.file);
        this.cache.set(fileId, normalizedFile);
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

  // 11. List Files by Collection with state filtering
  async listFilesByCollection(
    collectionId,
    forceRefresh = false,
    includeStates = null,
  ) {
    try {
      this.isLoading = true;
      console.log(
        "[FileService] Listing files in collection:",
        collectionId,
        "forceRefresh:",
        forceRefresh,
        "includeStates:",
        includeStates,
      );

      // If force refresh, clear cache for this collection first
      if (forceRefresh) {
        this.invalidateCollectionFilesCache(collectionId);
      }

      const apiClient = await this.getApiClient();
      let url = `/collections/${collectionId}/files`;

      // Add state filtering if specified
      if (includeStates && Array.isArray(includeStates)) {
        const params = new URLSearchParams();
        includeStates.forEach((state) => params.append("states", state));
        url += `?${params.toString()}`;
      }

      const response = await apiClient.getMapleFile(url);

      let files = response.files || [];

      // Normalize all files with new fields
      files = files.map((file) => this.normalizeFile(file));

      console.log(`[FileService] Found ${files.length} files:`, {
        pending: files.filter((f) => f._is_pending).length,
        active: files.filter((f) => f._is_active).length,
        archived: files.filter((f) => f._is_archived).length,
        deleted: files.filter((f) => f._is_deleted).length,
        withTombstones: files.filter((f) => f._has_tombstone).length,
      });

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

  // 12. List File Sync Data with version handling
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

      // Normalize sync response files
      if (response.files) {
        response.files = response.files.map((file) => this.normalizeFile(file));
      }

      console.log("[FileService] Files synced:", {
        count: response.files?.length || 0,
        hasMore: response.has_more || false,
        nextCursor: response.next_cursor || null,
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

      console.log(
        "[FileService] Pending file created with version:",
        pendingResponse.file.version,
      );

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

      console.log(
        "[FileService] File upload workflow completed successfully. Final version:",
        completeResponse.file.version,
        "state:",
        completeResponse.file.state,
      );

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

  // Invalidate cached files for a specific collection
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

      console.log(
        "[FileService] Downloading file version:",
        fileMetadata.version,
        "state:",
        fileMetadata.state,
      );

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
        metadata: this.normalizeFile(fileMetadata),
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

  // Batch operations with version handling
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

  // Get files by state with enhanced filtering
  getFilesByState(collectionId, state = this.FILE_STATES.ACTIVE) {
    const files = [];
    for (const file of this.cache.values()) {
      if (file.collection_id === collectionId && file.state === state) {
        files.push(file);
      }
    }
    return files;
  }

  // Get files by multiple states
  getFilesByStates(collectionId, states = [this.FILE_STATES.ACTIVE]) {
    const files = [];
    for (const file of this.cache.values()) {
      if (file.collection_id === collectionId && states.includes(file.state)) {
        files.push(file);
      }
    }
    return files;
  }

  // Get files with tombstones (deleted files)
  getTombstoneFiles(collectionId) {
    const files = [];
    for (const file of this.cache.values()) {
      if (file.collection_id === collectionId && file._has_tombstone) {
        files.push(file);
      }
    }
    return files;
  }

  // Get expired tombstone files that can be permanently deleted
  getExpiredTombstoneFiles(collectionId) {
    const files = [];
    const now = new Date();
    for (const file of this.cache.values()) {
      if (file.collection_id === collectionId && file._tombstone_expired) {
        files.push(file);
      }
    }
    return files;
  }

  // Check if file can be restored (has tombstone but not expired)
  canRestoreFile(file) {
    return file._has_tombstone && !file._tombstone_expired && file._is_deleted;
  }

  // Check if file can be permanently deleted
  canPermanentlyDeleteFile(file) {
    return file._tombstone_expired || (file._has_tombstone && file._is_deleted);
  }

  // Get file version history (if implemented by backend)
  async getFileVersionHistory(fileId) {
    try {
      console.log("[FileService] Getting version history for file:", fileId);

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/files/${fileId}/versions`,
      );

      // Normalize version history
      if (response.versions) {
        response.versions = response.versions.map((version) =>
          this.normalizeFile(version),
        );
      }

      return response;
    } catch (error) {
      console.error("[FileService] Failed to get file version history:", error);
      throw error;
    }
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

  // Validate file metadata before upload with version support
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

    // Version should be a positive number if provided
    if (
      fileData.version !== undefined &&
      (typeof fileData.version !== "number" || fileData.version < 1)
    ) {
      errors.push("Version must be a positive number");
    }

    // State should be valid if provided
    if (
      fileData.state !== undefined &&
      !Object.values(this.FILE_STATES).includes(fileData.state)
    ) {
      errors.push(
        `State must be one of: ${Object.values(this.FILE_STATES).join(", ")}`,
      );
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  // Get debug information
  getDebugInfo() {
    const cacheStats = {
      total: this.cache.size,
      byState: {},
      withTombstones: 0,
      expiredTombstones: 0,
    };

    // Calculate cache statistics
    for (const file of this.cache.values()) {
      cacheStats.byState[file.state] =
        (cacheStats.byState[file.state] || 0) + 1;
      if (file._has_tombstone) cacheStats.withTombstones++;
      if (file._tombstone_expired) cacheStats.expiredTombstones++;
    }

    return {
      cacheSize: this.cache.size,
      cacheStats,
      cachedFileIds: Array.from(this.cache.keys()),
      uploadQueueSize: this.uploadQueue.size,
      uploadingFileIds: Array.from(this.uploadQueue.keys()),
      isLoading: this.isLoading,
      fileStates: this.FILE_STATES,
    };
  }

  // Update or add a file to cache
  updateFileInCache(file) {
    const normalizedFile = this.normalizeFile(file);
    this.cache.set(normalizedFile.id, normalizedFile);
    console.log(
      `[FileService] Updated file ${normalizedFile.id} in cache (v${normalizedFile.version}, ${normalizedFile.state})`,
    );
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

      // Check if file can be downloaded (not deleted or expired)
      if (fileMetadata._is_deleted && !this.canRestoreFile(fileMetadata)) {
        throw new Error("File is deleted and cannot be downloaded.");
      }

      console.log(
        "[FileService] File metadata found:",
        fileMetadata.name,
        "version:",
        fileMetadata.version,
        "state:",
        fileMetadata.state,
      );

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
        version: fileMetadata.version,
        state: fileMetadata.state,
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

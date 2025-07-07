// File: monorepo/web/maplefile-frontend/src/services/Manager/File/CreateFileManager.js
// Create File Manager - Orchestrates API, Storage, and Crypto services for file creation

import CreateFileAPIService from "../../API/File/CreateFileAPIService.js";
import CreateFileStorageService from "../../Storage/File/CreateFileStorageService.js";

class CreateFileManager {
  constructor(authManager) {
    // CreateFileManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new CreateFileAPIService(authManager);
    this.storageService = new CreateFileStorageService();

    // Event listeners for file creation events
    this.fileCreationListeners = new Set();

    console.log(
      "[CreateFileManager] File manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[CreateFileManager] Initializing file manager...");

      // Initialize crypto service
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      await CryptoService.initialize();
      this.cryptoService = CryptoService;

      // Initialize collection crypto service
      const { default: CollectionCryptoService } = await import(
        "../../Crypto/CollectionCryptoService.js"
      );
      await CollectionCryptoService.initialize();
      this.collectionCryptoService = CollectionCryptoService;

      console.log("[CreateFileManager] File manager initialized successfully");
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to initialize file manager:",
        error,
      );
    }
  }

  // === S3 Upload Methods ===

  // Upload file data directly to S3 using presigned URL
  async uploadFileToS3(presignedUrl, fileData, onProgress = null) {
    try {
      console.log("[CreateFileManager] Uploading file to S3");

      const response = await fetch(presignedUrl, {
        method: "PUT",
        body: fileData,
        headers: {
          "Content-Type": "application/octet-stream",
        },
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 upload failed with status: ${response.status}`);
      }

      console.log("[CreateFileManager] File uploaded to S3 successfully");
      return true;
    } catch (error) {
      console.error("[CreateFileManager] Failed to upload file to S3:", error);

      // Provide better error messages for CORS issues
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

  // Complete file upload by notifying backend
  async completeFileUpload(fileId, completionData = {}) {
    try {
      console.log("[CreateFileManager] Completing file upload:", fileId);

      // Get API client
      const { default: ApiClient } = await import("../../API/ApiClient.js");

      const response = await ApiClient.postMapleFile(
        `/files/${fileId}/complete`,
        completionData,
      );

      console.log(
        "[CreateFileManager] File upload completed successfully:",
        fileId,
        "new state:",
        response.file?.state,
        "version:",
        response.file?.version,
      );

      return response;
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to complete file upload:",
        error,
      );
      throw error;
    }
  }

  // === Complete File Upload Workflow ===

  // Complete file upload workflow with all steps
  async uploadFileComplete(
    fileData,
    encryptedFileContent,
    encryptedThumbnail = null,
  ) {
    let fileId = null;
    let pendingFileCreated = false;

    try {
      // Step 1: Create pending file
      console.log("[CreateFileManager] Starting complete file upload workflow");
      const pendingResponse = await this.apiService.createPendingFile(fileData);

      fileId = pendingResponse.file.id;
      pendingFileCreated = true;
      const uploadUrl = pendingResponse.presigned_upload_url;
      const thumbnailUrl = pendingResponse.presigned_thumbnail_url;

      console.log(
        "[CreateFileManager] Pending file created with version:",
        pendingResponse.file.version,
      );

      // Store pending file info
      this.storageService.storePendingFile({
        file: pendingResponse.file,
        presigned_upload_url: pendingResponse.presigned_upload_url,
        presigned_thumbnail_url: pendingResponse.presigned_thumbnail_url,
        upload_url_expiration_time: pendingResponse.upload_url_expiration_time,
        encrypted_content: encryptedFileContent,
        encrypted_thumbnail: encryptedThumbnail,
      });

      // Add to upload queue with "uploading" status
      this.storageService.addToUploadQueue(fileId, {
        status: "uploading",
        upload_url: pendingResponse.presigned_upload_url,
        thumbnail_url: pendingResponse.presigned_thumbnail_url,
        expiration_time: pendingResponse.upload_url_expiration_time,
      });

      // Notify listeners of upload start
      this.notifyFileCreationListeners("file_upload_started", {
        fileId,
        name: fileData.name || "Unknown",
        uploadUrl: pendingResponse.presigned_upload_url,
      });

      // Step 2: Upload encrypted file content to S3
      console.log("[CreateFileManager] Uploading encrypted file content to S3");
      await this.uploadFileToS3(uploadUrl, encryptedFileContent);

      // Update upload queue status
      this.storageService.addToUploadQueue(fileId, {
        status: "uploading_thumbnail",
        upload_url: pendingResponse.presigned_upload_url,
        thumbnail_url: pendingResponse.presigned_thumbnail_url,
        expiration_time: pendingResponse.upload_url_expiration_time,
      });

      // Step 3: Upload thumbnail if provided
      if (encryptedThumbnail && thumbnailUrl) {
        console.log("[CreateFileManager] Uploading encrypted thumbnail to S3");
        await this.uploadFileToS3(thumbnailUrl, encryptedThumbnail);
      }

      // Update upload queue status
      this.storageService.addToUploadQueue(fileId, {
        status: "completing",
        upload_url: pendingResponse.presigned_upload_url,
        thumbnail_url: pendingResponse.presigned_thumbnail_url,
        expiration_time: pendingResponse.upload_url_expiration_time,
      });

      // Step 4: Complete the upload
      console.log("[CreateFileManager] Completing file upload");
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

      // Update storage with completed file
      this.storageService.updatePendingFile(fileId, {
        file: completeResponse.file,
        upload_completed: true,
        completed_at: new Date().toISOString(),
      });

      // Remove from upload queue (upload completed)
      this.storageService.removeFromUploadQueue(fileId);

      // Notify listeners of successful completion
      this.notifyFileCreationListeners("file_upload_completed", {
        fileId,
        file: completeResponse.file,
        name: fileData.name || "Unknown",
      });

      console.log(
        "[CreateFileManager] Complete file upload workflow finished successfully. Final version:",
        completeResponse.file.version,
        "state:",
        completeResponse.file.state,
      );

      return {
        file: completeResponse.file,
        fileId,
        success: true,
        uploadCompleted: true,
      };
    } catch (error) {
      console.error(
        "[CreateFileManager] Complete file upload workflow failed:",
        error,
      );

      // Update upload queue with error status
      if (fileId) {
        this.storageService.addToUploadQueue(fileId, {
          status: "error",
          error: error.message,
          failed_at: new Date().toISOString(),
        });
      }

      // Notify listeners of failure
      this.notifyFileCreationListeners("file_upload_failed", {
        fileId,
        error: error.message,
        metadata: fileData,
      });

      // Note: We don't try to clean up the pending file as the backend
      // doesn't allow deleting pending files and will clean them up automatically
      if (pendingFileCreated && fileId) {
        console.log(
          "[CreateFileManager] Pending file created but upload failed. File ID:",
          fileId,
        );
        console.log(
          "[CreateFileManager] Note: Pending files cannot be deleted and will be cleaned up by the backend",
        );
      }

      throw error;
    }
  }

  // === Original Pending File Creation (for backwards compatibility) ===

  // Create pending file with full E2EE encryption (original method)
  async createPendingFile(
    fileContent,
    collectionId,
    metadata,
    password = null,
  ) {
    try {
      this.isLoading = true;
      console.log("[CreateFileManager] Starting pending file creation");
      console.log("[CreateFileManager] File metadata:", {
        name: metadata.name,
        size: metadata.size,
        type: metadata.mime_type,
        collectionId: collectionId,
      });

      // Validate input
      if (!fileContent) {
        throw new Error("File content is required");
      }
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }
      if (!metadata.name) {
        throw new Error("File name is required");
      }

      // Get password for encryption
      const userPassword = password || (await this.getUserPassword());
      if (!userPassword) {
        throw new Error("Password required for file encryption");
      }

      console.log("[CreateFileManager] Generating file encryption key");

      // Step 1: Generate file encryption key
      const fileKey = this.cryptoService.generateRandomKey();
      console.log(
        "[CreateFileManager] File key generated, length:",
        fileKey.length,
      );

      // Step 2: Convert file content to Uint8Array if needed
      let fileBytes;
      if (fileContent instanceof ArrayBuffer) {
        fileBytes = new Uint8Array(fileContent);
      } else if (fileContent instanceof Uint8Array) {
        fileBytes = fileContent;
      } else {
        throw new Error("File content must be ArrayBuffer or Uint8Array");
      }

      console.log("[CreateFileManager] Encrypting file content");

      // Step 3: Encrypt file content
      const encryptedContent = await this.cryptoService.encryptWithKey(
        fileBytes,
        fileKey,
      );
      console.log("[CreateFileManager] File content encrypted");

      // Step 4: Generate file hash
      const fileHash = await this.cryptoService.hashData(fileBytes);
      const encryptedHash = this.cryptoService.uint8ArrayToBase64(fileHash);
      console.log("[CreateFileManager] File hash generated");

      // Step 5: Get collection key
      console.log("[CreateFileManager] Getting collection key");
      let collectionKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey) {
        console.log(
          "[CreateFileManager] No cached collection key, loading collection",
        );
        // Try to get the collection to ensure it's loaded and key is cached
        const { default: LocalStorageService } = await import(
          "../../Storage/LocalStorageService.js"
        );
        const userKeys = await this.collectionCryptoService.getUserKeys();

        // We might need to load the collection first to get its key
        throw new Error(
          "Collection key not found. Please ensure the collection is loaded first.",
        );
      }

      console.log(
        "[CreateFileManager] Encrypting file key with collection key",
      );

      // Step 6: Encrypt file key with collection key
      const encryptedFileKeyData = await this.cryptoService.encryptFileKey(
        fileKey,
        collectionKey,
      );

      // Step 7: Prepare metadata
      const fullMetadata = {
        name: metadata.name,
        mime_type: metadata.mime_type || "application/octet-stream",
        size: metadata.size || fileBytes.length,
        created_at: new Date().toISOString(),
        uploaded_at: new Date().toISOString(),
        ...metadata, // Include any additional metadata
      };

      console.log("[CreateFileManager] Encrypting metadata");

      // Step 8: Encrypt metadata
      const encryptedMetadata = await this.cryptoService.encryptWithKey(
        JSON.stringify(fullMetadata),
        fileKey,
      );

      // Step 9: Convert encrypted content for size calculation
      const encryptedBytes =
        this.cryptoService.tryDecodeBase64(encryptedContent);

      // Step 10: Generate file ID
      const fileId = this.cryptoService.generateUUID();

      // Step 11: Prepare API data
      const apiData = {
        id: fileId,
        collection_id: collectionId,
        encrypted_metadata: encryptedMetadata,
        encrypted_file_key: {
          ciphertext: this.cryptoService.uint8ArrayToBase64(
            encryptedFileKeyData.ciphertext,
          ),
          nonce: this.cryptoService.uint8ArrayToBase64(
            encryptedFileKeyData.nonce,
          ),
          key_version: encryptedFileKeyData.key_version || 1,
        },
        encryption_version: "v1.0",
        encrypted_hash: encryptedHash,
        expected_file_size_in_bytes: encryptedBytes.length,
      };

      // Add thumbnail size if provided
      if (metadata.expected_thumbnail_size_in_bytes) {
        apiData.expected_thumbnail_size_in_bytes =
          metadata.expected_thumbnail_size_in_bytes;
      }

      console.log("[CreateFileManager] Sending to API");

      // Step 12: Create pending file via API
      const response = await this.apiService.createPendingFile(apiData);

      console.log(
        "[CreateFileManager] Pending file created via API successfully",
      );

      // Store pending file info and upload URLs
      this.storageService.storePendingFile({
        file: response.file,
        presigned_upload_url: response.presigned_upload_url,
        presigned_thumbnail_url: response.presigned_thumbnail_url,
        upload_url_expiration_time: response.upload_url_expiration_time,
        encrypted_content_base64: encryptedContent,
        file_key: fileKey, // Store for later use
      });

      // Add to upload queue
      this.storageService.addToUploadQueue(fileId, {
        status: "pending",
        upload_url: response.presigned_upload_url,
        thumbnail_url: response.presigned_thumbnail_url,
        expiration_time: response.upload_url_expiration_time,
      });

      // Notify listeners
      this.notifyFileCreationListeners("pending_file_created", {
        fileId,
        name: metadata.name,
        collectionId,
        uploadUrl: response.presigned_upload_url,
      });

      console.log(
        "[CreateFileManager] Pending file created successfully:",
        fileId,
      );

      return {
        file: response.file,
        fileId,
        uploadUrl: response.presigned_upload_url,
        thumbnailUrl: response.presigned_thumbnail_url,
        uploadUrlExpirationTime: response.upload_url_expiration_time,
        encryptedContent: encryptedBytes,
        success: true,
      };
    } catch (error) {
      console.error("[CreateFileManager] Pending file creation failed:", error);

      // Notify listeners of failure
      this.notifyFileCreationListeners("file_creation_failed", {
        error: error.message,
        metadata,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Convenience Methods ===

  // Create pending file from File object with complete upload
  async createAndUploadFileFromFile(file, collectionId, password = null) {
    try {
      console.log(
        "[CreateFileManager] Creating and uploading file:",
        file.name,
      );

      // Read file content
      const fileContent = await this.readFileAsArrayBuffer(file);

      // Prepare metadata
      const metadata = {
        name: file.name,
        mime_type: file.type || "application/octet-stream",
        size: file.size,
        last_modified: file.lastModified
          ? new Date(file.lastModified).toISOString()
          : undefined,
      };

      // Encrypt the file content first
      const userPassword = password || (await this.getUserPassword());
      if (!userPassword) {
        throw new Error("Password required for file encryption");
      }

      // Generate file encryption key
      const fileKey = this.cryptoService.generateRandomKey();

      // Convert file content to Uint8Array
      const fileBytes = new Uint8Array(fileContent);

      // Encrypt file content
      const encryptedContentBase64 = await this.cryptoService.encryptWithKey(
        fileBytes,
        fileKey,
      );

      // Convert to blob for S3 upload
      const encryptedBytes = this.cryptoService.tryDecodeBase64(
        encryptedContentBase64,
      );
      const encryptedBlob = new Blob([encryptedBytes], {
        type: "application/octet-stream",
      });

      // Generate file hash
      const fileHash = await this.cryptoService.hashData(fileBytes);
      const encryptedHash = this.cryptoService.uint8ArrayToBase64(fileHash);

      // Get collection key
      let collectionKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);
      if (!collectionKey) {
        throw new Error(
          "Collection key not found. Please ensure the collection is loaded first.",
        );
      }

      // Encrypt file key with collection key
      const encryptedFileKeyData = await this.cryptoService.encryptFileKey(
        fileKey,
        collectionKey,
      );

      // Prepare metadata
      const fullMetadata = {
        name: file.name,
        mime_type: file.type || "application/octet-stream",
        size: file.size,
        created_at: new Date().toISOString(),
        uploaded_at: new Date().toISOString(),
        last_modified: file.lastModified
          ? new Date(file.lastModified).toISOString()
          : undefined,
      };

      // Encrypt metadata
      const encryptedMetadata = await this.cryptoService.encryptWithKey(
        JSON.stringify(fullMetadata),
        fileKey,
      );

      // Generate file ID
      const fileId = this.cryptoService.generateUUID();

      // Prepare API data
      const fileData = {
        id: fileId,
        collection_id: collectionId,
        encrypted_metadata: encryptedMetadata,
        encrypted_file_key: {
          ciphertext: this.cryptoService.uint8ArrayToBase64(
            encryptedFileKeyData.ciphertext,
          ),
          nonce: this.cryptoService.uint8ArrayToBase64(
            encryptedFileKeyData.nonce,
          ),
          key_version: encryptedFileKeyData.key_version || 1,
        },
        encryption_version: "v1.0",
        encrypted_hash: encryptedHash,
        expected_file_size_in_bytes: encryptedBytes.length,
        name: file.name, // For logging/events
      };

      // Perform complete upload workflow
      return await this.uploadFileComplete(fileData, encryptedBlob);
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to create and upload file from File object:",
        error,
      );
      throw error;
    }
  }

  // Create pending file from File object (original method)
  async createPendingFileFromFile(file, collectionId, password = null) {
    try {
      console.log("[CreateFileManager] Reading file:", file.name);

      // Read file content
      const fileContent = await this.readFileAsArrayBuffer(file);

      // Prepare metadata
      const metadata = {
        name: file.name,
        mime_type: file.type || "application/octet-stream",
        size: file.size,
        last_modified: file.lastModified
          ? new Date(file.lastModified).toISOString()
          : undefined,
      };

      // Create pending file
      return await this.createPendingFile(
        fileContent,
        collectionId,
        metadata,
        password,
      );
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to create pending file from File object:",
        error,
      );
      throw error;
    }
  }

  // === Helper Methods ===

  // Read file as ArrayBuffer
  async readFileAsArrayBuffer(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (e) => resolve(e.target.result);
      reader.onerror = (e) => reject(new Error("Failed to read file"));
      reader.readAsArrayBuffer(file);
    });
  }

  // === Password Management ===

  // Get user password from password storage service
  async getUserPassword() {
    try {
      const { default: passwordStorageService } = await import(
        "../../PasswordStorageService.js"
      );
      return passwordStorageService.getPassword();
    } catch (error) {
      console.error("[CreateFileManager] Failed to get user password:", error);
      return null;
    }
  }

  // === File Management ===

  // Get all pending files
  getPendingFiles() {
    return this.storageService.getPendingFiles();
  }

  // Get pending file by ID
  getPendingFileById(fileId) {
    return this.storageService.getPendingFileById(fileId);
  }

  // Get upload queue
  getUploadQueue() {
    return this.storageService.getUploadQueue();
  }

  // Remove pending file
  async removePendingFile(fileId) {
    try {
      console.log("[CreateFileManager] Removing pending file:", fileId);

      const removed = this.storageService.removePendingFile(fileId);

      if (removed) {
        this.notifyFileCreationListeners("pending_file_removed", {
          fileId,
        });
      }

      return removed;
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to remove pending file:",
        error,
      );
      throw error;
    }
  }

  // Clear all pending files
  async clearAllPendingFiles() {
    try {
      console.log("[CreateFileManager] Clearing all pending files");

      this.storageService.clearAllPendingFiles();

      this.notifyFileCreationListeners("all_pending_files_cleared", {});

      console.log("[CreateFileManager] All pending files cleared");
    } catch (error) {
      console.error(
        "[CreateFileManager] Failed to clear pending files:",
        error,
      );
      throw error;
    }
  }

  // === Event Management ===

  // Add file creation listener
  addFileCreationListener(callback) {
    if (typeof callback === "function") {
      this.fileCreationListeners.add(callback);
      console.log(
        "[CreateFileManager] File creation listener added. Total listeners:",
        this.fileCreationListeners.size,
      );
    }
  }

  // Remove file creation listener
  removeFileCreationListener(callback) {
    this.fileCreationListeners.delete(callback);
    console.log(
      "[CreateFileManager] File creation listener removed. Total listeners:",
      this.fileCreationListeners.size,
    );
  }

  // Notify file creation listeners
  notifyFileCreationListeners(eventType, eventData) {
    console.log(
      `[CreateFileManager] Notifying ${this.fileCreationListeners.size} listeners of ${eventType}`,
    );

    this.fileCreationListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[CreateFileManager] Error in file creation listener:",
          error,
        );
      }
    });
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canCreateFiles: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.fileCreationListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "CreateFileManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default CreateFileManager;

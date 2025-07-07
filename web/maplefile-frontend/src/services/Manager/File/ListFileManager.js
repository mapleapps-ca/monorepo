// File: monorepo/web/maplefile-frontend/src/services/Manager/File/ListFileManager.js
// List File Manager - Orchestrates API, Storage, and Crypto services for file listing

import ListFileAPIService from "../../API/File/ListFileAPIService.js";
import ListFileStorageService from "../../Storage/File/ListFileStorageService.js";

class ListFileManager {
  constructor(authManager) {
    // ListFileManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new ListFileAPIService(authManager);
    this.storageService = new ListFileStorageService();

    // Event listeners for file listing events
    this.fileListingListeners = new Set();

    console.log(
      "[ListFileManager] File manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[ListFileManager] Initializing file manager...");

      // Initialize crypto services
      const { default: FileCryptoService } = await import(
        "../../Crypto/FileCryptoService.js"
      );
      await FileCryptoService.initialize();
      this.fileCryptoService = FileCryptoService;

      // Initialize collection crypto service for collection keys
      const { default: CollectionCryptoService } = await import(
        "../../Crypto/CollectionCryptoService.js"
      );
      await CollectionCryptoService.initialize();
      this.collectionCryptoService = CollectionCryptoService;

      console.log("[ListFileManager] File manager initialized successfully");
    } catch (error) {
      console.error(
        "[ListFileManager] Failed to initialize file manager:",
        error,
      );
    }
  }

  // === File States Constants ===

  get FILE_STATES() {
    return {
      PENDING: "pending",
      ACTIVE: "active",
      DELETED: "deleted",
      ARCHIVED: "archived",
    };
  }

  // === Core File Listing Methods ===

  // List files by collection with optional state filtering
  async listFilesByCollection(
    collectionId,
    includeStates = null,
    forceRefresh = false,
  ) {
    try {
      this.isLoading = true;
      console.log(
        "[ListFileManager] Listing files for collection:",
        collectionId,
      );
      console.log("[ListFileManager] Include states:", includeStates);
      console.log("[ListFileManager] Force refresh:", forceRefresh);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedFileList = this.storageService.getFileList(collectionId);
        if (cachedFileList) {
          console.log(
            "[ListFileManager] Using cached file list:",
            cachedFileList.files.length,
            "files",
          );

          // Filter by states if requested
          let files = cachedFileList.files;
          if (includeStates && Array.isArray(includeStates)) {
            files = files.filter((file) => includeStates.includes(file.state));
          }

          this.notifyFileListingListeners("files_loaded_from_cache", {
            collectionId,
            count: files.length,
            fromCache: true,
          });

          return files;
        }
      }

      // Fetch from API
      console.log("[ListFileManager] Fetching files from API");
      const response = await this.apiService.listFilesByCollection(
        collectionId,
        includeStates,
      );

      let files = response.files || [];

      // Normalize files
      files = files.map((file) => this.fileCryptoService.normalizeFile(file));

      console.log(`[ListFileManager] Fetched ${files.length} files from API`);

      // Get collection key for decryption
      let collectionKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey) {
        console.log(
          "[ListFileManager] No cached collection key, trying to load collection...",
        );

        // Try to load collection to get its key
        try {
          await this.ensureCollectionLoaded(collectionId);
          collectionKey =
            this.collectionCryptoService.getCachedCollectionKey(collectionId);
        } catch (collectionError) {
          console.warn(
            "[ListFileManager] Could not load collection:",
            collectionError.message,
          );
        }
      }

      // Decrypt files if we have collection key
      if (collectionKey) {
        console.log("[ListFileManager] Decrypting files with collection key");
        files = await this.fileCryptoService.decryptFilesFromAPI(
          files,
          collectionKey,
        );
        console.log("[ListFileManager] File decryption completed");
      } else {
        console.warn(
          "[ListFileManager] No collection key available - files will show as encrypted",
        );
      }

      // Store in cache if no decryption errors
      const hasDecryptErrors = files.some((f) => f._decryptionError);
      if (!hasDecryptErrors && files.length > 0) {
        this.storageService.storeFileList(collectionId, files, {
          includeStates,
          fetched_at: new Date().toISOString(),
        });
      }

      // Store individual files in cache
      files.forEach((file) => {
        this.storageService.storeFile(file);
      });

      this.notifyFileListingListeners("files_loaded_from_api", {
        collectionId,
        count: files.length,
        fromCache: false,
        hasDecryptErrors,
      });

      console.log(
        "[ListFileManager] Files retrieved and processed:",
        files.length,
      );
      return files;
    } catch (error) {
      console.error("[ListFileManager] Failed to list files:", error);

      this.notifyFileListingListeners("files_load_failed", {
        collectionId,
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get individual file by ID
  async getFileById(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[ListFileManager] Getting file by ID:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedFile = this.storageService.getFile(fileId);
        if (cachedFile) {
          console.log("[ListFileManager] Using cached file:", fileId);
          return cachedFile;
        }
      }

      // Fetch from API
      const file = await this.apiService.getFileById(fileId);
      const normalizedFile = this.fileCryptoService.normalizeFile(file);

      // Try to decrypt if we have collection key
      const collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        normalizedFile.collection_id,
      );

      let finalFile = normalizedFile;
      if (collectionKey) {
        finalFile = await this.fileCryptoService.decryptFileFromAPI(
          normalizedFile,
          collectionKey,
        );
      }

      // Store in cache
      this.storageService.storeFile(finalFile);

      console.log("[ListFileManager] File retrieved:", fileId);
      return finalFile;
    } catch (error) {
      console.error("[ListFileManager] Failed to get file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === File Download Methods ===

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[ListFileManager] Getting download URL for:", fileId);

      const response = await this.apiService.getPresignedDownloadUrl(
        fileId,
        urlDuration,
      );

      console.log("[ListFileManager] Download URL obtained for:", fileId);
      return response;
    } catch (error) {
      console.error("[ListFileManager] Failed to get download URL:", error);
      throw error;
    }
  }

  // Download file content from S3
  async downloadFileFromS3(presignedUrl) {
    try {
      console.log("[ListFileManager] Downloading file from S3");

      const response = await fetch(presignedUrl, {
        method: "GET",
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 download failed with status: ${response.status}`);
      }

      const blob = await response.blob();
      console.log("[ListFileManager] File downloaded from S3 successfully");

      return blob;
    } catch (error) {
      console.error(
        "[ListFileManager] Failed to download file from S3:",
        error,
      );
      throw error;
    }
  }

  // Complete download and decrypt workflow
  async downloadAndDecryptFile(fileId) {
    try {
      console.log(
        "[ListFileManager] Starting download and decrypt for:",
        fileId,
      );

      // Step 1: Get file metadata (should be cached and decrypted)
      const fileMetadata = this.storageService.getFile(fileId);
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

      // Check if file can be downloaded
      if (fileMetadata._is_deleted && !this.canRestoreFile(fileMetadata)) {
        throw new Error("File is deleted and cannot be downloaded.");
      }

      console.log(
        "[ListFileManager] File metadata found:",
        fileMetadata.name,
        "version:",
        fileMetadata.version,
        "state:",
        fileMetadata.state,
      );

      // Step 2: Get presigned download URL and download encrypted content
      const downloadResponse = await this.getPresignedDownloadUrl(fileId);
      const encryptedContent = await this.downloadFileFromS3(
        downloadResponse.presigned_download_url,
      );

      console.log(
        "[ListFileManager] Encrypted content downloaded, size:",
        encryptedContent.size,
      );

      // Step 3: Decrypt the file content
      console.log("[ListFileManager] Decrypting file content...");
      const decryptedBytes = await this.fileCryptoService.decryptFileContent(
        encryptedContent,
        fileMetadata._file_key,
      );

      console.log(
        "[ListFileManager] File decrypted successfully, size:",
        decryptedBytes.length,
      );

      // Step 4: Create blob with proper MIME type
      const mimeType = fileMetadata.mime_type || "application/octet-stream";
      const decryptedBlob = new Blob([decryptedBytes], { type: mimeType });

      // Step 5: Get the original filename
      const filename =
        fileMetadata.name || `downloaded_file_${fileId.substring(0, 8)}`;

      console.log(
        "[ListFileManager] Download prepared:",
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
        "[ListFileManager] Failed to download and decrypt file:",
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

      console.log(
        "[ListFileManager] Browser download triggered for:",
        filename,
      );
    } catch (error) {
      console.error(
        "[ListFileManager] Failed to trigger browser download:",
        error,
      );
      throw error;
    }
  }

  // Combined download and save function
  async downloadAndSaveFile(fileId) {
    try {
      console.log(
        "[ListFileManager] Starting download and save for file:",
        fileId,
      );

      const downloadResult = await this.downloadAndDecryptFile(fileId);

      // Trigger browser download
      this.downloadBlobAsFile(downloadResult.blob, downloadResult.filename);

      console.log("[ListFileManager] File download completed successfully");
      return downloadResult;
    } catch (error) {
      console.error("[ListFileManager] Download and save failed:", error);
      throw error;
    }
  }

  // === File State Queries ===

  // Get files by state for a collection
  getFilesByState(collectionId, state = this.FILE_STATES.ACTIVE) {
    return this.storageService.getFilesByState(collectionId, [state]);
  }

  // Get files by multiple states
  getFilesByStates(collectionId, states = [this.FILE_STATES.ACTIVE]) {
    return this.storageService.getFilesByState(collectionId, states);
  }

  // Get active files for a collection
  getActiveFiles(collectionId) {
    return this.getFilesByState(collectionId, this.FILE_STATES.ACTIVE);
  }

  // Get archived files for a collection
  getArchivedFiles(collectionId) {
    return this.getFilesByState(collectionId, this.FILE_STATES.ARCHIVED);
  }

  // Get deleted files for a collection
  getDeletedFiles(collectionId) {
    return this.getFilesByState(collectionId, this.FILE_STATES.DELETED);
  }

  // Get pending files for a collection
  getPendingFiles(collectionId) {
    return this.getFilesByState(collectionId, this.FILE_STATES.PENDING);
  }

  // === File Capability Checks ===

  // Check if a file can be downloaded
  canDownloadFile(file) {
    return !file._is_deleted || this.canRestoreFile(file);
  }

  // Check if a file can be edited
  canEditFile(file) {
    return file._is_active || file._is_archived;
  }

  // Check if a file can be restored
  canRestoreFile(file) {
    return file._has_tombstone && !file._tombstone_expired && file._is_deleted;
  }

  // Check if a file can be permanently deleted
  canPermanentlyDeleteFile(file) {
    return file._tombstone_expired || (file._has_tombstone && file._is_deleted);
  }

  // === File Statistics ===

  // Get file statistics for a collection
  getFileStats(collectionId) {
    return this.storageService.getFileStats(collectionId);
  }

  // Get file version information
  getFileVersionInfo(file) {
    return {
      currentVersion: file.version || 1,
      hasTombstone: file._has_tombstone,
      tombstoneVersion: file.tombstone_version || 0,
      tombstoneExpiry: file.tombstone_expiry,
      isExpired: file._tombstone_expired,
      canRestore: this.canRestoreFile(file),
      canPermanentlyDelete: this.canPermanentlyDeleteFile(file),
    };
  }

  // === Sync Operations ===

  // Sync files for offline support
  async syncFiles(cursor = null, limit = 5000) {
    try {
      this.isLoading = true;
      console.log("[ListFileManager] Syncing files", { cursor, limit });

      const response = await this.apiService.syncFiles(cursor, limit);

      // Normalize and process sync response
      if (response.files) {
        response.files = response.files.map((file) =>
          this.fileCryptoService.normalizeFile(file),
        );
      }

      console.log("[ListFileManager] Files synced:", {
        count: response.files?.length || 0,
        hasMore: response.has_more || false,
      });

      return response;
    } catch (error) {
      console.error("[ListFileManager] Failed to sync files:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Collection Loading Helper ===

  // Ensure collection is loaded and key is cached
  async ensureCollectionLoaded(collectionId) {
    try {
      // Try to get collection via collection managers
      const password = await this.getUserPassword();

      // Import collection service
      const { default: CollectionService } = await import(
        "../../API/Collection/GetCollectionAPIService.js"
      );

      const collection = await CollectionService.getCollection(
        collectionId,
        password,
      );

      if (collection && collection.collection_key) {
        this.collectionCryptoService.cacheCollectionKey(
          collectionId,
          collection.collection_key,
        );
        console.log(
          "[ListFileManager] Collection loaded and key cached:",
          collectionId,
        );
      }
    } catch (error) {
      console.warn("[ListFileManager] Failed to load collection:", error);
      throw error;
    }
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
      console.error("[ListFileManager] Failed to get user password:", error);
      return null;
    }
  }

  // === Cache Management ===

  // Clear cache for specific collection
  clearCollectionCache(collectionId) {
    this.storageService.clearCollectionCache(collectionId);
    console.log(
      "[ListFileManager] Cache cleared for collection:",
      collectionId,
    );
  }

  // Clear all file caches
  clearAllCaches() {
    this.storageService.clearAllFileCaches();
    this.fileCryptoService.clearFileKeyCache();
    console.log("[ListFileManager] All caches cleared");
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // === Event Management ===

  // Add file listing listener
  addFileListingListener(callback) {
    if (typeof callback === "function") {
      this.fileListingListeners.add(callback);
      console.log(
        "[ListFileManager] File listing listener added. Total listeners:",
        this.fileListingListeners.size,
      );
    }
  }

  // Remove file listing listener
  removeFileListingListener(callback) {
    this.fileListingListeners.delete(callback);
    console.log(
      "[ListFileManager] File listing listener removed. Total listeners:",
      this.fileListingListeners.size,
    );
  }

  // Notify file listing listeners
  notifyFileListingListeners(eventType, eventData) {
    console.log(
      `[ListFileManager] Notifying ${this.fileListingListeners.size} listeners of ${eventType}`,
    );

    this.fileListingListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[ListFileManager] Error in file listing listener:",
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
      canListFiles: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.fileListingListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "ListFileManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      fileCryptoService: this.fileCryptoService?.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default ListFileManager;

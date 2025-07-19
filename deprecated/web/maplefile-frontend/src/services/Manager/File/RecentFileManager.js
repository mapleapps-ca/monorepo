// File: monorepo/web/maplefile-frontend/src/services/Manager/File/RecentFileManager.js
// Recent File Manager - Orchestrates API, Storage, and Crypto services for recent file listing

import RecentFileAPIService from "../../API/File/RecentFileAPIService.js";
import RecentFileStorageService from "../../Storage/File/RecentFileStorageService.js";

class RecentFileManager {
  constructor(
    authManager,
    getCollectionManager = null,
    listCollectionManager = null,
  ) {
    // RecentFileManager depends on AuthManager and collection managers
    this.authManager = authManager;
    this.getCollectionManager = getCollectionManager;
    this.listCollectionManager = listCollectionManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new RecentFileAPIService(authManager);
    this.storageService = new RecentFileStorageService();

    // Event listeners for recent file events
    this.recentFileListeners = new Set();

    // Pagination state
    this.paginationState = {
      currentCursor: null,
      hasMore: false,
      totalLoaded: 0,
    };

    console.log(
      "[RecentFileManager] Manager initialized with AuthManager and collection managers",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[RecentFileManager] Initializing manager...");

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

      console.log("[RecentFileManager] Manager initialized successfully");
    } catch (error) {
      console.error("[RecentFileManager] Failed to initialize manager:", error);
    }
  }

  // === Core Recent Files Methods ===

  // Get recent files with caching and decryption
  async getRecentFiles(limit = 30, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[RecentFileManager] === Getting Recent Files ===");
      console.log("[RecentFileManager] Limit:", limit);
      console.log("[RecentFileManager] Force refresh:", forceRefresh);

      // STEP 1: Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedRecentFiles = this.storageService.getRecentFiles();
        if (cachedRecentFiles) {
          console.log(
            "[RecentFileManager] Found cached recent files:",
            cachedRecentFiles.files.length,
            "files",
          );

          // ✅ ALWAYS re-decrypt cached files since _file_key is not stored in cache
          console.log(
            "[RecentFileManager] Re-decrypting cached files (file keys not stored in cache for security)...",
          );

          // Re-decrypt the cached files
          const decryptedFiles = await this.decryptRecentFiles(
            cachedRecentFiles.files,
          );

          const decryptedCount = decryptedFiles.filter(
            (f) => f._isDecrypted,
          ).length;
          const errorCount = decryptedFiles.filter(
            (f) => f._decryptionError,
          ).length;

          console.log(
            "[RecentFileManager] Cache re-decryption results:",
            decryptedCount,
            "successful,",
            errorCount,
            "errors",
          );

          this.notifyRecentFileListeners("recent_files_loaded_from_cache", {
            count: decryptedFiles.length,
            fromCache: true,
            decryptedCount,
            errorCount,
            reDecrypted: true,
          });

          console.log(
            "[RecentFileManager] Returning cached files (re-decrypted):",
            decryptedFiles.length,
          );
          return decryptedFiles;
        }
      }

      // STEP 2: Fetch from API
      console.log("[RecentFileManager] Fetching recent files from API");
      const response = await this.apiService.listRecentFiles(limit);

      let files = response.files || [];

      // STEP 3: Normalize files
      files = files.map((file) => this.fileCryptoService.normalizeFile(file));

      console.log(`[RecentFileManager] Fetched ${files.length} files from API`);

      // STEP 4: Decrypt files
      console.log("[RecentFileManager] Decrypting files...");
      const decryptedFiles = await this.decryptRecentFiles(files);
      console.log("[RecentFileManager] File decryption completed");

      // Log decryption results
      const decryptedCount = decryptedFiles.filter(
        (f) => f._isDecrypted,
      ).length;
      const errorCount = decryptedFiles.filter(
        (f) => f._decryptionError,
      ).length;
      console.log(
        `[RecentFileManager] Decryption results: ${decryptedCount} successful, ${errorCount} errors`,
      );

      // STEP 5: Store in cache if no major decryption errors
      const hasDecryptErrors = decryptedFiles.some((f) => f._decryptionError);
      if (!hasDecryptErrors && decryptedFiles.length > 0) {
        this.storageService.storeRecentFiles(decryptedFiles, {
          fetched_at: new Date().toISOString(),
          decrypted_at: new Date().toISOString(),
          limit,
          hasMore: response.has_more || false,
          nextCursor: response.next_cursor || null,
        });
      }

      // STEP 6: Store individual files in cache
      decryptedFiles.forEach((file) => {
        this.storageService.storeFile(file);
      });

      // STEP 7: Update pagination state
      this.paginationState = {
        currentCursor: response.next_cursor || null,
        hasMore: response.has_more || false,
        totalLoaded: decryptedFiles.length,
      };

      this.notifyRecentFileListeners("recent_files_loaded_from_api", {
        count: decryptedFiles.length,
        fromCache: false,
        hasDecryptErrors,
        decryptedCount,
        errorCount,
        hasMore: response.has_more || false,
        nextCursor: response.next_cursor || null,
      });

      console.log(
        "[RecentFileManager] Recent files retrieved and processed successfully:",
        decryptedFiles.length,
      );
      return decryptedFiles;
    } catch (error) {
      console.error("[RecentFileManager] Failed to get recent files:", error);

      this.notifyRecentFileListeners("recent_files_load_failed", {
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Load more recent files (pagination)
  async loadMoreRecentFiles(limit = 30) {
    try {
      this.isLoading = true;
      console.log("[RecentFileManager] === Loading More Recent Files ===");
      console.log(
        "[RecentFileManager] Current cursor:",
        this.paginationState.currentCursor,
      );
      console.log(
        "[RecentFileManager] Has more:",
        this.paginationState.hasMore,
      );

      if (!this.paginationState.hasMore) {
        console.log("[RecentFileManager] No more files to load");
        return [];
      }

      if (!this.paginationState.currentCursor) {
        throw new Error("No cursor available for pagination");
      }

      // Fetch next page from API
      const response = await this.apiService.listRecentFiles(
        limit,
        this.paginationState.currentCursor,
      );

      let files = response.files || [];

      // Normalize files
      files = files.map((file) => this.fileCryptoService.normalizeFile(file));

      console.log(
        `[RecentFileManager] Fetched ${files.length} more files from API`,
      );

      // Decrypt files
      const decryptedFiles = await this.decryptRecentFiles(files);

      // Update pagination state
      this.paginationState = {
        currentCursor: response.next_cursor || null,
        hasMore: response.has_more || false,
        totalLoaded: this.paginationState.totalLoaded + decryptedFiles.length,
      };

      // Store individual files in cache
      decryptedFiles.forEach((file) => {
        this.storageService.storeFile(file);
      });

      // Note: We don't update the main recent files cache for pagination
      // This is because the cache should represent the first page for quick access

      this.notifyRecentFileListeners("more_recent_files_loaded", {
        count: decryptedFiles.length,
        totalLoaded: this.paginationState.totalLoaded,
        hasMore: this.paginationState.hasMore,
      });

      console.log(
        "[RecentFileManager] More recent files loaded successfully:",
        decryptedFiles.length,
      );
      return decryptedFiles;
    } catch (error) {
      console.error(
        "[RecentFileManager] Failed to load more recent files:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === File Decryption ===

  // Decrypt recent files by ensuring collection keys are available
  async decryptRecentFiles(files) {
    if (!files || files.length === 0) {
      return [];
    }

    console.log("[RecentFileManager] === Decrypting Recent Files ===");
    console.log("[RecentFileManager] Files to decrypt:", files.length);

    // Group files by collection to optimize collection key loading
    const filesByCollection = this.groupFilesByCollection(files);
    const collectionIds = Object.keys(filesByCollection);

    console.log(
      "[RecentFileManager] Files grouped by collection:",
      collectionIds.length,
      "collections",
    );

    // Load collection keys for all collections
    await this.ensureCollectionKeysLoaded(collectionIds);

    // Decrypt files by collection
    const decryptedFiles = [];

    for (const collectionId of collectionIds) {
      const collectionFiles = filesByCollection[collectionId];

      console.log(
        `[RecentFileManager] Decrypting ${collectionFiles.length} files for collection: ${collectionId}`,
      );

      // Get collection key
      const collectionKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey) {
        console.warn(
          `[RecentFileManager] No collection key available for collection: ${collectionId}`,
        );

        // Add files with error marker
        collectionFiles.forEach((file) => {
          decryptedFiles.push({
            ...file,
            name: "[Collection key unavailable]",
            _isDecrypted: false,
            _decryptionError: "Collection key not available",
          });
        });
        continue;
      }

      // Decrypt files with collection key
      const decryptedCollectionFiles =
        await this.fileCryptoService.decryptFilesFromAPI(
          collectionFiles,
          collectionKey,
        );

      decryptedFiles.push(...decryptedCollectionFiles);
    }

    const successCount = decryptedFiles.filter((f) => f._isDecrypted).length;
    const errorCount = decryptedFiles.filter((f) => f._decryptionError).length;

    console.log(`[RecentFileManager] === Decryption Summary ===`);
    console.log(`[RecentFileManager] Total files: ${decryptedFiles.length}`);
    console.log(`[RecentFileManager] Successfully decrypted: ${successCount}`);
    console.log(`[RecentFileManager] Decryption errors: ${errorCount}`);

    return decryptedFiles;
  }

  // Group files by collection ID
  groupFilesByCollection(files) {
    const grouped = {};

    files.forEach((file) => {
      const collectionId = file.collection_id;
      if (!grouped[collectionId]) {
        grouped[collectionId] = [];
      }
      grouped[collectionId].push(file);
    });

    return grouped;
  }

  // Ensure collection keys are loaded for all collection IDs
  async ensureCollectionKeysLoaded(collectionIds) {
    console.log("[RecentFileManager] === Loading Collection Keys ===");
    console.log(
      "[RecentFileManager] Collections needed:",
      collectionIds.length,
    );

    if (!this.getCollectionManager) {
      throw new Error(
        "GetCollectionManager not available. Please pass it to the constructor.",
      );
    }

    const loadPromises = collectionIds.map(async (collectionId) => {
      try {
        // Check if we already have the collection key cached
        let cachedKey =
          this.collectionCryptoService.getCachedCollectionKey(collectionId);
        if (cachedKey) {
          console.log(
            "[RecentFileManager] Collection key already cached:",
            collectionId,
          );
          return;
        }

        // Load collection using collection manager
        console.log(
          "[RecentFileManager] Loading collection to get key:",
          collectionId,
        );

        const collection =
          await this.getCollectionManager.getCollection(collectionId);

        console.log("[RecentFileManager] Collection loaded:", {
          id: collection.id,
          name: collection.name,
          hasCollectionKey: !!collection.collection_key,
        });

        // Verify collection key is available
        if (!collection.collection_key) {
          throw new Error(
            "Collection key not available after loading collection",
          );
        }

        // Cache the collection key
        this.collectionCryptoService.cacheCollectionKey(
          collectionId,
          collection.collection_key,
        );

        console.log(
          "[RecentFileManager] Collection key cached successfully:",
          collectionId,
        );
      } catch (error) {
        console.error(
          `[RecentFileManager] Failed to load collection ${collectionId}:`,
          error,
        );
        // Continue with other collections even if one fails
      }
    });

    // Wait for all collection keys to be loaded
    await Promise.allSettled(loadPromises);

    console.log("[RecentFileManager] Collection key loading completed");
  }

  // === File Operations ===

  // Get individual file by ID
  async getFileById(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[RecentFileManager] Getting file by ID:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedFile = this.storageService.getFile(fileId);
        if (cachedFile) {
          console.log("[RecentFileManager] Using cached file:", fileId);
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
      } else {
        // Try to load collection key
        await this.ensureCollectionKeysLoaded([normalizedFile.collection_id]);
        const newCollectionKey =
          this.collectionCryptoService.getCachedCollectionKey(
            normalizedFile.collection_id,
          );

        if (newCollectionKey) {
          finalFile = await this.fileCryptoService.decryptFileFromAPI(
            normalizedFile,
            newCollectionKey,
          );
        }
      }

      // Store in cache
      this.storageService.storeFile(finalFile);

      console.log("[RecentFileManager] File retrieved:", fileId);
      return finalFile;
    } catch (error) {
      console.error("[RecentFileManager] Failed to get file:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Download Operations ===

  // Get presigned download URL for file
  async getPresignedDownloadUrl(fileId, urlDuration = null) {
    try {
      console.log("[RecentFileManager] Getting download URL for:", fileId);

      const response = await this.apiService.getPresignedDownloadUrl(
        fileId,
        urlDuration,
      );

      console.log("[RecentFileManager] Download URL obtained for:", fileId);
      return response;
    } catch (error) {
      console.error("[RecentFileManager] Failed to get download URL:", error);
      throw error;
    }
  }

  // Download file content from S3
  async downloadFileFromS3(presignedUrl) {
    try {
      console.log("[RecentFileManager] Downloading file from S3");

      const response = await fetch(presignedUrl, {
        method: "GET",
        mode: "cors",
      });

      if (!response.ok) {
        throw new Error(`S3 download failed with status: ${response.status}`);
      }

      const blob = await response.blob();
      console.log("[RecentFileManager] File downloaded from S3 successfully");

      return blob;
    } catch (error) {
      console.error(
        "[RecentFileManager] Failed to download file from S3:",
        error,
      );
      throw error;
    }
  }

  // Complete download and decrypt workflow
  async downloadAndDecryptFile(fileId) {
    try {
      console.log(
        "[RecentFileManager] Starting download and decrypt for:",
        fileId,
      );

      // Step 1: Get file metadata (should be cached and decrypted)
      let fileMetadata = this.storageService.getFile(fileId);
      if (!fileMetadata) {
        // Try to get from API
        fileMetadata = await this.getFileById(fileId);
      }

      if (!fileMetadata) {
        throw new Error(
          "File metadata not found. Please refresh the recent files list.",
        );
      }

      console.log(
        "[RecentFileManager] File metadata found:",
        fileMetadata.name,
        "version:",
        fileMetadata.version,
        "state:",
        fileMetadata.state,
      );

      // Check if file key is available, if not, re-decrypt it
      if (!fileMetadata._file_key) {
        console.log(
          "[RecentFileManager] File key not available in cache, re-decrypting file key...",
        );

        // Load collection key if needed
        await this.ensureCollectionKeysLoaded([fileMetadata.collection_id]);

        const collectionKey =
          this.collectionCryptoService.getCachedCollectionKey(
            fileMetadata.collection_id,
          );

        if (!collectionKey) {
          throw new Error(
            "Collection key not available for file decryption. Please try refreshing the page.",
          );
        }

        if (!fileMetadata.encrypted_file_key) {
          throw new Error(
            "Encrypted file key not available in file metadata. File may be corrupted.",
          );
        }

        // Decrypt the file key
        const fileKey = await this.fileCryptoService.decryptFileKey(
          fileMetadata.encrypted_file_key,
          collectionKey,
        );

        // Cache the file key in memory and update the file metadata
        this.fileCryptoService.cacheFileKey(fileId, fileKey);
        fileMetadata._file_key = fileKey;
        fileMetadata._hasFileKey = true;

        console.log(
          "[RecentFileManager] File key re-decrypted successfully, length:",
          fileKey.length,
        );
      }

      // Step 2: Get presigned download URL and download encrypted content
      const downloadResponse = await this.getPresignedDownloadUrl(fileId);
      const encryptedContent = await this.downloadFileFromS3(
        downloadResponse.presigned_download_url,
      );

      console.log(
        "[RecentFileManager] Encrypted content downloaded, size:",
        encryptedContent.size,
      );

      // Step 3: Decrypt the file content
      console.log("[RecentFileManager] Decrypting file content...");
      const decryptedBytes = await this.fileCryptoService.decryptFileContent(
        encryptedContent,
        fileMetadata._file_key,
      );

      console.log(
        "[RecentFileManager] File decrypted successfully, size:",
        decryptedBytes.length,
      );

      // Step 4: Create blob with proper MIME type
      const mimeType = fileMetadata.mime_type || "application/octet-stream";
      const decryptedBlob = new Blob([decryptedBytes], { type: mimeType });

      // Step 5: Get the original filename
      const filename =
        fileMetadata.name || `downloaded_file_${fileId.substring(0, 8)}`;

      console.log(
        "[RecentFileManager] Download prepared:",
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
        "[RecentFileManager] Failed to download and decrypt file:",
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
        "[RecentFileManager] Browser download triggered for:",
        filename,
      );
    } catch (error) {
      console.error(
        "[RecentFileManager] Failed to trigger browser download:",
        error,
      );
      throw error;
    }
  }

  // Combined download and save function
  async downloadAndSaveFile(fileId) {
    try {
      console.log(
        "[RecentFileManager] Starting download and save for file:",
        fileId,
      );

      const downloadResult = await this.downloadAndDecryptFile(fileId);

      // Trigger browser download
      this.downloadBlobAsFile(downloadResult.blob, downloadResult.filename);

      console.log("[RecentFileManager] File download completed successfully");
      return downloadResult;
    } catch (error) {
      console.error("[RecentFileManager] Download and save failed:", error);
      throw error;
    }
  }

  // === Cache Management ===

  // Clear all recent files caches
  clearAllCaches() {
    this.storageService.clearAllCaches();
    this.fileCryptoService.clearFileKeyCache();
    console.log("[RecentFileManager] All caches cleared");
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // === Pagination State ===

  // Get current pagination state
  getPaginationState() {
    return { ...this.paginationState };
  }

  // Reset pagination state
  resetPaginationState() {
    this.paginationState = {
      currentCursor: null,
      hasMore: false,
      totalLoaded: 0,
    };
    console.log("[RecentFileManager] Pagination state reset");
  }

  // === Event Management ===

  // Add recent file listener
  addRecentFileListener(callback) {
    if (typeof callback === "function") {
      this.recentFileListeners.add(callback);
      console.log(
        "[RecentFileManager] Recent file listener added. Total listeners:",
        this.recentFileListeners.size,
      );
    }
  }

  // Remove recent file listener
  removeRecentFileListener(callback) {
    this.recentFileListeners.delete(callback);
    console.log(
      "[RecentFileManager] Recent file listener removed. Total listeners:",
      this.recentFileListeners.size,
    );
  }

  // Notify recent file listeners
  notifyRecentFileListeners(eventType, eventData) {
    console.log(
      `[RecentFileManager] Notifying ${this.recentFileListeners.size} listeners of ${eventType}`,
    );

    this.recentFileListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[RecentFileManager] Error in recent file listener:",
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
      listenerCount: this.recentFileListeners.size,
      paginationState: this.getPaginationState(),
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "RecentFileManager",
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

export default RecentFileManager;

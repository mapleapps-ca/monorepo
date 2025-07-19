// File: monorepo/web/maplefile-frontend/src/services/Manager/File/GetFileManager.js
// Get File Manager - Orchestrates API, Storage, and Crypto services for file retrieval

import GetFileAPIService from "../../API/File/GetFileAPIService.js";
import GetFileStorageService from "../../Storage/File/GetFileStorageService.js";

class GetFileManager {
  constructor(
    authManager,
    getCollectionManager = null,
    listCollectionManager = null,
  ) {
    // GetFileManager depends on AuthManager and collection managers
    this.authManager = authManager;
    this.getCollectionManager = getCollectionManager;
    this.listCollectionManager = listCollectionManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new GetFileAPIService(authManager);
    this.storageService = new GetFileStorageService();

    // Event listeners for file retrieval events
    this.fileRetrievalListeners = new Set();

    console.log(
      "[GetFileManager] File manager initialized with AuthManager and collection managers",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[GetFileManager] Initializing file manager...");

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

      console.log("[GetFileManager] File manager initialized successfully");
    } catch (error) {
      console.error(
        "[GetFileManager] Failed to initialize file manager:",
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

  // === Core File Retrieval Methods ===

  // Get file by ID with full details and decryption
  async getFileById(
    fileId,
    forceRefresh = false,
    includeVersionHistory = false,
  ) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] === Starting File Retrieval Workflow ===");
      console.log("[GetFileManager] File ID:", fileId);
      console.log("[GetFileManager] Force refresh:", forceRefresh);
      console.log(
        "[GetFileManager] Include version history:",
        includeVersionHistory,
      );

      // STEP 1: Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedFileDetails = this.storageService.getFileDetails(fileId);
        if (cachedFileDetails) {
          console.log("[GetFileManager] Found cached file details:", fileId);

          // Check if cached file is properly decrypted
          if (
            cachedFileDetails.fileDetails._isDecrypted &&
            cachedFileDetails.fileDetails.name &&
            cachedFileDetails.fileDetails.name !== "[Encrypted]" &&
            cachedFileDetails.fileDetails.name !== "[Unable to decrypt]"
          ) {
            console.log(
              "[GetFileManager] Cached file is already decrypted:",
              cachedFileDetails.fileDetails.name,
            );

            this.notifyFileRetrievalListeners("file_loaded_from_cache", {
              fileId,
              fromCache: true,
              isDecrypted: true,
            });

            return cachedFileDetails.fileDetails;
          } else {
            console.log(
              "[GetFileManager] Cached file needs decryption, decrypting now...",
            );

            // Try to decrypt the cached file
            try {
              const decryptedFile = await this.decryptFileWithCollectionKey(
                cachedFileDetails.fileDetails,
              );

              if (decryptedFile._isDecrypted) {
                // Update cache with decrypted file
                this.storageService.storeFileDetails(fileId, decryptedFile, {
                  decrypted_at: new Date().toISOString(),
                });

                this.notifyFileRetrievalListeners("file_loaded_from_cache", {
                  fileId,
                  fromCache: true,
                  isDecrypted: true,
                  wasReDecrypted: true,
                });

                console.log(
                  "[GetFileManager] Cached file re-decrypted successfully:",
                  decryptedFile.name,
                );
                return decryptedFile;
              }
            } catch (decryptError) {
              console.warn(
                "[GetFileManager] Failed to decrypt cached file:",
                decryptError.message,
              );
              // Continue to fetch from API
            }
          }
        }
      }

      // STEP 2: Fetch from API
      console.log("[GetFileManager] Fetching file from API");
      const response = await this.apiService.getFileById(
        fileId,
        includeVersionHistory,
      );

      if (!response || !response.id) {
        throw new Error("File not found");
      }

      let file = response;

      // STEP 3: Normalize file
      file = this.fileCryptoService.normalizeFile(file);

      console.log(
        `[GetFileManager] Fetched file from API: ${file.id} (v${file.version}, ${file.state})`,
      );

      // STEP 4: Load collection to ensure collection key is available
      await this.ensureCollectionLoaded(file.collection_id);

      // STEP 5: Get collection key for decryption
      let collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        file.collection_id,
      );

      if (!collectionKey) {
        console.error(
          "[GetFileManager] CRITICAL: No collection key available after loading collection!",
        );
        throw new Error(
          "Collection key not available for file decryption. Please try refreshing the page.",
        );
      }

      console.log(
        "[GetFileManager] Collection key available for decryption, length:",
        collectionKey.length,
      );

      // STEP 6: Decrypt file with collection key
      console.log("[GetFileManager] Decrypting file with collection key");
      file = await this.fileCryptoService.decryptFileFromAPI(
        file,
        collectionKey,
      );
      console.log("[GetFileManager] File decryption completed");

      // STEP 7: Handle version history if included
      let versionHistory = null;
      if (includeVersionHistory && response.versions) {
        console.log("[GetFileManager] Processing version history");
        versionHistory = await this.fileCryptoService.decryptFilesFromAPI(
          response.versions,
          collectionKey,
        );

        // Store version history in cache
        this.storageService.storeFileVersionHistory(fileId, versionHistory, {
          fetched_at: new Date().toISOString(),
        });
      }

      // STEP 8: Store in cache if no decryption errors
      if (!file._decryptionError) {
        this.storageService.storeFileDetails(fileId, file, {
          fetched_at: new Date().toISOString(),
          includeVersionHistory,
        });
      }

      this.notifyFileRetrievalListeners("file_loaded_from_api", {
        fileId,
        fromCache: false,
        hasDecryptError: !!file._decryptionError,
        isDecrypted: file._isDecrypted,
        hasVersionHistory: !!versionHistory,
        versionCount: versionHistory?.length || 0,
      });

      console.log(
        "[GetFileManager] File retrieved and processed successfully:",
        file.name || "[Unable to decrypt]",
      );

      // Add version history to file object if requested
      if (versionHistory) {
        file._versionHistory = versionHistory;
      }

      return file;
    } catch (error) {
      console.error("[GetFileManager] Failed to get file:", error);

      this.notifyFileRetrievalListeners("file_load_failed", {
        fileId,
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get file version history
  async getFileVersionHistory(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] Getting version history for:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedVersions =
          this.storageService.getFileVersionHistory(fileId);
        if (cachedVersions) {
          console.log("[GetFileManager] Using cached version history:", fileId);
          return cachedVersions.versionHistory;
        }
      }

      // Fetch from API
      const response = await this.apiService.getFileVersionHistory(fileId);
      let versions = response.versions || [];

      if (versions.length === 0) {
        return [];
      }

      // Normalize versions
      versions = versions.map((version) =>
        this.fileCryptoService.normalizeFile(version),
      );

      // Get collection key for decryption
      const firstVersion = versions[0];
      await this.ensureCollectionLoaded(firstVersion.collection_id);

      const collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        firstVersion.collection_id,
      );

      if (collectionKey) {
        // Decrypt all versions
        versions = await this.fileCryptoService.decryptFilesFromAPI(
          versions,
          collectionKey,
        );
      }

      // Store in cache
      this.storageService.storeFileVersionHistory(fileId, versions, {
        fetched_at: new Date().toISOString(),
      });

      console.log(
        "[GetFileManager] Version history retrieved:",
        versions.length,
      );
      return versions;
    } catch (error) {
      console.error("[GetFileManager] Failed to get version history:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get file metadata only (lightweight)
  async getFileMetadata(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] Getting file metadata:", fileId);

      // Check if we have full file details first
      if (!forceRefresh) {
        const cachedFile = this.storageService.getFileDetails(fileId);
        if (cachedFile && cachedFile.fileDetails) {
          console.log(
            "[GetFileManager] Using cached file for metadata:",
            fileId,
          );
          return this.extractMetadata(cachedFile.fileDetails);
        }
      }

      // Fetch metadata from API
      const response = await this.apiService.getFileMetadata(fileId);
      let metadata = response.metadata || response;

      // Normalize and try to decrypt if we have collection key
      if (metadata.collection_id) {
        try {
          await this.ensureCollectionLoaded(metadata.collection_id);
          const collectionKey =
            this.collectionCryptoService.getCachedCollectionKey(
              metadata.collection_id,
            );

          if (collectionKey) {
            metadata = await this.fileCryptoService.decryptFileFromAPI(
              metadata,
              collectionKey,
            );
          }
        } catch (decryptError) {
          console.warn(
            "[GetFileManager] Could not decrypt metadata:",
            decryptError,
          );
        }
      }

      console.log("[GetFileManager] File metadata retrieved:", fileId);
      return this.extractMetadata(metadata);
    } catch (error) {
      console.error("[GetFileManager] Failed to get file metadata:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get file permissions
  async getFilePermissions(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] Getting file permissions:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedPermissions =
          this.storageService.getFilePermissions(fileId);
        if (cachedPermissions) {
          console.log("[GetFileManager] Using cached permissions:", fileId);
          return cachedPermissions.permissions;
        }
      }

      // Fetch from API
      const response = await this.apiService.getFilePermissions(fileId);

      // Store in cache
      this.storageService.storeFilePermissions(fileId, response, {
        fetched_at: new Date().toISOString(),
      });

      console.log("[GetFileManager] File permissions retrieved:", fileId);
      return response;
    } catch (error) {
      console.error("[GetFileManager] Failed to get file permissions:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get file statistics
  async getFileStats(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] Getting file statistics:", fileId);

      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedStats = this.storageService.getFileStats(fileId);
        if (cachedStats) {
          console.log("[GetFileManager] Using cached statistics:", fileId);
          return cachedStats.stats;
        }
      }

      // Fetch from API
      const response = await this.apiService.getFileStats(fileId);

      // Store in cache
      this.storageService.storeFileStats(fileId, response, {
        fetched_at: new Date().toISOString(),
      });

      console.log("[GetFileManager] File statistics retrieved:", fileId);
      return response;
    } catch (error) {
      console.error("[GetFileManager] Failed to get file statistics:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Check if file exists and is accessible
  async checkFileExists(fileId) {
    try {
      console.log("[GetFileManager] Checking file existence:", fileId);

      const response = await this.apiService.checkFileExists(fileId);

      console.log("[GetFileManager] File existence check:", {
        fileId,
        exists: response.exists,
        accessible: response.accessible,
      });

      return response;
    } catch (error) {
      console.error("[GetFileManager] Failed to check file existence:", error);
      throw error;
    }
  }

  // Get complete file data (all related information)
  async getFileComplete(fileId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetFileManager] Getting complete file data:", fileId);

      // Get all data simultaneously with error handling
      const [file, versionHistory, permissions, stats] =
        await Promise.allSettled([
          this.getFileById(fileId, forceRefresh, true),
          this.getFileVersionHistory(fileId, forceRefresh).catch((err) => {
            console.warn(
              "[GetFileManager] Version history failed:",
              err.message,
            );
            return [];
          }),
          this.getFilePermissions(fileId, forceRefresh).catch((err) => {
            console.warn("[GetFileManager] Permissions failed:", err.message);
            return null;
          }),
          this.getFileStats(fileId, forceRefresh).catch((err) => {
            console.warn("[GetFileManager] Stats failed:", err.message);
            return null;
          }),
        ]);

      const result = {
        file: file.status === "fulfilled" ? file.value : null,
        versionHistory:
          versionHistory.status === "fulfilled" ? versionHistory.value : [],
        permissions:
          permissions.status === "fulfilled" ? permissions.value : null,
        stats: stats.status === "fulfilled" ? stats.value : null,
        errors: [],
      };

      // Collect any errors (but don't fail the whole operation)
      if (file.status === "rejected")
        result.errors.push({ type: "file", error: file.reason.message });
      if (versionHistory.status === "rejected")
        result.errors.push({
          type: "versionHistory",
          error: versionHistory.reason.message,
        });
      if (permissions.status === "rejected")
        result.errors.push({
          type: "permissions",
          error: permissions.reason.message,
        });
      if (stats.status === "rejected")
        result.errors.push({ type: "stats", error: stats.reason.message });

      console.log("[GetFileManager] Complete file data retrieved:", {
        fileId,
        hasFile: !!result.file,
        versionCount: result.versionHistory.length,
        hasPermissions: !!result.permissions,
        hasStats: !!result.stats,
        errorCount: result.errors.length,
      });

      return result;
    } catch (error) {
      console.error(
        "[GetFileManager] Failed to get complete file data:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Helper Methods ===

  // Decrypt file with collection key
  async decryptFileWithCollectionKey(file) {
    try {
      await this.ensureCollectionLoaded(file.collection_id);

      const collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        file.collection_id,
      );

      if (!collectionKey) {
        throw new Error("Collection key not available");
      }

      return await this.fileCryptoService.decryptFileFromAPI(
        file,
        collectionKey,
      );
    } catch (error) {
      console.error("[GetFileManager] Failed to decrypt file:", error);
      throw error;
    }
  }

  // Extract metadata from file object
  extractMetadata(file) {
    return {
      id: file.id,
      collection_id: file.collection_id,
      name: file.name || "[Encrypted]",
      mime_type: file.mime_type,
      size: file.size,
      encrypted_file_size_in_bytes: file.encrypted_file_size_in_bytes,
      state: file.state,
      version: file.version,
      created_at: file.created_at,
      modified_at: file.modified_at,
      tombstone_version: file.tombstone_version,
      tombstone_expiry: file.tombstone_expiry,
      _is_active: file._is_active,
      _is_archived: file._is_archived,
      _is_deleted: file._is_deleted,
      _is_pending: file._is_pending,
      _has_tombstone: file._has_tombstone,
      _tombstone_expired: file._tombstone_expired,
      _isDecrypted: file._isDecrypted,
      _decryptionError: file._decryptionError,
    };
  }

  // === Collection Loading Helper ===

  // Ensure collection is loaded and key is cached
  async ensureCollectionLoaded(collectionId) {
    try {
      console.log(
        "[GetFileManager] === Loading Collection for File Decryption ===",
      );

      // Check if we already have the collection key cached
      let cachedKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);
      if (cachedKey) {
        console.log(
          "[GetFileManager] Collection key already cached:",
          collectionId,
        );
        return;
      }

      // Load collection using collection manager
      if (!this.getCollectionManager) {
        throw new Error(
          "GetCollectionManager not available. Please pass it to the constructor.",
        );
      }

      console.log(
        "[GetFileManager] Loading collection to get key:",
        collectionId,
      );
      const collection =
        await this.getCollectionManager.getCollection(collectionId);

      console.log("[GetFileManager] Collection loaded:", {
        id: collection.id,
        name: collection.name,
        hasCollectionKey: !!collection.collection_key,
        collectionKeyLength: collection.collection_key?.length,
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

      // Verify the key was cached properly
      cachedKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);
      console.log(
        "[GetFileManager] Collection key cached successfully:",
        !!cachedKey,
      );

      console.log(
        "[GetFileManager] Collection loaded and key cached successfully:",
        collectionId,
      );
    } catch (error) {
      console.error("[GetFileManager] Failed to load collection:", error);
      throw new Error(
        `Failed to load collection ${collectionId}: ${error.message}`,
      );
    }
  }

  // === File State Checks ===

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

  // === Password Management ===

  // Get user password from password storage service
  async getUserPassword() {
    try {
      const { default: passwordStorageService } = await import(
        "../../PasswordStorageService.js"
      );
      return passwordStorageService.getPassword();
    } catch (error) {
      console.error("[GetFileManager] Failed to get user password:", error);
      return null;
    }
  }

  // === Cache Management ===

  // Clear cache for specific file
  clearFileCache(fileId) {
    this.storageService.clearFileCache(fileId);
    console.log("[GetFileManager] Cache cleared for file:", fileId);
  }

  // Clear all file caches
  clearAllCaches() {
    this.storageService.clearAllFileDetailCaches();
    this.fileCryptoService.clearFileKeyCache();
    console.log("[GetFileManager] All caches cleared");
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // === Event Management ===

  // Add file retrieval listener
  addFileRetrievalListener(callback) {
    if (typeof callback === "function") {
      this.fileRetrievalListeners.add(callback);
      console.log(
        "[GetFileManager] File retrieval listener added. Total listeners:",
        this.fileRetrievalListeners.size,
      );
    }
  }

  // Remove file retrieval listener
  removeFileRetrievalListener(callback) {
    this.fileRetrievalListeners.delete(callback);
    console.log(
      "[GetFileManager] File retrieval listener removed. Total listeners:",
      this.fileRetrievalListeners.size,
    );
  }

  // Notify file retrieval listeners
  notifyFileRetrievalListeners(eventType, eventData) {
    console.log(
      `[GetFileManager] Notifying ${this.fileRetrievalListeners.size} listeners of ${eventType}`,
    );

    this.fileRetrievalListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[GetFileManager] Error in file retrieval listener:",
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
      canGetFiles: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.fileRetrievalListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "GetFileManager",
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

export default GetFileManager;

// File: monorepo/web/maplefile-frontend/src/services/Manager/File/DownloadFileManager.js
// Download File Manager - Orchestrates API, Storage, and Crypto services for file downloads

import DownloadFileAPIService from "../../API/File/DownloadFileAPIService.js";
import DownloadFileStorageService from "../../Storage/File/DownloadFileStorageService.js";

class DownloadFileManager {
  constructor(authManager, getFileManager = null, getCollectionManager = null) {
    // DownloadFileManager depends on AuthManager and other managers
    this.authManager = authManager;
    this.getFileManager = getFileManager;
    this.getCollectionManager = getCollectionManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new DownloadFileAPIService(authManager);
    this.storageService = new DownloadFileStorageService();

    // Event listeners for download events
    this.downloadListeners = new Set();

    // Download progress tracking
    this.activeDownloads = new Map();

    console.log(
      "[DownloadFileManager] Download manager initialized with AuthManager and dependent managers",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[DownloadFileManager] Initializing download manager...");

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

      console.log(
        "[DownloadFileManager] Download manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[DownloadFileManager] Failed to initialize download manager:",
        error,
      );
    }
  }

  // === Download States Constants ===

  get DOWNLOAD_STATES() {
    return {
      PREPARING: "preparing",
      DOWNLOADING: "downloading",
      DECRYPTING: "decrypting",
      COMPLETED: "completed",
      FAILED: "failed",
      CANCELLED: "cancelled",
    };
  }

  // === Core Download Methods ===

  // Download file with full workflow (get URLs, download, decrypt, save)
  async downloadFile(fileId, options = {}) {
    const {
      forceRefresh = false,
      onProgress = null,
      urlDuration = null,
      saveToDisk = true,
      returnBlob = false,
    } = options;

    let downloadId = null;

    try {
      this.isLoading = true;
      downloadId = `download_${fileId}_${Date.now()}`;

      console.log("[DownloadFileManager] === Starting Download Workflow ===");
      console.log("[DownloadFileManager] File ID:", fileId);
      console.log("[DownloadFileManager] Download ID:", downloadId);
      console.log("[DownloadFileManager] Force refresh:", forceRefresh);
      console.log("[DownloadFileManager] Save to disk:", saveToDisk);

      // Track download progress
      this.activeDownloads.set(downloadId, {
        fileId,
        state: this.DOWNLOAD_STATES.PREPARING,
        progress: 0,
        startTime: Date.now(),
      });

      this.notifyDownloadListeners("download_started", {
        downloadId,
        fileId,
        state: this.DOWNLOAD_STATES.PREPARING,
      });

      // STEP 1: Get file metadata and ensure it can be downloaded
      console.log("[DownloadFileManager] Step 1: Getting file metadata");
      const fileMetadata = await this.getFileForDownload(fileId, forceRefresh);

      if (!this.canDownloadFile(fileMetadata)) {
        throw new Error("File cannot be downloaded in its current state");
      }

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.PREPARING,
        10,
      );

      // STEP 2: Get or generate download URLs
      console.log("[DownloadFileManager] Step 2: Getting download URLs");
      const downloadUrls = await this.getDownloadUrls(
        fileId,
        forceRefresh,
        urlDuration,
      );

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.PREPARING,
        20,
      );

      // STEP 3: Ensure collection is loaded and key is available
      console.log(
        "[DownloadFileManager] Step 3: Loading collection for decryption",
      );
      await this.ensureCollectionLoaded(fileMetadata.collection_id);

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.PREPARING,
        30,
      );

      // STEP 4: Download encrypted content
      console.log(
        "[DownloadFileManager] Step 4: Downloading encrypted content",
      );
      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.DOWNLOADING,
        35,
      );

      const downloadProgressCallback = (received, total, percentage) => {
        const adjustedProgress = 35 + percentage * 0.4; // 35% to 75%
        this.updateDownloadProgress(
          downloadId,
          this.DOWNLOAD_STATES.DOWNLOADING,
          adjustedProgress,
        );

        if (onProgress) {
          onProgress({
            downloadId,
            state: this.DOWNLOAD_STATES.DOWNLOADING,
            progress: adjustedProgress,
            received,
            total,
          });
        }
      };

      const encryptedContent = await this.apiService.downloadFileFromS3(
        downloadUrls.downloadUrl,
        downloadProgressCallback,
      );

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.DOWNLOADING,
        75,
      );

      // STEP 5: Download thumbnail if available
      let encryptedThumbnail = null;
      if (
        downloadUrls.thumbnailUrl &&
        fileMetadata.encrypted_thumbnail_size_in_bytes > 0
      ) {
        console.log("[DownloadFileManager] Step 5a: Downloading thumbnail");
        try {
          encryptedThumbnail = await this.apiService.downloadThumbnailFromS3(
            downloadUrls.thumbnailUrl,
          );
        } catch (thumbError) {
          console.warn(
            "[DownloadFileManager] Thumbnail download failed:",
            thumbError,
          );
          // Thumbnail failure is not critical
        }
      }

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.DECRYPTING,
        80,
      );

      // STEP 6: Decrypt file content
      console.log("[DownloadFileManager] Step 6: Decrypting file content");
      const decryptedResult = await this.decryptDownloadedFile(
        fileMetadata,
        encryptedContent,
        encryptedThumbnail,
      );

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.DECRYPTING,
        90,
      );

      // STEP 7: Save to disk or return blob
      let result = {
        downloadId,
        fileId,
        fileName: decryptedResult.fileName,
        blob: decryptedResult.blob,
        mimeType: decryptedResult.mimeType,
        size: decryptedResult.size,
        thumbnail: decryptedResult.thumbnail,
      };

      if (saveToDisk) {
        console.log("[DownloadFileManager] Step 7: Saving to disk");
        this.saveBlobToFile(decryptedResult.blob, decryptedResult.fileName);
      }

      // STEP 8: Update storage and complete
      this.storageService.addToDownloadHistory(
        fileId,
        decryptedResult.fileName,
        {
          downloadId,
          downloadedAt: new Date().toISOString(),
          fileSize: decryptedResult.size,
          mimeType: decryptedResult.mimeType,
        },
      );

      // Report completion to API (for analytics)
      this.apiService.reportDownloadCompletion(fileId, {
        downloadId,
        downloadDuration:
          Date.now() - this.activeDownloads.get(downloadId).startTime,
        fileSize: decryptedResult.size,
      });

      this.updateDownloadProgress(
        downloadId,
        this.DOWNLOAD_STATES.COMPLETED,
        100,
      );

      this.notifyDownloadListeners("download_completed", {
        downloadId,
        fileId,
        fileName: decryptedResult.fileName,
        success: true,
      });

      console.log(
        "[DownloadFileManager] Download workflow completed successfully",
      );

      if (returnBlob) {
        return result;
      } else {
        return {
          downloadId,
          fileName: decryptedResult.fileName,
          success: true,
        };
      }
    } catch (error) {
      console.error("[DownloadFileManager] Download workflow failed:", error);

      if (downloadId) {
        this.updateDownloadProgress(downloadId, this.DOWNLOAD_STATES.FAILED, 0);
        this.notifyDownloadListeners("download_failed", {
          downloadId,
          fileId,
          error: error.message,
        });
      }

      throw error;
    } finally {
      this.isLoading = false;
      if (downloadId) {
        // Clean up after delay
        setTimeout(() => {
          this.activeDownloads.delete(downloadId);
        }, 5000);
      }
    }
  }

  // Download multiple files (batch download)
  async downloadMultipleFiles(fileIds, options = {}) {
    const {
      onProgress = null,
      onFileComplete = null,
      concurrent = 3,
    } = options;

    try {
      console.log(
        "[DownloadFileManager] Starting batch download for",
        fileIds.length,
        "files",
      );

      const results = [];
      const errors = [];

      // Process files in batches
      for (let i = 0; i < fileIds.length; i += concurrent) {
        const batch = fileIds.slice(i, i + concurrent);

        const batchPromises = batch.map(async (fileId) => {
          try {
            const result = await this.downloadFile(fileId, {
              ...options,
              onProgress: (progressData) => {
                if (onProgress) {
                  onProgress({
                    ...progressData,
                    batchProgress: {
                      completed: results.length,
                      total: fileIds.length,
                      current: fileId,
                    },
                  });
                }
              },
            });

            if (onFileComplete) {
              onFileComplete(result);
            }

            return result;
          } catch (error) {
            console.error(
              `[DownloadFileManager] Failed to download file ${fileId}:`,
              error,
            );
            return { fileId, error: error.message, success: false };
          }
        });

        const batchResults = await Promise.all(batchPromises);

        batchResults.forEach((result) => {
          if (result.success !== false) {
            results.push(result);
          } else {
            errors.push(result);
          }
        });
      }

      console.log(
        `[DownloadFileManager] Batch download completed: ${results.length} successful, ${errors.length} failed`,
      );

      return {
        successful: results,
        failed: errors,
        total: fileIds.length,
      };
    } catch (error) {
      console.error("[DownloadFileManager] Batch download failed:", error);
      throw error;
    }
  }

  // Download only thumbnail
  async downloadThumbnail(fileId, options = {}) {
    const { forceRefresh = false } = options;

    try {
      console.log("[DownloadFileManager] Downloading thumbnail for:", fileId);

      // Get file metadata
      const fileMetadata = await this.getFileForDownload(fileId, forceRefresh);

      if (!fileMetadata.encrypted_thumbnail_size_in_bytes) {
        throw new Error("File has no thumbnail");
      }

      // Get thumbnail URL
      const thumbnailData = await this.getThumbnailUrl(fileId, forceRefresh);

      // Download encrypted thumbnail
      const encryptedThumbnail = await this.apiService.downloadThumbnailFromS3(
        thumbnailData.thumbnailUrl,
      );

      // Decrypt thumbnail
      const decryptedThumbnail = await this.decryptThumbnail(
        fileMetadata,
        encryptedThumbnail,
      );

      console.log(
        "[DownloadFileManager] Thumbnail downloaded and decrypted successfully",
      );

      return {
        fileId,
        thumbnail: decryptedThumbnail,
        mimeType: "image/jpeg", // Thumbnails are typically JPEG
      };
    } catch (error) {
      console.error("[DownloadFileManager] Thumbnail download failed:", error);
      throw error;
    }
  }

  // === Helper Methods ===

  // Get file metadata for download
  async getFileForDownload(fileId, forceRefresh = false) {
    try {
      // Use GetFileManager if available, otherwise use API directly
      if (this.getFileManager) {
        return await this.getFileManager.getFileById(fileId, forceRefresh);
      } else {
        const fileData = await this.apiService.getFileForDownload(fileId);
        return this.fileCryptoService.normalizeFile(fileData);
      }
    } catch (error) {
      console.error(
        "[DownloadFileManager] Failed to get file for download:",
        error,
      );
      throw error;
    }
  }

  // Get download URLs (with caching)
  async getDownloadUrls(fileId, forceRefresh = false, urlDuration = null) {
    try {
      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedUrls = this.storageService.getDownloadUrl(fileId);
        if (cachedUrls) {
          console.log(
            "[DownloadFileManager] Using cached download URLs for:",
            fileId,
          );
          return {
            downloadUrl: cachedUrls.downloadUrl,
            thumbnailUrl: cachedUrls.thumbnailUrl,
            fileMetadata: cachedUrls.fileMetadata,
          };
        }
      }

      // Get fresh URLs from API
      console.log(
        "[DownloadFileManager] Fetching fresh download URLs for:",
        fileId,
      );
      const response = await this.apiService.getPresignedDownloadUrl(
        fileId,
        urlDuration,
      );

      // Cache the URLs
      this.storageService.storeDownloadUrl(fileId, response, {
        requested_at: new Date().toISOString(),
        url_duration: urlDuration,
      });

      return {
        downloadUrl: response.presigned_download_url,
        thumbnailUrl: response.presigned_thumbnail_url,
        fileMetadata: response.file,
      };
    } catch (error) {
      console.error(
        "[DownloadFileManager] Failed to get download URLs:",
        error,
      );
      throw error;
    }
  }

  // Get thumbnail URL specifically
  async getThumbnailUrl(fileId, forceRefresh = false, urlDuration = null) {
    try {
      // Check cache first unless forcing refresh
      if (!forceRefresh) {
        const cachedThumbnail = this.storageService.getThumbnailUrl(fileId);
        if (cachedThumbnail) {
          console.log(
            "[DownloadFileManager] Using cached thumbnail URL for:",
            fileId,
          );
          return {
            thumbnailUrl: cachedThumbnail.thumbnailUrl,
          };
        }
      }

      // Get fresh thumbnail URL from API
      console.log(
        "[DownloadFileManager] Fetching fresh thumbnail URL for:",
        fileId,
      );
      const response = await this.apiService.getPresignedThumbnailUrl(
        fileId,
        urlDuration,
      );

      // Cache the thumbnail URL
      this.storageService.storeThumbnailUrl(fileId, response, {
        requested_at: new Date().toISOString(),
        url_duration: urlDuration,
      });

      return {
        thumbnailUrl: response.presigned_thumbnail_url,
      };
    } catch (error) {
      console.error(
        "[DownloadFileManager] Failed to get thumbnail URL:",
        error,
      );
      throw error;
    }
  }

  // Decrypt downloaded file content
  async decryptDownloadedFile(
    fileMetadata,
    encryptedContent,
    encryptedThumbnail = null,
  ) {
    try {
      console.log("[DownloadFileManager] Decrypting downloaded file");

      // Get collection key
      const collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        fileMetadata.collection_id,
      );

      if (!collectionKey) {
        throw new Error("Collection key not available for file decryption");
      }

      // Convert blob to array buffer
      const encryptedArrayBuffer = await encryptedContent.arrayBuffer();
      const encryptedBytes = new Uint8Array(encryptedArrayBuffer);

      // Get file key from metadata or decrypt it
      let fileKey = fileMetadata._file_key;
      if (!fileKey) {
        console.log("[DownloadFileManager] Decrypting file key");
        fileKey = await this.fileCryptoService.decryptFileKey(
          fileMetadata.encrypted_file_key,
          collectionKey,
        );
      }

      // Decrypt file content
      console.log("[DownloadFileManager] Decrypting file content");
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );

      const decryptedBytes = await CryptoService.decryptWithKey(
        CryptoService.uint8ArrayToBase64(encryptedBytes),
        fileKey,
      );

      // Get file metadata (name, mime type, etc.)
      let fileName =
        fileMetadata.name || `file_${fileMetadata.id.substring(0, 8)}`;
      let mimeType = fileMetadata.mime_type || "application/octet-stream";

      // If metadata is encrypted, decrypt it
      if (!fileMetadata._isDecrypted && fileMetadata.encrypted_metadata) {
        try {
          const decryptedMetadata =
            await this.fileCryptoService.decryptFileMetadata(
              fileMetadata.encrypted_metadata,
              fileKey,
            );
          fileName = decryptedMetadata.name || fileName;
          mimeType = decryptedMetadata.mime_type || mimeType;
        } catch (metadataError) {
          console.warn(
            "[DownloadFileManager] Failed to decrypt metadata:",
            metadataError,
          );
          // Continue with existing values
        }
      }

      // Create decrypted blob
      const decryptedBlob = new Blob([decryptedBytes], { type: mimeType });

      // Decrypt thumbnail if provided
      let decryptedThumbnail = null;
      if (encryptedThumbnail) {
        try {
          decryptedThumbnail = await this.decryptThumbnail(
            fileMetadata,
            encryptedThumbnail,
          );
        } catch (thumbnailError) {
          console.warn(
            "[DownloadFileManager] Failed to decrypt thumbnail:",
            thumbnailError,
          );
        }
      }

      console.log("[DownloadFileManager] File decryption completed");

      return {
        blob: decryptedBlob,
        fileName,
        mimeType,
        size: decryptedBlob.size,
        thumbnail: decryptedThumbnail,
      };
    } catch (error) {
      console.error("[DownloadFileManager] File decryption failed:", error);
      throw new Error(`Failed to decrypt file: ${error.message}`);
    }
  }

  // Decrypt thumbnail
  async decryptThumbnail(fileMetadata, encryptedThumbnail) {
    try {
      console.log("[DownloadFileManager] Decrypting thumbnail");

      // Get collection key
      const collectionKey = this.collectionCryptoService.getCachedCollectionKey(
        fileMetadata.collection_id,
      );

      if (!collectionKey) {
        throw new Error(
          "Collection key not available for thumbnail decryption",
        );
      }

      // Get file key
      let fileKey = fileMetadata._file_key;
      if (!fileKey) {
        fileKey = await this.fileCryptoService.decryptFileKey(
          fileMetadata.encrypted_file_key,
          collectionKey,
        );
      }

      // Convert blob to array buffer
      const encryptedArrayBuffer = await encryptedThumbnail.arrayBuffer();
      const encryptedBytes = new Uint8Array(encryptedArrayBuffer);

      // Decrypt thumbnail content
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );

      const decryptedBytes = await CryptoService.decryptWithKey(
        CryptoService.uint8ArrayToBase64(encryptedBytes),
        fileKey,
      );

      // Create thumbnail blob (usually JPEG)
      const thumbnailBlob = new Blob([decryptedBytes], { type: "image/jpeg" });

      console.log("[DownloadFileManager] Thumbnail decryption completed");
      return thumbnailBlob;
    } catch (error) {
      console.error(
        "[DownloadFileManager] Thumbnail decryption failed:",
        error,
      );
      throw error;
    }
  }

  // Save blob as file download
  saveBlobToFile(blob, fileName) {
    try {
      // Create object URL
      const url = URL.createObjectURL(blob);

      // Create temporary download link
      const downloadLink = document.createElement("a");
      downloadLink.href = url;
      downloadLink.download = fileName;
      downloadLink.style.display = "none";

      // Add to document, click, and remove
      document.body.appendChild(downloadLink);
      downloadLink.click();
      document.body.removeChild(downloadLink);

      // Clean up object URL
      setTimeout(() => {
        URL.revokeObjectURL(url);
      }, 1000);

      console.log("[DownloadFileManager] File saved to disk:", fileName);
    } catch (error) {
      console.error(
        "[DownloadFileManager] Failed to save file to disk:",
        error,
      );
      throw error;
    }
  }

  // Ensure collection is loaded and key is cached
  async ensureCollectionLoaded(collectionId) {
    try {
      // Check if we already have the collection key cached
      let cachedKey =
        this.collectionCryptoService.getCachedCollectionKey(collectionId);
      if (cachedKey) {
        console.log(
          "[DownloadFileManager] Collection key already cached:",
          collectionId,
        );
        return;
      }

      // Load collection using collection manager
      if (!this.getCollectionManager) {
        throw new Error(
          "GetCollectionManager not available for loading collection",
        );
      }

      console.log(
        "[DownloadFileManager] Loading collection to get key:",
        collectionId,
      );
      const collection =
        await this.getCollectionManager.getCollection(collectionId);

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
        "[DownloadFileManager] Collection loaded and key cached:",
        collectionId,
      );
    } catch (error) {
      console.error("[DownloadFileManager] Failed to load collection:", error);
      throw new Error(
        `Failed to load collection ${collectionId}: ${error.message}`,
      );
    }
  }

  // Check if file can be downloaded
  canDownloadFile(file) {
    // Can download if file is active, archived, or deleted (but restorable)
    if (file._is_active || file._is_archived) {
      return true;
    }

    if (file._is_deleted && file._has_tombstone && !file._tombstone_expired) {
      return true; // Can download deleted files that haven't expired
    }

    return false;
  }

  // Update download progress
  updateDownloadProgress(downloadId, state, progress) {
    if (this.activeDownloads.has(downloadId)) {
      const download = this.activeDownloads.get(downloadId);
      download.state = state;
      download.progress = progress;
      download.lastUpdate = Date.now();
      this.activeDownloads.set(downloadId, download);
    }
  }

  // Cancel download
  cancelDownload(downloadId) {
    if (this.activeDownloads.has(downloadId)) {
      const download = this.activeDownloads.get(downloadId);
      download.state = this.DOWNLOAD_STATES.CANCELLED;
      this.activeDownloads.set(downloadId, download);

      this.notifyDownloadListeners("download_cancelled", {
        downloadId,
        fileId: download.fileId,
      });

      console.log("[DownloadFileManager] Download cancelled:", downloadId);
    }
  }

  // === Cache Management ===

  // Clear download cache for specific file
  clearFileDownloadCache(fileId) {
    this.storageService.clearFileDownloadCache(fileId);
    console.log(
      "[DownloadFileManager] Download cache cleared for file:",
      fileId,
    );
  }

  // Clear all download caches
  clearAllDownloadCaches() {
    this.storageService.clearAllDownloadCaches();
    console.log("[DownloadFileManager] All download caches cleared");
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // === Event Management ===

  // Add download listener
  addDownloadListener(callback) {
    if (typeof callback === "function") {
      this.downloadListeners.add(callback);
      console.log(
        "[DownloadFileManager] Download listener added. Total listeners:",
        this.downloadListeners.size,
      );
    }
  }

  // Remove download listener
  removeDownloadListener(callback) {
    this.downloadListeners.delete(callback);
    console.log(
      "[DownloadFileManager] Download listener removed. Total listeners:",
      this.downloadListeners.size,
    );
  }

  // Notify download listeners
  notifyDownloadListeners(eventType, eventData) {
    console.log(
      `[DownloadFileManager] Notifying ${this.downloadListeners.size} listeners of ${eventType}`,
    );

    this.downloadListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[DownloadFileManager] Error in download listener:",
          error,
        );
      }
    });
  }

  // === Status and Information ===

  // Get active downloads
  getActiveDownloads() {
    return Array.from(this.activeDownloads.values());
  }

  // Get download history
  getDownloadHistory(limit = 20) {
    return this.storageService.getRecentDownloads(limit);
  }

  // Check if file was downloaded recently
  wasFileDownloadedRecently(fileId, withinMinutes = 60) {
    return this.storageService.wasFileDownloadedRecently(fileId, withinMinutes);
  }

  // Get download statistics
  getDownloadStats() {
    const activeDownloads = this.getActiveDownloads();
    const history = this.getDownloadHistory(100);

    return {
      activeDownloads: activeDownloads.length,
      totalDownloadsToday: history.filter((record) => {
        const downloadDate = new Date(record.downloadedAt);
        const today = new Date();
        return downloadDate.toDateString() === today.toDateString();
      }).length,
      totalDownloadsThisWeek: history.filter((record) => {
        const downloadDate = new Date(record.downloadedAt);
        const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
        return downloadDate > weekAgo;
      }).length,
      recentDownloads: history.slice(0, 5),
    };
  }

  // Get manager status
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canDownloadFiles: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.downloadListeners.size,
      activeDownloads: this.activeDownloads.size,
      downloadStates: this.DOWNLOAD_STATES,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "DownloadFileManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      fileCryptoService: this.fileCryptoService?.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      downloadStats: this.getDownloadStats(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default DownloadFileManager;

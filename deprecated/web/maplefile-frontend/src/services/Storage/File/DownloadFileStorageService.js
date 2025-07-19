// File: monorepo/web/maplefile-frontend/src/services/Storage/File/DownloadFileStorageService.js
// Download File Storage Service - Handles localStorage operations for download URLs and caching

class DownloadFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      DOWNLOAD_URLS: "mapleapps_download_urls",
      DOWNLOAD_HISTORY: "mapleapps_download_history",
      DOWNLOAD_METADATA: "mapleapps_download_metadata",
      THUMBNAIL_URLS: "mapleapps_thumbnail_urls",
    };

    // Cache configuration
    this.URL_CACHE_DURATION = 5 * 60 * 1000; // 5 minutes for URLs (they expire quickly)
    this.METADATA_CACHE_DURATION = 30 * 60 * 1000; // 30 minutes for metadata
    this.HISTORY_RETENTION_DAYS = 7; // Keep download history for 7 days

    console.log("[DownloadFileStorageService] Storage service initialized");
  }

  // === Download URL Caching ===

  // Store presigned download URL with expiration
  storeDownloadUrl(fileId, downloadUrlData, metadata = {}) {
    try {
      const downloadData = {
        fileId,
        downloadUrl: downloadUrlData.presigned_download_url,
        thumbnailUrl: downloadUrlData.presigned_thumbnail_url,
        fileMetadata: downloadUrlData.file,
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(
            Date.now() + this.URL_CACHE_DURATION,
          ).toISOString(),
          url_expires_at: downloadUrlData.download_url_expiration_time,
        },
      };

      const existingUrls = this.getAllDownloadUrls();
      existingUrls[fileId] = downloadData;

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_URLS,
        JSON.stringify(existingUrls),
      );

      console.log(
        "[DownloadFileStorageService] Download URL stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to store download URL:",
        error,
      );
    }
  }

  // Get cached download URL
  getDownloadUrl(fileId) {
    try {
      const allUrls = this.getAllDownloadUrls();
      const downloadData = allUrls[fileId];

      if (!downloadData) {
        return null;
      }

      // Check if cache has expired
      const cacheExpiresAt = new Date(downloadData.metadata.expires_at);
      const urlExpiresAt = new Date(downloadData.metadata.url_expires_at);
      const now = new Date();

      if (now > cacheExpiresAt || now > urlExpiresAt) {
        console.log(
          "[DownloadFileStorageService] Download URL cache expired for:",
          fileId,
        );
        this.removeDownloadUrl(fileId);
        return null;
      }

      console.log(
        "[DownloadFileStorageService] Download URL retrieved from cache:",
        fileId,
      );

      return downloadData;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get download URL:",
        error,
      );
      return null;
    }
  }

  // Get all stored download URLs
  getAllDownloadUrls() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DOWNLOAD_URLS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get all download URLs:",
        error,
      );
      return {};
    }
  }

  // Remove download URL from cache
  removeDownloadUrl(fileId) {
    try {
      const allUrls = this.getAllDownloadUrls();
      delete allUrls[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_URLS,
        JSON.stringify(allUrls),
      );

      console.log(
        "[DownloadFileStorageService] Download URL removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to remove download URL:",
        error,
      );
      return false;
    }
  }

  // === Thumbnail URL Caching ===

  // Store thumbnail URL separately
  storeThumbnailUrl(fileId, thumbnailUrlData, metadata = {}) {
    try {
      const thumbnailData = {
        fileId,
        thumbnailUrl: thumbnailUrlData.presigned_thumbnail_url,
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(
            Date.now() + this.URL_CACHE_DURATION,
          ).toISOString(),
          url_expires_at: thumbnailUrlData.thumbnail_url_expiration_time,
        },
      };

      const existingThumbnails = this.getAllThumbnailUrls();
      existingThumbnails[fileId] = thumbnailData;

      localStorage.setItem(
        this.STORAGE_KEYS.THUMBNAIL_URLS,
        JSON.stringify(existingThumbnails),
      );

      console.log(
        "[DownloadFileStorageService] Thumbnail URL stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to store thumbnail URL:",
        error,
      );
    }
  }

  // Get cached thumbnail URL
  getThumbnailUrl(fileId) {
    try {
      const allThumbnails = this.getAllThumbnailUrls();
      const thumbnailData = allThumbnails[fileId];

      if (!thumbnailData) {
        return null;
      }

      // Check if cache has expired
      const cacheExpiresAt = new Date(thumbnailData.metadata.expires_at);
      const urlExpiresAt = new Date(thumbnailData.metadata.url_expires_at);
      const now = new Date();

      if (now > cacheExpiresAt || now > urlExpiresAt) {
        console.log(
          "[DownloadFileStorageService] Thumbnail URL cache expired for:",
          fileId,
        );
        this.removeThumbnailUrl(fileId);
        return null;
      }

      console.log(
        "[DownloadFileStorageService] Thumbnail URL retrieved from cache:",
        fileId,
      );

      return thumbnailData;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get thumbnail URL:",
        error,
      );
      return null;
    }
  }

  // Get all stored thumbnail URLs
  getAllThumbnailUrls() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.THUMBNAIL_URLS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get all thumbnail URLs:",
        error,
      );
      return {};
    }
  }

  // Remove thumbnail URL from cache
  removeThumbnailUrl(fileId) {
    try {
      const allThumbnails = this.getAllThumbnailUrls();
      delete allThumbnails[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.THUMBNAIL_URLS,
        JSON.stringify(allThumbnails),
      );

      console.log(
        "[DownloadFileStorageService] Thumbnail URL removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to remove thumbnail URL:",
        error,
      );
      return false;
    }
  }

  // === Download History ===

  // Add download to history
  addToDownloadHistory(fileId, fileName, downloadMetadata = {}) {
    try {
      const downloadRecord = {
        fileId,
        fileName,
        downloadedAt: new Date().toISOString(),
        metadata: downloadMetadata,
      };

      const history = this.getDownloadHistory();

      // Remove existing entry for this file to avoid duplicates
      const filteredHistory = history.filter(
        (record) => record.fileId !== fileId,
      );

      // Add new record at the beginning
      filteredHistory.unshift(downloadRecord);

      // Keep only recent downloads (limit by count and age)
      const maxRecords = 100;
      const cutoffDate = new Date(
        Date.now() - this.HISTORY_RETENTION_DAYS * 24 * 60 * 60 * 1000,
      );

      const trimmedHistory = filteredHistory
        .slice(0, maxRecords)
        .filter((record) => new Date(record.downloadedAt) > cutoffDate);

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_HISTORY,
        JSON.stringify(trimmedHistory),
      );

      console.log(
        "[DownloadFileStorageService] Added to download history:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to add to download history:",
        error,
      );
    }
  }

  // Get download history
  getDownloadHistory() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DOWNLOAD_HISTORY);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get download history:",
        error,
      );
      return [];
    }
  }

  // Get recent downloads (last N downloads)
  getRecentDownloads(limit = 10) {
    try {
      const history = this.getDownloadHistory();
      return history.slice(0, limit);
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get recent downloads:",
        error,
      );
      return [];
    }
  }

  // Check if file was downloaded recently
  wasFileDownloadedRecently(fileId, withinMinutes = 60) {
    try {
      const history = this.getDownloadHistory();
      const cutoffTime = new Date(Date.now() - withinMinutes * 60 * 1000);

      return history.some(
        (record) =>
          record.fileId === fileId &&
          new Date(record.downloadedAt) > cutoffTime,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to check recent download:",
        error,
      );
      return false;
    }
  }

  // Clear download history
  clearDownloadHistory() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DOWNLOAD_HISTORY);
      console.log("[DownloadFileStorageService] Download history cleared");
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to clear download history:",
        error,
      );
    }
  }

  // === Download Metadata Storage ===

  // Store download metadata (file info, progress, etc.)
  storeDownloadMetadata(fileId, metadata) {
    try {
      const downloadMetadata = {
        fileId,
        metadata: {
          ...metadata,
          stored_at: new Date().toISOString(),
          expires_at: new Date(
            Date.now() + this.METADATA_CACHE_DURATION,
          ).toISOString(),
        },
      };

      const existingMetadata = this.getAllDownloadMetadata();
      existingMetadata[fileId] = downloadMetadata;

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_METADATA,
        JSON.stringify(existingMetadata),
      );

      console.log(
        "[DownloadFileStorageService] Download metadata stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to store download metadata:",
        error,
      );
    }
  }

  // Get download metadata
  getDownloadMetadata(fileId) {
    try {
      const allMetadata = this.getAllDownloadMetadata();
      const metadata = allMetadata[fileId];

      if (!metadata) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(metadata.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[DownloadFileStorageService] Download metadata cache expired for:",
          fileId,
        );
        this.removeDownloadMetadata(fileId);
        return null;
      }

      console.log(
        "[DownloadFileStorageService] Download metadata retrieved from cache:",
        fileId,
      );

      return metadata;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get download metadata:",
        error,
      );
      return null;
    }
  }

  // Get all download metadata
  getAllDownloadMetadata() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DOWNLOAD_METADATA);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to get all download metadata:",
        error,
      );
      return {};
    }
  }

  // Remove download metadata
  removeDownloadMetadata(fileId) {
    try {
      const allMetadata = this.getAllDownloadMetadata();
      delete allMetadata[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_METADATA,
        JSON.stringify(allMetadata),
      );

      console.log(
        "[DownloadFileStorageService] Download metadata removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to remove download metadata:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired download URLs
      const allUrls = this.getAllDownloadUrls();
      Object.keys(allUrls).forEach((fileId) => {
        const urlData = allUrls[fileId];
        const expiresAt = new Date(urlData.metadata.expires_at);
        if (now > expiresAt) {
          delete allUrls[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.DOWNLOAD_URLS,
          JSON.stringify(allUrls),
        );
      }

      // Clear expired thumbnail URLs
      const allThumbnails = this.getAllThumbnailUrls();
      Object.keys(allThumbnails).forEach((fileId) => {
        const thumbnailData = allThumbnails[fileId];
        const expiresAt = new Date(thumbnailData.metadata.expires_at);
        if (now > expiresAt) {
          delete allThumbnails[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.THUMBNAIL_URLS,
          JSON.stringify(allThumbnails),
        );
      }

      // Clear expired download metadata
      const allMetadata = this.getAllDownloadMetadata();
      Object.keys(allMetadata).forEach((fileId) => {
        const metadata = allMetadata[fileId];
        const expiresAt = new Date(metadata.metadata.expires_at);
        if (now > expiresAt) {
          delete allMetadata[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.DOWNLOAD_METADATA,
          JSON.stringify(allMetadata),
        );
      }

      console.log(
        "[DownloadFileStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all download caches
  clearAllDownloadCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DOWNLOAD_URLS);
      localStorage.removeItem(this.STORAGE_KEYS.THUMBNAIL_URLS);
      localStorage.removeItem(this.STORAGE_KEYS.DOWNLOAD_METADATA);
      console.log("[DownloadFileStorageService] All download caches cleared");
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to clear download caches:",
        error,
      );
    }
  }

  // Clear cache for specific file
  clearFileDownloadCache(fileId) {
    try {
      this.removeDownloadUrl(fileId);
      this.removeThumbnailUrl(fileId);
      this.removeDownloadMetadata(fileId);

      console.log(
        "[DownloadFileStorageService] Download cache cleared for file:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to clear file download cache:",
        error,
      );
    }
  }

  // === Batch Operations ===

  // Store multiple download URLs at once
  storeMultipleDownloadUrls(downloadUrlsMap) {
    try {
      const allUrls = this.getAllDownloadUrls();
      let storedCount = 0;

      Object.entries(downloadUrlsMap).forEach(([fileId, urlData]) => {
        allUrls[fileId] = {
          fileId,
          downloadUrl: urlData.presigned_download_url,
          thumbnailUrl: urlData.presigned_thumbnail_url,
          fileMetadata: urlData.file,
          metadata: {
            cached_at: new Date().toISOString(),
            expires_at: new Date(
              Date.now() + this.URL_CACHE_DURATION,
            ).toISOString(),
            url_expires_at: urlData.download_url_expiration_time,
          },
        };
        storedCount++;
      });

      localStorage.setItem(
        this.STORAGE_KEYS.DOWNLOAD_URLS,
        JSON.stringify(allUrls),
      );

      console.log(
        "[DownloadFileStorageService] Batch stored",
        storedCount,
        "download URLs",
      );
      return storedCount;
    } catch (error) {
      console.error(
        "[DownloadFileStorageService] Failed to batch store download URLs:",
        error,
      );
      return 0;
    }
  }

  // === Configuration ===

  // Set URL cache duration
  setUrlCacheDuration(duration) {
    this.URL_CACHE_DURATION = duration;
    console.log(
      "[DownloadFileStorageService] URL cache duration set to:",
      duration,
      "ms",
    );
  }

  // Set metadata cache duration
  setMetadataCacheDuration(duration) {
    this.METADATA_CACHE_DURATION = duration;
    console.log(
      "[DownloadFileStorageService] Metadata cache duration set to:",
      duration,
      "ms",
    );
  }

  // Set history retention days
  setHistoryRetentionDays(days) {
    this.HISTORY_RETENTION_DAYS = days;
    console.log(
      "[DownloadFileStorageService] History retention set to:",
      days,
      "days",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const allUrls = this.getAllDownloadUrls();
    const allThumbnails = this.getAllThumbnailUrls();
    const allMetadata = this.getAllDownloadMetadata();
    const history = this.getDownloadHistory();

    return {
      downloadUrlsCount: Object.keys(allUrls).length,
      thumbnailUrlsCount: Object.keys(allThumbnails).length,
      metadataCount: Object.keys(allMetadata).length,
      historyCount: history.length,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      urlCacheDuration: this.URL_CACHE_DURATION,
      metadataCacheDuration: this.METADATA_CACHE_DURATION,
      historyRetentionDays: this.HISTORY_RETENTION_DAYS,
    };
  }

  // Get debug information
  getDebugInfo() {
    const allUrls = this.getAllDownloadUrls();
    const history = this.getDownloadHistory();

    return {
      serviceName: "DownloadFileStorageService",
      storageInfo: this.getStorageInfo(),
      cachedDownloadFileIds: Object.keys(allUrls),
      recentDownloads: this.getRecentDownloads(5).map((record) => ({
        fileId: record.fileId,
        fileName: record.fileName,
        downloadedAt: record.downloadedAt,
      })),
    };
  }
}

export default DownloadFileStorageService;

// File: monorepo/web/maplefile-frontend/src/services/Storage/File/GetFileStorageService.js
// Get File Storage Service - Handles localStorage operations for individual file details and caching

class GetFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      FILE_DETAILS: "mapleapps_file_details",
      FILE_VERSIONS: "mapleapps_file_versions",
      FILE_PERMISSIONS: "mapleapps_file_permissions",
      FILE_STATS: "mapleapps_file_stats",
    };

    // Cache configuration
    this.CACHE_DURATION = 10 * 60 * 1000; // 10 minutes default for file details
    this.EXTENDED_CACHE_DURATION = 30 * 60 * 1000; // 30 minutes for version history

    console.log("[GetFileStorageService] Storage service initialized");
  }

  // === File Details Storage Operations ===

  // Store complete file details
  storeFileDetails(fileId, fileDetails, metadata = {}) {
    try {
      const fileDetailsData = {
        fileId,
        fileDetails: this.sanitizeFileDetailsForStorage(fileDetails),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingDetails = this.getAllFileDetails();
      existingDetails[fileId] = fileDetailsData;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_DETAILS,
        JSON.stringify(existingDetails),
      );

      console.log("[GetFileStorageService] File details stored for:", fileId);
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to store file details:",
        error,
      );
    }
  }

  // Get file details from cache
  getFileDetails(fileId) {
    try {
      const allDetails = this.getAllFileDetails();
      const fileDetails = allDetails[fileId];

      if (!fileDetails) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(fileDetails.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[GetFileStorageService] File details cache expired for:",
          fileId,
        );
        this.removeFileDetails(fileId);
        return null;
      }

      console.log(
        "[GetFileStorageService] File details retrieved from cache:",
        fileId,
      );

      return fileDetails;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get file details:",
        error,
      );
      return null;
    }
  }

  // Get all stored file details
  getAllFileDetails() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_DETAILS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get all file details:",
        error,
      );
      return {};
    }
  }

  // Remove file details from cache
  removeFileDetails(fileId) {
    try {
      const allDetails = this.getAllFileDetails();
      delete allDetails[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_DETAILS,
        JSON.stringify(allDetails),
      );

      console.log("[GetFileStorageService] File details removed for:", fileId);
      return true;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to remove file details:",
        error,
      );
      return false;
    }
  }

  // === File Version History Storage ===

  // Store file version history
  storeFileVersionHistory(fileId, versionHistory, metadata = {}) {
    try {
      const versionData = {
        fileId,
        versionHistory: versionHistory.map((v) =>
          this.sanitizeFileDetailsForStorage(v),
        ),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(
            Date.now() + this.EXTENDED_CACHE_DURATION,
          ).toISOString(),
          version_count: versionHistory.length,
        },
      };

      const existingVersions = this.getAllFileVersions();
      existingVersions[fileId] = versionData;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_VERSIONS,
        JSON.stringify(existingVersions),
      );

      console.log(
        "[GetFileStorageService] Version history stored for:",
        fileId,
        "versions:",
        versionHistory.length,
      );
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to store version history:",
        error,
      );
    }
  }

  // Get file version history from cache
  getFileVersionHistory(fileId) {
    try {
      const allVersions = this.getAllFileVersions();
      const versionData = allVersions[fileId];

      if (!versionData) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(versionData.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[GetFileStorageService] Version history cache expired for:",
          fileId,
        );
        this.removeFileVersionHistory(fileId);
        return null;
      }

      console.log(
        "[GetFileStorageService] Version history retrieved from cache:",
        fileId,
        "versions:",
        versionData.versionHistory.length,
      );

      return versionData;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get version history:",
        error,
      );
      return null;
    }
  }

  // Get all stored version histories
  getAllFileVersions() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_VERSIONS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get all version histories:",
        error,
      );
      return {};
    }
  }

  // Remove version history from cache
  removeFileVersionHistory(fileId) {
    try {
      const allVersions = this.getAllFileVersions();
      delete allVersions[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_VERSIONS,
        JSON.stringify(allVersions),
      );

      console.log(
        "[GetFileStorageService] Version history removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to remove version history:",
        error,
      );
      return false;
    }
  }

  // === File Permissions Storage ===

  // Store file permissions
  storeFilePermissions(fileId, permissions, metadata = {}) {
    try {
      const permissionsData = {
        fileId,
        permissions,
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingPermissions = this.getAllFilePermissions();
      existingPermissions[fileId] = permissionsData;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_PERMISSIONS,
        JSON.stringify(existingPermissions),
      );

      console.log(
        "[GetFileStorageService] File permissions stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to store file permissions:",
        error,
      );
    }
  }

  // Get file permissions from cache
  getFilePermissions(fileId) {
    try {
      const allPermissions = this.getAllFilePermissions();
      const permissionsData = allPermissions[fileId];

      if (!permissionsData) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(permissionsData.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[GetFileStorageService] File permissions cache expired for:",
          fileId,
        );
        this.removeFilePermissions(fileId);
        return null;
      }

      console.log(
        "[GetFileStorageService] File permissions retrieved from cache:",
        fileId,
      );

      return permissionsData;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get file permissions:",
        error,
      );
      return null;
    }
  }

  // Get all stored file permissions
  getAllFilePermissions() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_PERMISSIONS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get all file permissions:",
        error,
      );
      return {};
    }
  }

  // Remove file permissions from cache
  removeFilePermissions(fileId) {
    try {
      const allPermissions = this.getAllFilePermissions();
      delete allPermissions[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_PERMISSIONS,
        JSON.stringify(allPermissions),
      );

      console.log(
        "[GetFileStorageService] File permissions removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to remove file permissions:",
        error,
      );
      return false;
    }
  }

  // === File Statistics Storage ===

  // Store file statistics
  storeFileStats(fileId, stats, metadata = {}) {
    try {
      const statsData = {
        fileId,
        stats,
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingStats = this.getAllFileStats();
      existingStats[fileId] = statsData;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_STATS,
        JSON.stringify(existingStats),
      );

      console.log(
        "[GetFileStorageService] File statistics stored for:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to store file statistics:",
        error,
      );
    }
  }

  // Get file statistics from cache
  getFileStats(fileId) {
    try {
      const allStats = this.getAllFileStats();
      const statsData = allStats[fileId];

      if (!statsData) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(statsData.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[GetFileStorageService] File statistics cache expired for:",
          fileId,
        );
        this.removeFileStats(fileId);
        return null;
      }

      console.log(
        "[GetFileStorageService] File statistics retrieved from cache:",
        fileId,
      );

      return statsData;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get file statistics:",
        error,
      );
      return null;
    }
  }

  // Get all stored file statistics
  getAllFileStats() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_STATS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to get all file statistics:",
        error,
      );
      return {};
    }
  }

  // Remove file statistics from cache
  removeFileStats(fileId) {
    try {
      const allStats = this.getAllFileStats();
      delete allStats[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_STATS,
        JSON.stringify(allStats),
      );

      console.log(
        "[GetFileStorageService] File statistics removed for:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to remove file statistics:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Check if file details cache is valid
  hasValidFileDetailsCache(fileId) {
    const fileDetails = this.getFileDetails(fileId);
    return fileDetails !== null;
  }

  // Check if version history cache is valid
  hasValidVersionHistoryCache(fileId) {
    const versionHistory = this.getFileVersionHistory(fileId);
    return versionHistory !== null;
  }

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired file details
      const allDetails = this.getAllFileDetails();
      Object.keys(allDetails).forEach((fileId) => {
        const details = allDetails[fileId];
        const expiresAt = new Date(details.metadata.expires_at);
        if (now > expiresAt) {
          delete allDetails[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.FILE_DETAILS,
          JSON.stringify(allDetails),
        );
      }

      // Clear expired version histories
      const allVersions = this.getAllFileVersions();
      Object.keys(allVersions).forEach((fileId) => {
        const versions = allVersions[fileId];
        const expiresAt = new Date(versions.metadata.expires_at);
        if (now > expiresAt) {
          delete allVersions[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.FILE_VERSIONS,
          JSON.stringify(allVersions),
        );
      }

      // Clear expired permissions
      const allPermissions = this.getAllFilePermissions();
      Object.keys(allPermissions).forEach((fileId) => {
        const permissions = allPermissions[fileId];
        const expiresAt = new Date(permissions.metadata.expires_at);
        if (now > expiresAt) {
          delete allPermissions[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.FILE_PERMISSIONS,
          JSON.stringify(allPermissions),
        );
      }

      // Clear expired statistics
      const allStats = this.getAllFileStats();
      Object.keys(allStats).forEach((fileId) => {
        const stats = allStats[fileId];
        const expiresAt = new Date(stats.metadata.expires_at);
        if (now > expiresAt) {
          delete allStats[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.FILE_STATS,
          JSON.stringify(allStats),
        );
      }

      console.log(
        "[GetFileStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all file detail caches
  clearAllFileDetailCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.FILE_DETAILS);
      localStorage.removeItem(this.STORAGE_KEYS.FILE_VERSIONS);
      localStorage.removeItem(this.STORAGE_KEYS.FILE_PERMISSIONS);
      localStorage.removeItem(this.STORAGE_KEYS.FILE_STATS);
      console.log("[GetFileStorageService] All file detail caches cleared");
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to clear file detail caches:",
        error,
      );
    }
  }

  // Clear cache for specific file
  clearFileCache(fileId) {
    try {
      this.removeFileDetails(fileId);
      this.removeFileVersionHistory(fileId);
      this.removeFilePermissions(fileId);
      this.removeFileStats(fileId);

      console.log("[GetFileStorageService] Cache cleared for file:", fileId);
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to clear file cache:",
        error,
      );
    }
  }

  // === Data Sanitization ===

  // Sanitize file details for storage (remove sensitive data)
  sanitizeFileDetailsForStorage(fileDetails) {
    const sanitized = { ...fileDetails };

    // Remove sensitive data that shouldn't be stored
    delete sanitized._file_key; // Never store file keys
    delete sanitized._collection_key; // Never store collection keys
    delete sanitized._decrypted_content; // Never store decrypted content

    // Keep metadata and computed properties
    return sanitized;
  }

  // === Batch Operations ===

  // Store multiple file details at once
  storeMultipleFileDetails(fileDetailsMap) {
    try {
      const allDetails = this.getAllFileDetails();
      let storedCount = 0;

      Object.entries(fileDetailsMap).forEach(([fileId, details]) => {
        const sanitizedDetails = this.sanitizeFileDetailsForStorage(details);
        allDetails[fileId] = {
          fileId,
          fileDetails: sanitizedDetails,
          metadata: {
            cached_at: new Date().toISOString(),
            expires_at: new Date(
              Date.now() + this.CACHE_DURATION,
            ).toISOString(),
          },
        };
        storedCount++;
      });

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_DETAILS,
        JSON.stringify(allDetails),
      );

      console.log(
        "[GetFileStorageService] Batch stored",
        storedCount,
        "file details",
      );
      return storedCount;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to batch store file details:",
        error,
      );
      return 0;
    }
  }

  // Update file details in cache
  updateFileDetails(fileId, updates) {
    try {
      const details = this.getFileDetails(fileId);
      if (!details) {
        console.warn(
          "[GetFileStorageService] File details not found for update:",
          fileId,
        );
        return false;
      }

      const updatedDetails = {
        ...details.fileDetails,
        ...updates,
        updated_at: new Date().toISOString(),
      };

      this.storeFileDetails(fileId, updatedDetails);
      console.log(
        "[GetFileStorageService] File details updated in cache:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[GetFileStorageService] Failed to update file details:",
        error,
      );
      return false;
    }
  }

  // === Configuration ===

  // Set cache duration
  setCacheDuration(duration) {
    this.CACHE_DURATION = duration;
    console.log(
      "[GetFileStorageService] Cache duration set to:",
      duration,
      "ms",
    );
  }

  // Set extended cache duration for version histories
  setExtendedCacheDuration(duration) {
    this.EXTENDED_CACHE_DURATION = duration;
    console.log(
      "[GetFileStorageService] Extended cache duration set to:",
      duration,
      "ms",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const allDetails = this.getAllFileDetails();
    const allVersions = this.getAllFileVersions();
    const allPermissions = this.getAllFilePermissions();
    const allStats = this.getAllFileStats();

    return {
      fileDetailsCount: Object.keys(allDetails).length,
      versionHistoriesCount: Object.keys(allVersions).length,
      permissionsCount: Object.keys(allPermissions).length,
      statisticsCount: Object.keys(allStats).length,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      cacheDuration: this.CACHE_DURATION,
      extendedCacheDuration: this.EXTENDED_CACHE_DURATION,
    };
  }

  // Get debug information
  getDebugInfo() {
    const allDetails = this.getAllFileDetails();
    const allVersions = this.getAllFileVersions();

    return {
      serviceName: "GetFileStorageService",
      storageInfo: this.getStorageInfo(),
      cachedFileIds: Object.keys(allDetails),
      versionHistoryFileIds: Object.keys(allVersions),
      recentActivity: Object.values(allDetails)
        .sort(
          (a, b) =>
            new Date(b.metadata.cached_at) - new Date(a.metadata.cached_at),
        )
        .slice(0, 5)
        .map((details) => ({
          fileId: details.fileId,
          cached_at: details.metadata.cached_at,
        })),
    };
  }
}

export default GetFileStorageService;

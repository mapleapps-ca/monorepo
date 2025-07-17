// File: monorepo/web/maplefile-frontend/src/services/Storage/File/RecentFileStorageService.js
// Recent File Storage Service - Handles localStorage operations for recent files and pagination

class RecentFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      RECENT_FILES: "mapleapps_recent_files",
      RECENT_FILES_METADATA: "mapleapps_recent_files_metadata",
      RECENT_FILES_CURSORS: "mapleapps_recent_files_cursors",
    };

    // Cache configuration
    this.CACHE_DURATION = 10 * 60 * 1000; // 10 minutes (recent files change more frequently)

    console.log("[RecentFileStorageService] Storage service initialized");
  }

  // === Recent Files Storage Operations ===

  // Store recent files with pagination metadata
  storeRecentFiles(files, metadata = {}) {
    try {
      const recentFilesData = {
        files: files.map((file) => this.sanitizeFileForStorage(file)),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          count: files.length,
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      localStorage.setItem(
        this.STORAGE_KEYS.RECENT_FILES,
        JSON.stringify(recentFilesData),
      );

      console.log(
        "[RecentFileStorageService] Recent files stored:",
        files.length,
        "files",
      );
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to store recent files:",
        error,
      );
    }
  }

  // Get stored recent files
  getRecentFiles() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.RECENT_FILES);
      if (!stored) {
        return null;
      }

      const recentFilesData = JSON.parse(stored);

      // Check if cache has expired
      const expiresAt = new Date(recentFilesData.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log("[RecentFileStorageService] Recent files cache expired");
        this.clearRecentFiles();
        return null;
      }

      console.log(
        "[RecentFileStorageService] Recent files retrieved from cache:",
        recentFilesData.files.length,
        "files",
      );

      return recentFilesData;
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to get recent files:",
        error,
      );
      return null;
    }
  }

  // Clear recent files cache
  clearRecentFiles() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES);
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES_METADATA);
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES_CURSORS);
      console.log("[RecentFileStorageService] Recent files cache cleared");
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to clear recent files:",
        error,
      );
    }
  }

  // === Pagination Cursor Management ===

  // Store pagination cursors for recent files
  storePaginationCursors(cursors) {
    try {
      const cursorData = {
        cursors,
        stored_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
      };

      localStorage.setItem(
        this.STORAGE_KEYS.RECENT_FILES_CURSORS,
        JSON.stringify(cursorData),
      );

      console.log(
        "[RecentFileStorageService] Pagination cursors stored:",
        Object.keys(cursors).length,
        "cursors",
      );
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to store pagination cursors:",
        error,
      );
    }
  }

  // Get pagination cursors
  getPaginationCursors() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.RECENT_FILES_CURSORS,
      );
      if (!stored) {
        return {};
      }

      const cursorData = JSON.parse(stored);

      // Check if cache has expired
      const expiresAt = new Date(cursorData.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[RecentFileStorageService] Pagination cursors cache expired",
        );
        localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES_CURSORS);
        return {};
      }

      return cursorData.cursors || {};
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to get pagination cursors:",
        error,
      );
      return {};
    }
  }

  // Store cursor for specific page
  storeCursorForPage(pageKey, cursor) {
    const cursors = this.getPaginationCursors();
    cursors[pageKey] = cursor;
    this.storePaginationCursors(cursors);
  }

  // Get cursor for specific page
  getCursorForPage(pageKey) {
    const cursors = this.getPaginationCursors();
    return cursors[pageKey] || null;
  }

  // === Individual File Cache Operations ===

  // Store individual file metadata
  storeFile(file) {
    try {
      // We'll leverage the existing file cache from ListFileStorageService pattern
      const sanitizedFile = this.sanitizeFileForStorage(file);

      // For simplicity, we'll store recent files in the same individual file cache
      // This allows for easy retrieval and avoids duplication
      const fileCache = this.getAllFiles();

      fileCache[file.id] = {
        ...sanitizedFile,
        cached_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        source: "recent_files", // Mark as coming from recent files
      };

      localStorage.setItem(
        this.STORAGE_KEYS.RECENT_FILES_METADATA,
        JSON.stringify(fileCache),
      );

      console.log("[RecentFileStorageService] File cached:", file.id);
    } catch (error) {
      console.error("[RecentFileStorageService] Failed to store file:", error);
    }
  }

  // Get individual file from cache
  getFile(fileId) {
    try {
      const fileCache = this.getAllFiles();
      const cachedFile = fileCache[fileId];

      if (!cachedFile) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(cachedFile.expires_at);
      if (new Date() > expiresAt) {
        console.log("[RecentFileStorageService] File cache expired:", fileId);
        this.removeFile(fileId);
        return null;
      }

      console.log(
        "[RecentFileStorageService] File retrieved from cache:",
        fileId,
      );
      return cachedFile;
    } catch (error) {
      console.error("[RecentFileStorageService] Failed to get file:", error);
      return null;
    }
  }

  // Get all cached files
  getAllFiles() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.RECENT_FILES_METADATA,
      );
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to get all files:",
        error,
      );
      return {};
    }
  }

  // Remove individual file from cache
  removeFile(fileId) {
    try {
      const fileCache = this.getAllFiles();
      delete fileCache[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.RECENT_FILES_METADATA,
        JSON.stringify(fileCache),
      );

      console.log(
        "[RecentFileStorageService] File removed from cache:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error("[RecentFileStorageService] Failed to remove file:", error);
      return false;
    }
  }

  // === Cache Validation ===

  // Check if recent files cache is valid
  hasValidRecentFilesCache() {
    const recentFiles = this.getRecentFiles();
    return recentFiles !== null;
  }

  // Get cache age in minutes
  getCacheAge() {
    const recentFiles = this.getRecentFiles();
    if (!recentFiles) {
      return null;
    }

    const cachedAt = new Date(recentFiles.metadata.cached_at);
    const now = new Date();
    return Math.floor((now - cachedAt) / (1000 * 60));
  }

  // === Cache Management ===

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired recent files
      const recentFiles = this.getRecentFiles();
      if (recentFiles) {
        const expiresAt = new Date(recentFiles.metadata.expires_at);
        if (now > expiresAt) {
          this.clearRecentFiles();
          clearedCount++;
        }
      }

      // Clear expired individual files
      const allFiles = this.getAllFiles();
      Object.keys(allFiles).forEach((fileId) => {
        const file = allFiles[fileId];
        const expiresAt = new Date(file.expires_at);
        if (now > expiresAt) {
          delete allFiles[fileId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.RECENT_FILES_METADATA,
          JSON.stringify(allFiles),
        );
      }

      console.log(
        "[RecentFileStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all caches
  clearAllCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES);
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES_METADATA);
      localStorage.removeItem(this.STORAGE_KEYS.RECENT_FILES_CURSORS);
      console.log("[RecentFileStorageService] All caches cleared");
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to clear all caches:",
        error,
      );
    }
  }

  // === Data Sanitization ===

  // Sanitize file for storage (remove sensitive data)
  sanitizeFileForStorage(file) {
    const sanitized = { ...file };

    // Remove sensitive data that shouldn't be stored
    delete sanitized._file_key; // Never store file keys
    delete sanitized._collection_key; // Never store collection keys
    delete sanitized._decrypted_content; // Never store decrypted content

    // Keep metadata and computed properties
    return sanitized;
  }

  // === Batch Operations ===

  // Store multiple files at once
  storeFiles(files) {
    try {
      const fileCache = this.getAllFiles();
      let storedCount = 0;

      files.forEach((file) => {
        const sanitizedFile = this.sanitizeFileForStorage(file);
        fileCache[file.id] = {
          ...sanitizedFile,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
          source: "recent_files",
        };
        storedCount++;
      });

      localStorage.setItem(
        this.STORAGE_KEYS.RECENT_FILES_METADATA,
        JSON.stringify(fileCache),
      );

      console.log(
        "[RecentFileStorageService] Batch stored",
        storedCount,
        "files",
      );
      return storedCount;
    } catch (error) {
      console.error(
        "[RecentFileStorageService] Failed to batch store files:",
        error,
      );
      return 0;
    }
  }

  // === Configuration ===

  // Set cache duration
  setCacheDuration(duration) {
    this.CACHE_DURATION = duration;
    console.log(
      "[RecentFileStorageService] Cache duration set to:",
      duration,
      "ms",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const recentFiles = this.getRecentFiles();
    const allFiles = this.getAllFiles();
    const cursors = this.getPaginationCursors();

    return {
      hasRecentFilesCache: !!recentFiles,
      recentFilesCount: recentFiles?.files?.length || 0,
      individualFilesCount: Object.keys(allFiles).length,
      paginationCursorsCount: Object.keys(cursors).length,
      cacheDuration: this.CACHE_DURATION,
      cacheAge: this.getCacheAge(),
      isValid: this.hasValidRecentFilesCache(),
    };
  }

  // Get debug information
  getDebugInfo() {
    const recentFiles = this.getRecentFiles();
    const allFiles = this.getAllFiles();
    const cursors = this.getPaginationCursors();

    return {
      serviceName: "RecentFileStorageService",
      storageInfo: this.getStorageInfo(),
      cachedFileIds: Object.keys(allFiles).slice(0, 10), // Show first 10
      paginationKeys: Object.keys(cursors),
      lastCached: recentFiles?.metadata?.cached_at || null,
      cacheExpiry: recentFiles?.metadata?.expires_at || null,
    };
  }
}

export default RecentFileStorageService;

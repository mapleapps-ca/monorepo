// File: monorepo/web/maplefile-frontend/src/services/Storage/File/ListFileStorageService.js
// List File Storage Service - Handles localStorage operations for file lists and caching

class ListFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      FILE_LISTS: "mapleapps_file_lists",
      FILE_CACHE: "mapleapps_file_cache",
      FILE_METADATA: "mapleapps_file_metadata",
    };

    // Cache configuration
    this.CACHE_DURATION = 15 * 60 * 1000; // 15 minutes default

    console.log("[ListFileStorageService] Storage service initialized");
  }

  // === File List Storage Operations ===

  // Store file list for a collection
  storeFileList(collectionId, files, metadata = {}) {
    try {
      const fileListData = {
        collectionId,
        files: files.map((file) => this.sanitizeFileForStorage(file)),
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          count: files.length,
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      const existingLists = this.getAllFileLists();
      existingLists[collectionId] = fileListData;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_LISTS,
        JSON.stringify(existingLists),
      );

      console.log(
        "[ListFileStorageService] File list stored for collection:",
        collectionId,
        "count:",
        files.length,
      );
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to store file list:",
        error,
      );
    }
  }

  // Get file list for a collection
  getFileList(collectionId) {
    try {
      const allLists = this.getAllFileLists();
      const fileList = allLists[collectionId];

      if (!fileList) {
        return null;
      }

      // Check if cache has expired
      const expiresAt = new Date(fileList.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log(
          "[ListFileStorageService] File list cache expired for collection:",
          collectionId,
        );
        this.removeFileList(collectionId);
        return null;
      }

      console.log(
        "[ListFileStorageService] File list retrieved from cache:",
        collectionId,
        "count:",
        fileList.files.length,
      );

      return fileList;
    } catch (error) {
      console.error("[ListFileStorageService] Failed to get file list:", error);
      return null;
    }
  }

  // Get all stored file lists
  getAllFileLists() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_LISTS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to get all file lists:",
        error,
      );
      return {};
    }
  }

  // Remove file list for a collection
  removeFileList(collectionId) {
    try {
      const allLists = this.getAllFileLists();
      delete allLists[collectionId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_LISTS,
        JSON.stringify(allLists),
      );

      console.log(
        "[ListFileStorageService] File list removed for collection:",
        collectionId,
      );
      return true;
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to remove file list:",
        error,
      );
      return false;
    }
  }

  // === Individual File Cache Operations ===

  // Store individual file metadata
  storeFile(file) {
    try {
      const sanitizedFile = this.sanitizeFileForStorage(file);
      const fileCache = this.getAllFiles();

      fileCache[file.id] = {
        ...sanitizedFile,
        cached_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
      };

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_CACHE,
        JSON.stringify(fileCache),
      );

      console.log("[ListFileStorageService] File cached:", file.id);
    } catch (error) {
      console.error("[ListFileStorageService] Failed to store file:", error);
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
        console.log("[ListFileStorageService] File cache expired:", fileId);
        this.removeFile(fileId);
        return null;
      }

      console.log(
        "[ListFileStorageService] File retrieved from cache:",
        fileId,
      );
      return cachedFile;
    } catch (error) {
      console.error("[ListFileStorageService] Failed to get file:", error);
      return null;
    }
  }

  // Get all cached files
  getAllFiles() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_CACHE);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error("[ListFileStorageService] Failed to get all files:", error);
      return {};
    }
  }

  // Remove individual file from cache
  removeFile(fileId) {
    try {
      const fileCache = this.getAllFiles();
      delete fileCache[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_CACHE,
        JSON.stringify(fileCache),
      );

      console.log("[ListFileStorageService] File removed from cache:", fileId);
      return true;
    } catch (error) {
      console.error("[ListFileStorageService] Failed to remove file:", error);
      return false;
    }
  }

  // === File Statistics and Queries ===

  // Get file statistics for a collection
  getFileStats(collectionId) {
    const fileList = this.getFileList(collectionId);

    if (!fileList) {
      return {
        total: 0,
        active: 0,
        archived: 0,
        deleted: 0,
        pending: 0,
        withTombstones: 0,
        expiredTombstones: 0,
        recent: 0,
      };
    }

    const files = fileList.files;
    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    const stats = {
      total: files.length,
      active: 0,
      archived: 0,
      deleted: 0,
      pending: 0,
      withTombstones: 0,
      expiredTombstones: 0,
      recent: 0,
    };

    files.forEach((file) => {
      // Count by state
      if (file._is_active) stats.active++;
      if (file._is_archived) stats.archived++;
      if (file._is_deleted) stats.deleted++;
      if (file._is_pending) stats.pending++;

      // Count tombstones
      if (file._has_tombstone) stats.withTombstones++;
      if (file._tombstone_expired) stats.expiredTombstones++;

      // Count recent files (created in last 24 hours)
      const createdAt = new Date(file.created_at || file.stored_at).getTime();
      if (createdAt > oneDayAgo) stats.recent++;
    });

    return stats;
  }

  // Get files by state for a collection
  getFilesByState(collectionId, states = ["active"]) {
    const fileList = this.getFileList(collectionId);

    if (!fileList) {
      return [];
    }

    return fileList.files.filter((file) => states.includes(file.state));
  }

  // Get files with specific properties
  getFilesByProperty(collectionId, propertyName, propertyValue) {
    const fileList = this.getFileList(collectionId);

    if (!fileList) {
      return [];
    }

    return fileList.files.filter(
      (file) => file[propertyName] === propertyValue,
    );
  }

  // === Cache Management ===

  // Check if file list cache is valid for a collection
  hasValidFileListCache(collectionId) {
    const fileList = this.getFileList(collectionId);
    return fileList !== null;
  }

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired file lists
      const allLists = this.getAllFileLists();
      Object.keys(allLists).forEach((collectionId) => {
        const fileList = allLists[collectionId];
        const expiresAt = new Date(fileList.metadata.expires_at);
        if (now > expiresAt) {
          delete allLists[collectionId];
          clearedCount++;
        }
      });

      if (clearedCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.FILE_LISTS,
          JSON.stringify(allLists),
        );
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
          this.STORAGE_KEYS.FILE_CACHE,
          JSON.stringify(allFiles),
        );
      }

      console.log(
        "[ListFileStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all file caches
  clearAllFileCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.FILE_LISTS);
      localStorage.removeItem(this.STORAGE_KEYS.FILE_CACHE);
      console.log("[ListFileStorageService] All file caches cleared");
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to clear file caches:",
        error,
      );
    }
  }

  // Clear cache for specific collection
  clearCollectionCache(collectionId) {
    try {
      // Remove file list
      this.removeFileList(collectionId);

      // Remove individual files belonging to this collection
      const allFiles = this.getAllFiles();
      const filesToRemove = Object.keys(allFiles).filter(
        (fileId) => allFiles[fileId].collection_id === collectionId,
      );

      filesToRemove.forEach((fileId) => {
        delete allFiles[fileId];
      });

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_CACHE,
        JSON.stringify(allFiles),
      );

      console.log(
        "[ListFileStorageService] Cache cleared for collection:",
        collectionId,
        "removed",
        filesToRemove.length,
        "files",
      );
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to clear collection cache:",
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
        };
        storedCount++;
      });

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_CACHE,
        JSON.stringify(fileCache),
      );

      console.log(
        "[ListFileStorageService] Batch stored",
        storedCount,
        "files",
      );
      return storedCount;
    } catch (error) {
      console.error(
        "[ListFileStorageService] Failed to batch store files:",
        error,
      );
      return 0;
    }
  }

  // Update file in cache
  updateFile(fileId, updates) {
    try {
      const file = this.getFile(fileId);
      if (!file) {
        console.warn(
          "[ListFileStorageService] File not found for update:",
          fileId,
        );
        return false;
      }

      const updatedFile = {
        ...file,
        ...updates,
        updated_at: new Date().toISOString(),
      };

      this.storeFile(updatedFile);
      console.log("[ListFileStorageService] File updated in cache:", fileId);
      return true;
    } catch (error) {
      console.error("[ListFileStorageService] Failed to update file:", error);
      return false;
    }
  }

  // === Configuration ===

  // Set cache duration
  setCacheDuration(duration) {
    this.CACHE_DURATION = duration;
    console.log(
      "[ListFileStorageService] Cache duration set to:",
      duration,
      "ms",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const allLists = this.getAllFileLists();
    const allFiles = this.getAllFiles();

    return {
      fileListsCount: Object.keys(allLists).length,
      individualFilesCount: Object.keys(allFiles).length,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      cacheDuration: this.CACHE_DURATION,
      totalCollections: Object.keys(allLists).length,
    };
  }

  // Get debug information
  getDebugInfo() {
    const allLists = this.getAllFileLists();
    const allFiles = this.getAllFiles();

    return {
      serviceName: "ListFileStorageService",
      storageInfo: this.getStorageInfo(),
      cachedCollections: Object.keys(allLists),
      cachedFileIds: Object.keys(allFiles).slice(0, 10), // Show first 10
      recentActivity: Object.values(allLists)
        .sort(
          (a, b) =>
            new Date(b.metadata.cached_at) - new Date(a.metadata.cached_at),
        )
        .slice(0, 5)
        .map((list) => ({
          collectionId: list.collectionId,
          count: list.files.length,
          cached_at: list.metadata.cached_at,
        })),
    };
  }
}

export default ListFileStorageService;

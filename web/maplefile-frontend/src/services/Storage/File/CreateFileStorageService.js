// File: monorepo/web/maplefile-frontend/src/services/Storage/File/CreateFileStorageService.js
// Create File Storage Service - Handles localStorage operations for pending files

class CreateFileStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      PENDING_FILES: "mapleapps_pending_files",
      FILE_UPLOAD_QUEUE: "mapleapps_file_upload_queue",
    };

    console.log("[CreateFileStorageService] Storage service initialized");
  }

  // === Pending File Storage Operations ===

  // Store pending file info
  storePendingFile(pendingFileInfo) {
    try {
      const existingFiles = this.getPendingFiles();

      // Remove any existing file with same ID
      const filteredFiles = existingFiles.filter(
        (f) => f.file.id !== pendingFileInfo.file.id,
      );

      // Add the new pending file
      filteredFiles.push({
        ...pendingFileInfo,
        stored_at: new Date().toISOString(),
      });

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_FILES,
        JSON.stringify(filteredFiles),
      );
      console.log(
        "[CreateFileStorageService] Pending file stored:",
        pendingFileInfo.file.id,
      );
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to store pending file:",
        error,
      );
    }
  }

  // Get all pending files
  getPendingFiles() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.PENDING_FILES);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to get pending files:",
        error,
      );
      return [];
    }
  }

  // Get pending file by ID
  getPendingFileById(fileId) {
    const files = this.getPendingFiles();
    return files.find((f) => f.file.id === fileId) || null;
  }

  // === Upload Queue Management ===

  // Add file to upload queue
  addToUploadQueue(fileId, uploadInfo) {
    try {
      const queue = this.getUploadQueue();
      queue[fileId] = {
        ...uploadInfo,
        queued_at: new Date().toISOString(),
      };

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_UPLOAD_QUEUE,
        JSON.stringify(queue),
      );
      console.log(
        "[CreateFileStorageService] File added to upload queue:",
        fileId,
      );
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to add to upload queue:",
        error,
      );
    }
  }

  // Get upload queue
  getUploadQueue() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.FILE_UPLOAD_QUEUE);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to get upload queue:",
        error,
      );
      return {};
    }
  }

  // Remove from upload queue
  removeFromUploadQueue(fileId) {
    try {
      const queue = this.getUploadQueue();
      delete queue[fileId];

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_UPLOAD_QUEUE,
        JSON.stringify(queue),
      );
      console.log(
        "[CreateFileStorageService] File removed from upload queue:",
        fileId,
      );
      return true;
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to remove from upload queue:",
        error,
      );
      return false;
    }
  }

  // === File Statistics ===

  // Get file creation statistics
  getFileStats() {
    const pendingFiles = this.getPendingFiles();
    const uploadQueue = this.getUploadQueue();
    const stats = {
      totalPending: pendingFiles.length,
      inUploadQueue: Object.keys(uploadQueue).length,
      recent: 0, // created in last 24 hours
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    pendingFiles.forEach((fileInfo) => {
      const createdAt = new Date(fileInfo.stored_at).getTime();
      if (createdAt > oneDayAgo) {
        stats.recent++;
      }
    });

    return stats;
  }

  // === Data Management ===

  // Update pending file
  updatePendingFile(fileId, updates) {
    try {
      const files = this.getPendingFiles();
      const index = files.findIndex((f) => f.file.id === fileId);

      if (index !== -1) {
        files[index] = {
          ...files[index],
          ...updates,
          updated_at: new Date().toISOString(),
        };

        localStorage.setItem(
          this.STORAGE_KEYS.PENDING_FILES,
          JSON.stringify(files),
        );
        console.log("[CreateFileStorageService] Pending file updated:", fileId);
        return files[index];
      }

      throw new Error(`Pending file not found: ${fileId}`);
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to update pending file:",
        error,
      );
      throw error;
    }
  }

  // Remove pending file from storage
  removePendingFile(fileId) {
    try {
      const files = this.getPendingFiles();
      const filteredFiles = files.filter((f) => f.file.id !== fileId);

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_FILES,
        JSON.stringify(filteredFiles),
      );

      // Also remove from upload queue
      this.removeFromUploadQueue(fileId);

      console.log("[CreateFileStorageService] Pending file removed:", fileId);
      return true;
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to remove pending file:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Clear all pending files
  clearAllPendingFiles() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.PENDING_FILES);
      localStorage.removeItem(this.STORAGE_KEYS.FILE_UPLOAD_QUEUE);
      console.log("[CreateFileStorageService] All pending files cleared");
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to clear pending files:",
        error,
      );
    }
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const pendingFiles = this.getPendingFiles();
    const uploadQueue = this.getUploadQueue();
    const stats = this.getFileStats();

    return {
      pendingFilesCount: pendingFiles.length,
      uploadQueueCount: Object.keys(uploadQueue).length,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasStoredFiles: pendingFiles.length > 0,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CreateFileStorageService",
      storageInfo: this.getStorageInfo(),
      pendingFileIds: this.getPendingFiles().map((f) => f.file.id),
      uploadQueueIds: Object.keys(this.getUploadQueue()),
    };
  }
}

export default CreateFileStorageService;

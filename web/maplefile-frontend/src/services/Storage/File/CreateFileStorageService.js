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

      // Add the new pending file with enhanced metadata
      const enhancedFileInfo = {
        ...pendingFileInfo,
        stored_at: new Date().toISOString(),
        upload_status: pendingFileInfo.upload_completed
          ? "completed"
          : "pending",
        last_updated: new Date().toISOString(),
      };

      filteredFiles.push(enhancedFileInfo);

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_FILES,
        JSON.stringify(filteredFiles),
      );
      console.log(
        "[CreateFileStorageService] Pending file stored:",
        pendingFileInfo.file.id,
        "status:",
        enhancedFileInfo.upload_status,
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

  // Add file to upload queue with enhanced status tracking
  addToUploadQueue(fileId, uploadInfo) {
    try {
      const queue = this.getUploadQueue();

      // Enhanced upload info with timestamps
      const enhancedUploadInfo = {
        ...uploadInfo,
        queued_at: queue[fileId]?.queued_at || new Date().toISOString(),
        last_updated: new Date().toISOString(),
        status_history: [
          ...(queue[fileId]?.status_history || []),
          {
            status: uploadInfo.status,
            timestamp: new Date().toISOString(),
          },
        ].slice(-10), // Keep last 10 status changes
      };

      queue[fileId] = enhancedUploadInfo;

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_UPLOAD_QUEUE,
        JSON.stringify(queue),
      );
      console.log(
        "[CreateFileStorageService] File upload queue updated:",
        fileId,
        "status:",
        uploadInfo.status,
      );
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to update upload queue:",
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

  // Get upload status for a file
  getUploadStatus(fileId) {
    const queue = this.getUploadQueue();
    const queueItem = queue[fileId];

    if (!queueItem) {
      // Check if file is completed
      const pendingFile = this.getPendingFileById(fileId);
      if (pendingFile?.upload_completed) {
        return {
          status: "completed",
          completed_at: pendingFile.completed_at,
        };
      }
      return null;
    }

    return {
      status: queueItem.status,
      queued_at: queueItem.queued_at,
      last_updated: queueItem.last_updated,
      error: queueItem.error,
      status_history: queueItem.status_history,
    };
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
      byStatus: {
        pending: 0,
        uploading: 0,
        completed: 0,
        error: 0,
      },
      byUploadStatus: {
        pending: 0,
        uploading: 0,
        uploading_thumbnail: 0,
        completing: 0,
        completed: 0,
        error: 0,
      },
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    // Count pending files by status
    pendingFiles.forEach((fileInfo) => {
      const createdAt = new Date(fileInfo.stored_at).getTime();
      if (createdAt > oneDayAgo) {
        stats.recent++;
      }

      const status = fileInfo.upload_status || "pending";
      if (stats.byStatus[status] !== undefined) {
        stats.byStatus[status]++;
      }
    });

    // Count upload queue by status
    Object.values(uploadQueue).forEach((queueItem) => {
      const status = queueItem.status || "pending";
      if (stats.byUploadStatus[status] !== undefined) {
        stats.byUploadStatus[status]++;
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
          last_updated: new Date().toISOString(),
        };

        // Update upload status based on updates
        if (updates.upload_completed) {
          files[index].upload_status = "completed";
        } else if (updates.upload_failed) {
          files[index].upload_status = "error";
        }

        localStorage.setItem(
          this.STORAGE_KEYS.PENDING_FILES,
          JSON.stringify(files),
        );
        console.log(
          "[CreateFileStorageService] Pending file updated:",
          fileId,
          "status:",
          files[index].upload_status,
        );
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

  // === Enhanced Query Methods ===

  // Get files by upload status
  getFilesByUploadStatus(status) {
    const files = this.getPendingFiles();
    return files.filter((f) => f.upload_status === status);
  }

  // Get files by queue status
  getFilesByQueueStatus(status) {
    const queue = this.getUploadQueue();
    const fileIds = Object.keys(queue).filter(
      (id) => queue[id].status === status,
    );
    const files = this.getPendingFiles();

    return files.filter((f) => fileIds.includes(f.file.id));
  }

  // Get completed files
  getCompletedFiles() {
    return this.getFilesByUploadStatus("completed");
  }

  // Get failed files
  getFailedFiles() {
    const uploadFailed = this.getFilesByUploadStatus("error");
    const queueFailed = this.getFilesByQueueStatus("error");

    // Combine and deduplicate
    const allFailed = [...uploadFailed, ...queueFailed];
    const seen = new Set();
    return allFailed.filter((f) => {
      if (seen.has(f.file.id)) return false;
      seen.add(f.file.id);
      return true;
    });
  }

  // Get files currently uploading
  getUploadingFiles() {
    const uploadStatuses = ["uploading", "uploading_thumbnail", "completing"];
    const queue = this.getUploadQueue();
    const uploadingIds = Object.keys(queue).filter((id) =>
      uploadStatuses.includes(queue[id].status),
    );

    const files = this.getPendingFiles();
    return files.filter((f) => uploadingIds.includes(f.file.id));
  }

  // Get files with expired upload URLs
  getFilesWithExpiredUrls() {
    const files = this.getPendingFiles();
    const now = new Date();

    return files.filter((f) => {
      if (!f.upload_url_expiration_time) return false;
      return new Date(f.upload_url_expiration_time) < now;
    });
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

  // Clear completed files
  clearCompletedFiles() {
    try {
      const files = this.getPendingFiles();
      const nonCompletedFiles = files.filter(
        (f) => f.upload_status !== "completed",
      );

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_FILES,
        JSON.stringify(nonCompletedFiles),
      );

      // Remove completed files from upload queue
      const queue = this.getUploadQueue();
      const filteredQueue = {};
      Object.keys(queue).forEach((fileId) => {
        if (queue[fileId].status !== "completed") {
          filteredQueue[fileId] = queue[fileId];
        }
      });

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_UPLOAD_QUEUE,
        JSON.stringify(filteredQueue),
      );

      console.log("[CreateFileStorageService] Completed files cleared");
      return files.length - nonCompletedFiles.length;
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to clear completed files:",
        error,
      );
      return 0;
    }
  }

  // Clear files with expired URLs
  clearExpiredFiles() {
    try {
      const files = this.getPendingFiles();
      const expiredFiles = this.getFilesWithExpiredUrls();
      const nonExpiredFiles = files.filter(
        (f) => !expiredFiles.some((ef) => ef.file.id === f.file.id),
      );

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_FILES,
        JSON.stringify(nonExpiredFiles),
      );

      // Remove expired files from upload queue
      const queue = this.getUploadQueue();
      expiredFiles.forEach((f) => {
        delete queue[f.file.id];
      });

      localStorage.setItem(
        this.STORAGE_KEYS.FILE_UPLOAD_QUEUE,
        JSON.stringify(queue),
      );

      console.log(
        "[CreateFileStorageService] Expired files cleared:",
        expiredFiles.length,
      );
      return expiredFiles.length;
    } catch (error) {
      console.error(
        "[CreateFileStorageService] Failed to clear expired files:",
        error,
      );
      return 0;
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
      uploadStatuses: {
        completed: this.getCompletedFiles().length,
        failed: this.getFailedFiles().length,
        uploading: this.getUploadingFiles().length,
        expired: this.getFilesWithExpiredUrls().length,
      },
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CreateFileStorageService",
      storageInfo: this.getStorageInfo(),
      pendingFileIds: this.getPendingFiles().map((f) => f.file.id),
      uploadQueueIds: Object.keys(this.getUploadQueue()),
      uploadStatuses: this.getFileStats().byUploadStatus,
      recentActivity: this.getPendingFiles()
        .sort((a, b) => new Date(b.stored_at) - new Date(a.stored_at))
        .slice(0, 5)
        .map((f) => ({
          id: f.file.id,
          status: f.upload_status,
          stored_at: f.stored_at,
        })),
    };
  }
}

export default CreateFileStorageService;

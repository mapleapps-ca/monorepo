// Custom hook for file management
// Updated to support version, state, tombstone_version, and tombstone_expiry fields
import { useState, useEffect, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useFiles = (collectionId = null) => {
  const { fileService, authService } = useServices();
  const [files, setFiles] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [uploadQueue, setUploadQueue] = useState(new Map());

  // Load files for a specific collection with state filtering
  const loadFilesByCollection = useCallback(
    async (targetCollectionId, forceRefresh = false, includeStates = null) => {
      if (!authService.isAuthenticated()) {
        console.log("[useFiles] User not authenticated, skipping load");
        return [];
      }

      if (!targetCollectionId) {
        console.log("[useFiles] No collection ID provided, skipping load");
        return [];
      }

      try {
        setIsLoading(true);
        setError(null);

        const fileList = await fileService.listFilesByCollection(
          targetCollectionId,
          forceRefresh,
          includeStates,
        );
        setFiles(fileList);

        return fileList;
      } catch (err) {
        console.error("[useFiles] Failed to load files:", err);
        setError(err.message);
        return [];
      } finally {
        setIsLoading(false);
      }
    },
    [fileService, authService],
  );

  // Create and upload a new file
  const uploadFile = useCallback(
    async (fileData, encryptedContent, encryptedThumbnail = null) => {
      try {
        setIsLoading(true);
        setError(null);

        // Add to upload queue
        setUploadQueue((prev) =>
          new Map(prev).set(fileData.id, {
            id: fileData.id,
            status: "uploading",
            progress: 0,
            version: fileData.version || 1,
          }),
        );

        const uploadedFile = await fileService.uploadFile(
          fileData,
          encryptedContent,
          encryptedThumbnail,
        );

        // Remove from upload queue
        setUploadQueue((prev) => {
          const next = new Map(prev);
          next.delete(fileData.id);
          return next;
        });

        // Add to local state if it belongs to the current collection
        if (collectionId && uploadedFile.collection_id === collectionId) {
          setFiles((prev) => [...prev, uploadedFile]);
        }

        return uploadedFile;
      } catch (err) {
        console.error("[useFiles] Failed to upload file:", err);
        setError(err.message);

        // Update upload queue with error
        setUploadQueue((prev) => {
          const next = new Map(prev);
          next.set(fileData.id, {
            id: fileData.id,
            status: "error",
            error: err.message,
          });
          return next;
        });

        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService, collectionId],
  );

  // Download a file
  const downloadFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        const result = await fileService.downloadFile(fileId);

        console.log("[useFiles] File downloaded successfully");
        return result;
      } catch (err) {
        console.error("[useFiles] Failed to download file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Update file metadata with version support for optimistic locking
  const updateFile = useCallback(
    async (fileId, updateData) => {
      try {
        setIsLoading(true);
        setError(null);

        // Get current file to ensure we have the latest version
        const currentFile = fileService.getCachedFile(fileId);
        if (!currentFile && !updateData.version) {
          // If not in cache and no version provided, fetch first
          const freshFile = await fileService.getFile(fileId);
          updateData.version = freshFile.version;
        } else if (currentFile && !updateData.version) {
          updateData.version = currentFile.version;
        }

        console.log(
          "[useFiles] Updating file with version:",
          updateData.version,
        );

        const updatedFile = await fileService.updateFile(fileId, updateData);

        // Update local state
        setFiles((prev) =>
          prev.map((file) => (file.id === fileId ? updatedFile : file)),
        );

        return updatedFile;
      } catch (err) {
        console.error("[useFiles] Failed to update file:", err);
        setError(err.message);

        // Check for version conflict and provide better error message
        if (
          err.message?.includes("version") ||
          err.message?.includes("conflict")
        ) {
          setError(
            "File has been updated by another process. Please refresh and try again.",
          );
        }

        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Soft delete a file (creates tombstone)
  const deleteFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        await fileService.deleteFile(fileId);

        // Update local state - mark as deleted with tombstone
        setFiles((prev) =>
          prev.map((file) => {
            if (file.id === fileId) {
              const newVersion = (file.version || 1) + 1;
              return {
                ...file,
                state: fileService.FILE_STATES.DELETED,
                version: newVersion,
                tombstone_version: newVersion,
                tombstone_expiry: new Date(
                  Date.now() + 30 * 24 * 60 * 60 * 1000,
                ).toISOString(),
                _is_deleted: true,
                _has_tombstone: true,
              };
            }
            return file;
          }),
        );

        return true;
      } catch (err) {
        console.error("[useFiles] Failed to delete file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Delete multiple files
  const deleteMultipleFiles = useCallback(
    async (fileIds) => {
      try {
        setIsLoading(true);
        setError(null);

        const result = await fileService.deleteMultipleFiles(fileIds);

        // Update local state - mark as deleted with tombstones
        setFiles((prev) =>
          prev.map((file) => {
            if (fileIds.includes(file.id)) {
              const newVersion = (file.version || 1) + 1;
              return {
                ...file,
                state: fileService.FILE_STATES.DELETED,
                version: newVersion,
                tombstone_version: newVersion,
                tombstone_expiry: new Date(
                  Date.now() + 30 * 24 * 60 * 60 * 1000,
                ).toISOString(),
                _is_deleted: true,
                _has_tombstone: true,
              };
            }
            return file;
          }),
        );

        return result;
      } catch (err) {
        console.error("[useFiles] Failed to delete multiple files:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Archive a file
  const archiveFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        await fileService.archiveFile(fileId);

        // Update local state
        setFiles((prev) =>
          prev.map((file) => {
            if (file.id === fileId) {
              return {
                ...file,
                state: fileService.FILE_STATES.ARCHIVED,
                version: (file.version || 1) + 1,
                _is_archived: true,
                _is_active: false,
              };
            }
            return file;
          }),
        );

        return true;
      } catch (err) {
        console.error("[useFiles] Failed to archive file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Restore a file (unarchive or restore from deletion)
  const restoreFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        await fileService.restoreFile(fileId);

        // Update local state
        setFiles((prev) =>
          prev.map((file) => {
            if (file.id === fileId) {
              return {
                ...file,
                state: fileService.FILE_STATES.ACTIVE,
                version: (file.version || 1) + 1,
                tombstone_version: 0,
                tombstone_expiry: "0001-01-01T00:00:00Z",
                _is_active: true,
                _is_archived: false,
                _is_deleted: false,
                _has_tombstone: false,
              };
            }
            return file;
          }),
        );

        return true;
      } catch (err) {
        console.error("[useFiles] Failed to restore file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Batch archive files
  const batchArchiveFiles = useCallback(
    async (fileIds) => {
      try {
        setIsLoading(true);
        setError(null);

        const result = await fileService.batchArchiveFiles(fileIds);

        // Update local state for successful archives
        const successfulIds = fileIds.filter(
          (id, index) => !result.failed.some((f) => f.fileId === id),
        );

        setFiles((prev) =>
          prev.map((file) => {
            if (successfulIds.includes(file.id)) {
              return {
                ...file,
                state: fileService.FILE_STATES.ARCHIVED,
                version: (file.version || 1) + 1,
                _is_archived: true,
                _is_active: false,
              };
            }
            return file;
          }),
        );

        return result;
      } catch (err) {
        console.error("[useFiles] Failed to batch archive files:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Batch restore files
  const batchRestoreFiles = useCallback(
    async (fileIds) => {
      try {
        setIsLoading(true);
        setError(null);

        const result = await fileService.batchRestoreFiles(fileIds);

        // Update local state for successful restores
        const successfulIds = fileIds.filter(
          (id, index) => !result.failed.some((f) => f.fileId === id),
        );

        setFiles((prev) =>
          prev.map((file) => {
            if (successfulIds.includes(file.id)) {
              return {
                ...file,
                state: fileService.FILE_STATES.ACTIVE,
                version: (file.version || 1) + 1,
                tombstone_version: 0,
                tombstone_expiry: "0001-01-01T00:00:00Z",
                _is_active: true,
                _is_archived: false,
                _is_deleted: false,
                _has_tombstone: false,
              };
            }
            return file;
          }),
        );

        return result;
      } catch (err) {
        console.error("[useFiles] Failed to batch restore files:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Get files by state
  const getFilesByState = useCallback(
    (state = fileService.FILE_STATES.ACTIVE) => {
      return files.filter((file) => file.state === state);
    },
    [files, fileService.FILE_STATES],
  );

  // Get files by multiple states
  const getFilesByStates = useCallback(
    (states = [fileService.FILE_STATES.ACTIVE]) => {
      return files.filter((file) => states.includes(file.state));
    },
    [files, fileService.FILE_STATES],
  );

  // Get active files
  const getActiveFiles = useCallback(() => {
    return getFilesByState(fileService.FILE_STATES.ACTIVE);
  }, [getFilesByState, fileService.FILE_STATES]);

  // Get archived files
  const getArchivedFiles = useCallback(() => {
    return getFilesByState(fileService.FILE_STATES.ARCHIVED);
  }, [getFilesByState, fileService.FILE_STATES]);

  // Get deleted files (tombstones)
  const getDeletedFiles = useCallback(() => {
    return getFilesByState(fileService.FILE_STATES.DELETED);
  }, [getFilesByState, fileService.FILE_STATES]);

  // Get pending files
  const getPendingFiles = useCallback(() => {
    return getFilesByState(fileService.FILE_STATES.PENDING);
  }, [getFilesByState, fileService.FILE_STATES]);

  // Get files with tombstones
  const getTombstoneFiles = useCallback(() => {
    return files.filter((file) => file._has_tombstone);
  }, [files]);

  // Get expired tombstone files
  const getExpiredTombstoneFiles = useCallback(() => {
    return files.filter((file) => file._tombstone_expired);
  }, [files]);

  // Get restorable files (deleted but not expired)
  const getRestorableFiles = useCallback(() => {
    return files.filter((file) => fileService.canRestoreFile(file));
  }, [files, fileService]);

  // Get permanently deletable files
  const getPermanentlyDeletableFiles = useCallback(() => {
    return files.filter((file) => fileService.canPermanentlyDeleteFile(file));
  }, [files, fileService]);

  // Get file statistics
  const getFileStats = useCallback(() => {
    const stats = {
      total: files.length,
      active: 0,
      archived: 0,
      deleted: 0,
      pending: 0,
      withTombstones: 0,
      expiredTombstones: 0,
      restorable: 0,
      permanentlyDeletable: 0,
    };

    files.forEach((file) => {
      if (file._is_active) stats.active++;
      if (file._is_archived) stats.archived++;
      if (file._is_deleted) stats.deleted++;
      if (file._is_pending) stats.pending++;
      if (file._has_tombstone) stats.withTombstones++;
      if (file._tombstone_expired) stats.expiredTombstones++;
      if (fileService.canRestoreFile(file)) stats.restorable++;
      if (fileService.canPermanentlyDeleteFile(file))
        stats.permanentlyDeletable++;
    });

    return stats;
  }, [files, fileService]);

  // Check if a file can be downloaded
  const canDownloadFile = useCallback(
    (file) => {
      return !file._is_deleted || fileService.canRestoreFile(file);
    },
    [fileService],
  );

  // Check if a file can be edited
  const canEditFile = useCallback((file) => {
    return file._is_active || file._is_archived;
  }, []);

  // Check if a file can be restored
  const canRestoreFile = useCallback(
    (file) => {
      return fileService.canRestoreFile(file);
    },
    [fileService],
  );

  // Check if a file can be permanently deleted
  const canPermanentlyDeleteFile = useCallback(
    (file) => {
      return fileService.canPermanentlyDeleteFile(file);
    },
    [fileService],
  );

  // Get file version information
  const getFileVersionInfo = useCallback(
    (file) => {
      return {
        currentVersion: file.version || 1,
        hasTombstone: file._has_tombstone,
        tombstoneVersion: file.tombstone_version || 0,
        tombstoneExpiry: file.tombstone_expiry,
        isExpired: file._tombstone_expired,
        canRestore: fileService.canRestoreFile(file),
        canPermanentlyDelete: fileService.canPermanentlyDeleteFile(file),
      };
    },
    [fileService],
  );

  // Sync files for offline support
  const syncFiles = useCallback(
    async (cursor = null, limit = 5000) => {
      try {
        setIsLoading(true);
        setError(null);

        return await fileService.syncFiles(cursor, limit);
      } catch (err) {
        console.error("[useFiles] Failed to sync files:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Sync all files
  const syncAllFiles = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const allSyncedFiles = await fileService.syncAllFiles();

      // You might want to update local state with synced files
      // depending on your use case

      return allSyncedFiles;
    } catch (err) {
      console.error("[useFiles] Failed to sync all files:", err);
      setError(err.message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [fileService]);

  // Get presigned URLs
  const getUploadUrl = useCallback(
    async (fileId, duration) => {
      try {
        return await fileService.getPresignedUploadUrl(fileId, duration);
      } catch (err) {
        console.error("[useFiles] Failed to get upload URL:", err);
        setError(err.message);
        throw err;
      }
    },
    [fileService],
  );

  const getDownloadUrl = useCallback(
    async (fileId, duration) => {
      try {
        return await fileService.getPresignedDownloadUrl(fileId, duration);
      } catch (err) {
        console.error("[useFiles] Failed to get download URL:", err);
        setError(err.message);
        throw err;
      }
    },
    [fileService],
  );

  // Clear cache
  const clearCache = useCallback(() => {
    fileService.clearCache();
    setFiles([]);
    setUploadQueue(new Map());
  }, [fileService]);

  // Reload files
  const reloadFiles = useCallback(
    (forceRefresh = true, includeStates = null) => {
      if (collectionId) {
        return loadFilesByCollection(collectionId, forceRefresh, includeStates);
      }
      return Promise.resolve([]);
    },
    [collectionId, loadFilesByCollection],
  );

  // Download and decrypt a file
  const downloadAndSaveFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        console.log("[useFiles] Starting file download:", fileId);

        // Check if file can be downloaded
        const file = fileService.getCachedFile(fileId);
        if (file && !canDownloadFile(file)) {
          throw new Error("File cannot be downloaded in its current state");
        }

        const result = await fileService.downloadAndSaveFile(fileId);

        console.log(
          "[useFiles] File downloaded successfully:",
          result.filename,
        );
        return result;
      } catch (err) {
        console.error("[useFiles] Failed to download file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService, canDownloadFile],
  );

  // Get file version history
  const getFileVersionHistory = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        return await fileService.getFileVersionHistory(fileId);
      } catch (err) {
        console.error("[useFiles] Failed to get file version history:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Initial load when collection ID changes
  useEffect(() => {
    if (authService.isAuthenticated() && collectionId) {
      loadFilesByCollection(collectionId);
    } else {
      setFiles([]);
    }
  }, [authService.isAuthenticated, collectionId, loadFilesByCollection]);

  return {
    // State
    files,
    isLoading,
    error,
    uploadQueue: Array.from(uploadQueue.values()),

    // File operations
    uploadFile,
    downloadFile,
    updateFile,
    deleteFile,
    deleteMultipleFiles,
    archiveFile,
    restoreFile,
    batchArchiveFiles,
    batchRestoreFiles,

    // Download file
    downloadAndSaveFile,

    // Collection operations
    loadFilesByCollection,
    reloadFiles,

    // Sync operations
    syncFiles,
    syncAllFiles,

    // URL operations
    getUploadUrl,
    getDownloadUrl,

    // File filtering and state queries
    getFilesByState,
    getFilesByStates,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
    getPendingFiles,
    getTombstoneFiles,
    getExpiredTombstoneFiles,
    getRestorableFiles,
    getPermanentlyDeletableFiles,

    // File capability checks
    canDownloadFile,
    canEditFile,
    canRestoreFile,
    canPermanentlyDeleteFile,

    // Statistics and metadata
    getFileStats,
    getFileVersionInfo,
    getFileVersionHistory,

    // Utility functions
    clearCache,

    // File states constants
    FILE_STATES: fileService.FILE_STATES,

    // Upload status
    UPLOAD_STATUS: {
      UPLOADING: "uploading",
      ERROR: "error",
      COMPLETED: "completed",
    },
  };
};

export default useFiles;

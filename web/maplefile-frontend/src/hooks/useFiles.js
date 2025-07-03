// Custom hook for file management
import { useState, useEffect, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useFiles = (collectionId = null) => {
  const { fileService, authService } = useServices();
  const [files, setFiles] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [uploadQueue, setUploadQueue] = useState(new Map());

  // Load files for a specific collection
  const loadFilesByCollection = useCallback(
    async (targetCollectionId, forceRefresh = false) => {
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

  // Update file metadata
  const updateFile = useCallback(
    async (fileId, updateData) => {
      try {
        setIsLoading(true);
        setError(null);

        const updatedFile = await fileService.updateFile(fileId, updateData);

        // Update local state
        setFiles((prev) =>
          prev.map((file) => (file.id === fileId ? updatedFile : file)),
        );

        return updatedFile;
      } catch (err) {
        console.error("[useFiles] Failed to update file:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [fileService],
  );

  // Delete a file
  const deleteFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        await fileService.deleteFile(fileId);

        // Update local state - mark as deleted
        setFiles((prev) =>
          prev.map((file) =>
            file.id === fileId ? { ...file, state: "deleted" } : file,
          ),
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

        // Update local state - mark as deleted
        setFiles((prev) =>
          prev.map((file) =>
            fileIds.includes(file.id) ? { ...file, state: "deleted" } : file,
          ),
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
          prev.map((file) =>
            file.id === fileId ? { ...file, state: "archived" } : file,
          ),
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

  // Restore a file
  const restoreFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        await fileService.restoreFile(fileId);

        // Update local state
        setFiles((prev) =>
          prev.map((file) =>
            file.id === fileId ? { ...file, state: "active" } : file,
          ),
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
          prev.map((file) =>
            successfulIds.includes(file.id)
              ? { ...file, state: "archived" }
              : file,
          ),
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
          prev.map((file) =>
            successfulIds.includes(file.id)
              ? { ...file, state: "active" }
              : file,
          ),
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
    (state = "active") => {
      return files.filter((file) => file.state === state);
    },
    [files],
  );

  // Get active files
  const getActiveFiles = useCallback(() => {
    return getFilesByState("active");
  }, [getFilesByState]);

  // Get archived files
  const getArchivedFiles = useCallback(() => {
    return getFilesByState("archived");
  }, [getFilesByState]);

  // Get deleted files
  const getDeletedFiles = useCallback(() => {
    return getFilesByState("deleted");
  }, [getFilesByState]);

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
    (forceRefresh = true) => {
      if (collectionId) {
        return loadFilesByCollection(collectionId, forceRefresh);
      }
      return Promise.resolve([]);
    },
    [collectionId, loadFilesByCollection],
  );

  // Initial load when collection ID changes
  useEffect(() => {
    if (authService.isAuthenticated() && collectionId) {
      loadFilesByCollection(collectionId);
    } else {
      setFiles([]);
    }
  }, [authService.isAuthenticated, collectionId, loadFilesByCollection]);

  // Download and decrypt a file
  const downloadAndSaveFile = useCallback(
    async (fileId) => {
      try {
        setIsLoading(true);
        setError(null);

        console.log("[useFiles] Starting file download:", fileId);
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
    [fileService],
  );

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

    // Utility functions
    getFilesByState,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
    clearCache,

    // File states
    FILE_STATES: {
      PENDING: "pending",
      ACTIVE: "active",
      DELETED: "deleted",
      ARCHIVED: "archived",
    },

    // Upload status
    UPLOAD_STATUS: {
      UPLOADING: "uploading",
      ERROR: "error",
      COMPLETED: "completed",
    },
  };
};

export default useFiles;

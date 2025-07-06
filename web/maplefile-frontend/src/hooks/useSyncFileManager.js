// File: monorepo/web/maplefile-frontend/src/hooks/useSyncFileManager.js
// Custom React hook for easy sync file management functionality

import { useState, useCallback, useEffect } from "react";
import { useServices } from "./useService.jsx";

const useSyncFileManager = () => {
  const { syncFileManager } = useServices();
  const [syncFiles, setSyncFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Get sync files with smart caching
  const getSyncFiles = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log("[useSyncFileManager] Getting sync files...");

        const files = await syncFileManager.getSyncFiles(options);
        setSyncFiles(files);

        console.log(
          "[useSyncFileManager] ✅ Retrieved:",
          files.length,
          "sync files",
        );

        return files;
      } catch (err) {
        console.error("[useSyncFileManager] ❌ Failed to get sync files:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileManager],
  );

  // Get sync files by collection
  const getSyncFilesByCollection = useCallback(
    async (collectionId, options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log(
          "[useSyncFileManager] Getting sync files for collection:",
          collectionId,
        );

        const files = await syncFileManager.getSyncFilesByCollection(
          collectionId,
          options,
        );
        setSyncFiles(files);

        console.log(
          "[useSyncFileManager] ✅ Retrieved:",
          files.length,
          "sync files for collection",
        );

        return files;
      } catch (err) {
        console.error(
          "[useSyncFileManager] ❌ Failed to get sync files by collection:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileManager],
  );

  // Get single sync file
  const getSyncFile = useCallback(
    async (fileId, options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log("[useSyncFileManager] Getting sync file:", fileId);

        const file = await syncFileManager.getSyncFile(fileId, options);

        console.log("[useSyncFileManager] ✅ Retrieved sync file:", fileId);
        return file;
      } catch (err) {
        console.error("[useSyncFileManager] ❌ Failed to get sync file:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileManager],
  );

  // Refresh sync files from API
  const refreshSyncFiles = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log("[useSyncFileManager] Refreshing sync files from API...");

        const files = await syncFileManager.refreshSyncFiles(options);
        setSyncFiles(files);

        console.log(
          "[useSyncFileManager] ✅ Refreshed:",
          files.length,
          "sync files",
        );

        return files;
      } catch (err) {
        console.error(
          "[useSyncFileManager] ❌ Failed to refresh sync files:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileManager],
  );

  // Force refresh sync files
  const forceRefreshSyncFiles = useCallback(
    async (options = {}) => {
      return await refreshSyncFiles(options);
    },
    [refreshSyncFiles],
  );

  // Clear sync files
  const clearSyncFiles = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncFileManager] Clearing sync files...");

      const success = syncFileManager.clearSyncFiles();
      if (success) {
        setSyncFiles([]);
        console.log("[useSyncFileManager] ✅ Sync files cleared");
        return true;
      } else {
        throw new Error("Failed to clear sync files");
      }
    } catch (err) {
      console.error("[useSyncFileManager] ❌ Failed to clear sync files:", err);
      setError(err.message);
      throw err;
    }
  }, [syncFileManager]);

  // Get active sync files
  const getActiveSyncFiles = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        const activeFiles = await syncFileManager.getActiveSyncFiles(options);
        console.log(
          "[useSyncFileManager] ✅ Retrieved:",
          activeFiles.length,
          "active sync files",
        );
        return activeFiles;
      } catch (err) {
        console.error(
          "[useSyncFileManager] ❌ Failed to get active sync files:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileManager],
  );

  // Get sync files from storage only
  const getSyncFilesFromStorage = useCallback(() => {
    try {
      setError(null);
      const storedFiles = syncFileManager.getSyncFilesFromStorage();
      setSyncFiles(storedFiles);
      console.log(
        "[useSyncFileManager] ✅ Loaded from storage:",
        storedFiles.length,
        "sync files",
      );
      return storedFiles;
    } catch (err) {
      console.error(
        "[useSyncFileManager] ❌ Failed to load from storage:",
        err,
      );
      setError(err.message);
      throw err;
    }
  }, [syncFileManager]);

  // Load initial data on mount
  useEffect(() => {
    getSyncFilesFromStorage();
  }, [getSyncFilesFromStorage]);

  // Get filtered sync files by state
  const filterByState = useCallback(
    (state) => {
      return syncFileManager.filterSyncFilesByState(syncFiles, state);
    },
    [syncFileManager, syncFiles],
  );

  // Get filtered sync files by collection
  const filterByCollection = useCallback(
    (collectionId) => {
      return syncFileManager.filterSyncFilesByCollection(
        syncFiles,
        collectionId,
      );
    },
    [syncFileManager, syncFiles],
  );

  // Get sync file statistics
  const getStats = useCallback(() => {
    return syncFileManager.getSyncFileStats(syncFiles);
  }, [syncFileManager, syncFiles]);

  // Get file size statistics
  const getSizeStats = useCallback(() => {
    return syncFileManager.getFileSizeStats(syncFiles);
  }, [syncFileManager, syncFiles]);

  return {
    // State
    syncFiles,
    loading,
    error,

    // Actions
    getSyncFiles,
    getSyncFilesByCollection,
    getSyncFile,
    refreshSyncFiles,
    forceRefreshSyncFiles,
    clearSyncFiles,
    getActiveSyncFiles,
    getSyncFilesFromStorage,

    // Utilities
    filterByState,
    filterByCollection,
    getStats,
    getSizeStats,

    // Computed values
    syncFilesCount: syncFiles.length,
    activeSyncFiles: filterByState("active"),
    deletedSyncFiles: filterByState("deleted"),
    archivedSyncFiles: filterByState("archived"),
    statistics: getStats(),
    sizeStatistics: getSizeStats(),

    // Manager state
    isManagerLoading: syncFileManager.getIsLoading(),
    isAPILoading: syncFileManager.isAPILoading(),
    storageInfo: syncFileManager.getStorageInfo(),
    managerStatus: syncFileManager.getManagerStatus(),

    // Debug info
    debugInfo: syncFileManager.getDebugInfo(),
  };
};

export default useSyncFileManager;

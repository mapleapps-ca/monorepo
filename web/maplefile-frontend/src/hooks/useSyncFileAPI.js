// File: monorepo/web/maplefile-frontend/src/hooks/useSyncFileAPI.js
// Custom React hook for easy sync file API functionality

import { useState, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useSyncFileAPI = () => {
  const { syncFileAPIService } = useServices();
  const [syncFiles, setSyncFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // The ONLY function you need - sync all files from API
  const syncAllFiles = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);
      setSyncFiles([]);

      try {
        console.log("[useSyncFileAPI] Starting complete API sync...");

        const allSyncFiles = await syncFileAPIService.syncAllFiles(options);

        setSyncFiles(allSyncFiles);
        console.log(
          "[useSyncFileAPI] ✅ API sync complete:",
          allSyncFiles.length,
          "sync files",
        );

        return allSyncFiles;
      } catch (err) {
        console.error("[useSyncFileAPI] ❌ API sync failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileAPIService],
  );

  // Sync files for a specific collection
  const syncFilesByCollection = useCallback(
    async (collectionId, options = {}) => {
      setLoading(true);
      setError(null);
      setSyncFiles([]);

      try {
        console.log(
          "[useSyncFileAPI] Syncing files for collection:",
          collectionId,
        );

        const collectionFiles = await syncFileAPIService.syncFilesByCollection(
          collectionId,
          options,
        );

        setSyncFiles(collectionFiles);
        console.log(
          "[useSyncFileAPI] ✅ Collection sync complete:",
          collectionFiles.length,
          "sync files",
        );

        return collectionFiles;
      } catch (err) {
        console.error("[useSyncFileAPI] ❌ Collection sync failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileAPIService],
  );

  // Get a single sync file
  const getSyncFile = useCallback(
    async (fileId) => {
      setLoading(true);
      setError(null);

      try {
        console.log("[useSyncFileAPI] Getting sync file:", fileId);

        const syncFile = await syncFileAPIService.getSyncFile(fileId);

        console.log("[useSyncFileAPI] ✅ Got sync file:", fileId);
        return syncFile;
      } catch (err) {
        console.error("[useSyncFileAPI] ❌ Failed to get sync file:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncFileAPIService],
  );

  return {
    // State
    syncFiles,
    loading,
    error,

    // Actions
    syncAllFiles,
    syncFilesByCollection,
    getSyncFile,

    // Computed values
    syncFilesCount: syncFiles.length,
    activeSyncFiles: syncFiles.filter((f) => f.state === "active"),
    deletedSyncFiles: syncFiles.filter((f) => f.state === "deleted"),
    archivedSyncFiles: syncFiles.filter((f) => f.state === "archived"),
    totalFileSize: syncFiles.reduce((sum, f) => sum + (f.file_size || 0), 0),

    // Service state
    isServiceLoading: syncFileAPIService.isLoadingSync(),

    // Debug info
    debugInfo: {
      syncFiles: syncFiles.length,
      loading,
      error,
      serviceDebug: syncFileAPIService.getDebugInfo(),
    },
  };
};

export default useSyncFileAPI;

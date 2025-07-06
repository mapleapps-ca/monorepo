// File: monorepo/web/maplefile-frontend/src/hooks/useSyncFileStorage.js
// Custom React hook for easy sync file storage functionality

import { useState, useCallback, useEffect } from "react";
import { useServices } from "./useService.jsx";

const useSyncFileStorage = () => {
  const { syncFileStorageService } = useServices();
  const [syncFiles, setSyncFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [storageInfo, setStorageInfo] = useState({});

  // Update storage info
  const updateStorageInfo = useCallback(() => {
    const info = syncFileStorageService.getStorageInfo();
    setStorageInfo(info);
    return info;
  }, [syncFileStorageService]);

  // Load sync files from localStorage
  const loadFromStorage = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncFileStorage] Loading from localStorage...");

      const stored = syncFileStorageService.getSyncFiles();
      setSyncFiles(stored);
      updateStorageInfo();

      console.log(
        "[useSyncFileStorage] ✅ Loaded:",
        stored.length,
        "sync files",
      );
      return stored;
    } catch (err) {
      setError(err.message);
      console.error("[useSyncFileStorage] ❌ Load failed:", err);
      throw err;
    }
  }, [syncFileStorageService, updateStorageInfo]);

  // Save sync files to localStorage
  const saveToStorage = useCallback(
    (files = syncFiles) => {
      try {
        setError(null);
        console.log("[useSyncFileStorage] Saving to localStorage...");

        const success = syncFileStorageService.saveSyncFiles(files);
        if (success) {
          setSyncFiles(files);
          updateStorageInfo();
          console.log(
            "[useSyncFileStorage] ✅ Saved:",
            files.length,
            "sync files",
          );
          return true;
        } else {
          throw new Error("Failed to save sync files");
        }
      } catch (err) {
        setError(err.message);
        console.error("[useSyncFileStorage] ❌ Save failed:", err);
        throw err;
      }
    },
    [syncFiles, syncFileStorageService, updateStorageInfo],
  );

  // Get sync files by collection
  const getFilesByCollection = useCallback(
    (collectionId) => {
      try {
        setError(null);
        console.log(
          "[useSyncFileStorage] Getting files for collection:",
          collectionId,
        );

        const files =
          syncFileStorageService.getSyncFilesByCollection(collectionId);
        console.log(
          "[useSyncFileStorage] ✅ Found:",
          files.length,
          "files in collection",
        );
        return files;
      } catch (err) {
        setError(err.message);
        console.error("[useSyncFileStorage] ❌ Get by collection failed:", err);
        throw err;
      }
    },
    [syncFileStorageService],
  );

  // Get single sync file
  const getFileById = useCallback(
    (fileId) => {
      try {
        setError(null);
        console.log("[useSyncFileStorage] Getting file:", fileId);

        const file = syncFileStorageService.getSyncFileById(fileId);
        if (file) {
          console.log("[useSyncFileStorage] ✅ Found file:", fileId);
        } else {
          console.log("[useSyncFileStorage] ❌ File not found:", fileId);
        }
        return file;
      } catch (err) {
        setError(err.message);
        console.error("[useSyncFileStorage] ❌ Get by ID failed:", err);
        throw err;
      }
    },
    [syncFileStorageService],
  );

  // Clear all stored sync files
  const clearStorage = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncFileStorage] Clearing storage...");

      const success = syncFileStorageService.clearSyncFiles();
      if (success) {
        setSyncFiles([]);
        updateStorageInfo();
        console.log("[useSyncFileStorage] ✅ Storage cleared");
        return true;
      } else {
        throw new Error("Failed to clear storage");
      }
    } catch (err) {
      setError(err.message);
      console.error("[useSyncFileStorage] ❌ Clear failed:", err);
      throw err;
    }
  }, [syncFileStorageService, updateStorageInfo]);

  // Load storage info on mount
  useEffect(() => {
    updateStorageInfo();
  }, [updateStorageInfo]);

  return {
    // State
    syncFiles,
    loading,
    error,
    storageInfo,

    // Actions
    loadFromStorage,
    saveToStorage,
    clearStorage,
    updateStorageInfo,
    getFilesByCollection,
    getFileById,

    // Computed values
    syncFilesCount: syncFiles.length,
    hasStoredSyncFiles: storageInfo.hasSyncFiles,
    lastSaved: storageInfo.metadata?.savedAt,
    collectionBreakdown: storageInfo.collectionBreakdown,

    // Helpers
    activeSyncFiles: syncFiles.filter((f) => f.state === "active"),
    deletedSyncFiles: syncFiles.filter((f) => f.state === "deleted"),
    archivedSyncFiles: syncFiles.filter((f) => f.state === "archived"),
    totalFileSize: syncFiles.reduce((sum, f) => sum + (f.file_size || 0), 0),
  };
};

export default useSyncFileStorage;

// useSyncCollectionStorage.js
// Custom React hook for easy sync collection storage functionality

import { useState, useCallback, useEffect } from "react";
import { useServices } from "./useService.jsx";

const useSyncCollectionStorage = () => {
  const { syncCollectionStorageService } = useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [storageInfo, setStorageInfo] = useState({});

  // Update storage info
  const updateStorageInfo = useCallback(() => {
    const info = syncCollectionStorageService.getStorageInfo();
    setStorageInfo(info);
    return info;
  }, [syncCollectionStorageService]);

  // Load sync collections from localStorage
  const loadFromStorage = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncCollectionStorage] Loading from localStorage...");

      const stored = syncCollectionStorageService.getSyncCollections();
      setSyncCollections(stored);
      updateStorageInfo();

      console.log(
        "[useSyncCollectionStorage] ✅ Loaded:",
        stored.length,
        "sync collections",
      );
      return stored;
    } catch (err) {
      console.error("[useSyncCollectionStorage] ❌ Load failed:", err);
      setError(err.message);
      throw err;
    }
  }, [syncCollectionStorageService, updateStorageInfo]);

  // Save sync collections to localStorage
  const saveToStorage = useCallback(
    (collections = syncCollections) => {
      try {
        setError(null);
        console.log("[useSyncCollectionStorage] Saving to localStorage...");

        const success =
          syncCollectionStorageService.saveSyncCollections(collections);
        if (success) {
          setSyncCollections(collections);
          updateStorageInfo();
          console.log(
            "[useSyncCollectionStorage] ✅ Saved:",
            collections.length,
            "sync collections",
          );
          return true;
        } else {
          throw new Error("Failed to save sync collections");
        }
      } catch (err) {
        console.error("[useSyncCollectionStorage] ❌ Save failed:", err);
        setError(err.message);
        throw err;
      }
    },
    [syncCollections, syncCollectionStorageService, updateStorageInfo],
  );

  // Clear all stored sync collections
  const clearStorage = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncCollectionStorage] Clearing storage...");

      const success = syncCollectionStorageService.clearSyncCollections();
      if (success) {
        setSyncCollections([]);
        updateStorageInfo();
        console.log("[useSyncCollectionStorage] ✅ Storage cleared");
        return true;
      } else {
        throw new Error("Failed to clear storage");
      }
    } catch (err) {
      console.error("[useSyncCollectionStorage] ❌ Clear failed:", err);
      setError(err.message);
      throw err;
    }
  }, [syncCollectionStorageService, updateStorageInfo]);

  // Load storage info on mount
  useEffect(() => {
    updateStorageInfo();
  }, [updateStorageInfo]);

  return {
    // State
    syncCollections,
    loading,
    error,
    storageInfo,

    // Actions
    loadFromStorage,
    saveToStorage,
    clearStorage,
    updateStorageInfo,

    // Computed values
    syncCollectionsCount: syncCollections.length,
    hasStoredSyncCollections: storageInfo.hasSyncCollections,
    lastSaved: storageInfo.metadata?.savedAt,

    // Helpers
    activeSyncCollections: syncCollections.filter((c) => c.state === "active"),
    deletedSyncCollections: syncCollections.filter(
      (c) => c.state === "deleted",
    ),
    archivedSyncCollections: syncCollections.filter(
      (c) => c.state === "archived",
    ),
  };
};

export default useSyncCollectionStorage;

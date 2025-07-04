// File: monorepo/web/maplefile-frontend/src/hooks/useSyncCollectionManager.js
// Custom React hook for easy sync collection management functionality

import { useState, useCallback, useEffect } from "react";
import { useServices } from "./useService.jsx";

const useSyncCollectionManager = () => {
  const { syncCollectionManager } = useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Get sync collections with smart caching
  const getSyncCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log("[useSyncCollectionManager] Getting sync collections...");

        const collections =
          await syncCollectionManager.getSyncCollections(options);
        setSyncCollections(collections);

        console.log(
          "[useSyncCollectionManager] ✅ Retrieved:",
          collections.length,
          "sync collections",
        );

        return collections;
      } catch (err) {
        console.error(
          "[useSyncCollectionManager] ❌ Failed to get sync collections:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionManager],
  );

  // Refresh sync collections from API
  const refreshSyncCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log(
          "[useSyncCollectionManager] Refreshing sync collections from API...",
        );

        const collections =
          await syncCollectionManager.refreshSyncCollections(options);
        setSyncCollections(collections);

        console.log(
          "[useSyncCollectionManager] ✅ Refreshed:",
          collections.length,
          "sync collections",
        );

        return collections;
      } catch (err) {
        console.error(
          "[useSyncCollectionManager] ❌ Failed to refresh sync collections:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionManager],
  );

  // Force refresh sync collections
  const forceRefreshSyncCollections = useCallback(
    async (options = {}) => {
      return await refreshSyncCollections(options);
    },
    [refreshSyncCollections],
  );

  // Clear sync collections
  const clearSyncCollections = useCallback(() => {
    try {
      setError(null);
      console.log("[useSyncCollectionManager] Clearing sync collections...");

      const success = syncCollectionManager.clearSyncCollections();
      if (success) {
        setSyncCollections([]);
        console.log("[useSyncCollectionManager] ✅ Sync collections cleared");
        return true;
      } else {
        throw new Error("Failed to clear sync collections");
      }
    } catch (err) {
      console.error(
        "[useSyncCollectionManager] ❌ Failed to clear sync collections:",
        err,
      );
      setError(err.message);
      throw err;
    }
  }, [syncCollectionManager]);

  // Get active sync collections
  const getActiveSyncCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        const activeCollections =
          await syncCollectionManager.getActiveSyncCollections(options);
        console.log(
          "[useSyncCollectionManager] ✅ Retrieved:",
          activeCollections.length,
          "active sync collections",
        );
        return activeCollections;
      } catch (err) {
        console.error(
          "[useSyncCollectionManager] ❌ Failed to get active sync collections:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionManager],
  );

  // Get sync collections from storage only
  const getSyncCollectionsFromStorage = useCallback(() => {
    try {
      setError(null);
      const storedCollections =
        syncCollectionManager.getSyncCollectionsFromStorage();
      setSyncCollections(storedCollections);
      console.log(
        "[useSyncCollectionManager] ✅ Loaded from storage:",
        storedCollections.length,
        "sync collections",
      );
      return storedCollections;
    } catch (err) {
      console.error(
        "[useSyncCollectionManager] ❌ Failed to load from storage:",
        err,
      );
      setError(err.message);
      throw err;
    }
  }, [syncCollectionManager]);

  // Load initial data on mount
  useEffect(() => {
    getSyncCollectionsFromStorage();
  }, [getSyncCollectionsFromStorage]);

  // Get filtered sync collections by state
  const filterByState = useCallback(
    (state) => {
      return syncCollectionManager.filterSyncCollectionsByState(
        syncCollections,
        state,
      );
    },
    [syncCollectionManager, syncCollections],
  );

  // Get sync collection statistics
  const getStats = useCallback(() => {
    return syncCollectionManager.getSyncCollectionStats(syncCollections);
  }, [syncCollectionManager, syncCollections]);

  return {
    // State
    syncCollections,
    loading,
    error,

    // Actions
    getSyncCollections,
    refreshSyncCollections,
    forceRefreshSyncCollections,
    clearSyncCollections,
    getActiveSyncCollections,
    getSyncCollectionsFromStorage,

    // Utilities
    filterByState,
    getStats,

    // Computed values
    syncCollectionsCount: syncCollections.length,
    activeSyncCollections: filterByState("active"),
    deletedSyncCollections: filterByState("deleted"),
    archivedSyncCollections: filterByState("archived"),
    statistics: getStats(),

    // Manager state
    isManagerLoading: syncCollectionManager.getIsLoading(),
    isAPILoading: syncCollectionManager.isAPILoading(),
    storageInfo: syncCollectionManager.getStorageInfo(),
    managerStatus: syncCollectionManager.getManagerStatus(),

    // Debug info
    debugInfo: syncCollectionManager.getDebugInfo(),
  };
};

export default useSyncCollectionManager;

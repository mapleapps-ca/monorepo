// useSyncCollections.js - UPDATED
// Custom React hook for easy sync collection API functionality

import { useState, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useSyncCollections = () => {
  const { syncCollectionAPIService } = useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // The ONLY function you need - sync all collections from API
  const syncAllCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);
      setSyncCollections([]);

      try {
        console.log("[useSyncCollections] Starting complete API sync...");

        const allSyncCollections =
          await syncCollectionAPIService.syncAllCollections(options);

        setSyncCollections(allSyncCollections);
        console.log(
          "[useSyncCollections] ✅ API sync complete:",
          allSyncCollections.length,
          "sync collections",
        );

        return allSyncCollections;
      } catch (err) {
        console.error("[useSyncCollections] ❌ API sync failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionAPIService],
  );

  return {
    // State
    syncCollections,
    loading,
    error,

    // The ONE function you need
    syncAllCollections,

    // Computed values
    syncCollectionsCount: syncCollections.length,
    activeSyncCollections: syncCollections.filter((c) => c.state === "active"),
    deletedSyncCollections: syncCollections.filter(
      (c) => c.state === "deleted",
    ),
    archivedSyncCollections: syncCollections.filter(
      (c) => c.state === "archived",
    ),

    // Service state
    isServiceLoading: syncCollectionAPIService.isLoadingSync(),

    // Debug info
    debugInfo: {
      syncCollections: syncCollections.length,
      loading,
      error,
      serviceDebug: syncCollectionAPIService.getDebugInfo(),
    },
  };
};

export default useSyncCollections;

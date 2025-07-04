// useSyncCollections.js - SIMPLIFIED HOOK
// Just one function: syncAllCollections()

import { useState, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useSyncCollections = () => {
  const { syncCollectionsService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // The ONLY function you need - sync all collections
  const syncAllCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);
      setCollections([]);

      try {
        console.log("[useSyncCollections] Starting complete sync...");

        const allCollections =
          await syncCollectionsService.syncAllCollections(options);

        setCollections(allCollections);
        console.log(
          "[useSyncCollections] ✅ Sync complete:",
          allCollections.length,
          "collections",
        );

        return allCollections;
      } catch (err) {
        console.error("[useSyncCollections] ❌ Sync failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionsService],
  );

  return {
    // State
    collections,
    loading,
    error,

    // The ONE function you need
    syncAllCollections,

    // Computed values
    collectionsCount: collections.length,
    activeCollections: collections.filter((c) => c.state === "active"),
    deletedCollections: collections.filter((c) => c.state === "deleted"),
    archivedCollections: collections.filter((c) => c.state === "archived"),
  };
};

export default useSyncCollections;

// useSyncCollections.js
// Custom React hook for easy sync collections functionality

import { useState, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useSyncCollections = () => {
  const { syncCollectionsService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [nextCursor, setNextCursor] = useState(null);
  const [hasMore, setHasMore] = useState(false);
  const [pageCount, setPageCount] = useState(0);

  // Clear all state
  const clearState = useCallback(() => {
    setCollections([]);
    setError(null);
    setNextCursor(null);
    setHasMore(false);
    setPageCount(0);
  }, []);

  // Sync collections with options
  const syncCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);

      try {
        console.log(
          "[useSyncCollections] Syncing collections with options:",
          options,
        );

        const response = await syncCollectionsService.syncCollections(options);

        // Replace collections if this is a fresh sync (no cursor in options)
        if (!options.cursor) {
          setCollections(response.collections || []);
          setPageCount(1);
        } else {
          // Append to existing collections if using cursor (pagination)
          setCollections((prev) => [...prev, ...(response.collections || [])]);
          setPageCount((prev) => prev + 1);
        }

        setNextCursor(response.next_cursor);
        setHasMore(response.has_more);

        return response;
      } catch (err) {
        console.error("[useSyncCollections] Sync failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionsService],
  );

  // Load next page
  const loadNextPage = useCallback(
    async (limit = 1000) => {
      if (!nextCursor) {
        throw new Error("No next cursor available");
      }

      return await syncCollections({ limit, cursor: nextCursor });
    },
    [nextCursor, syncCollections],
  );

  // Sync all collections with automatic pagination
  const syncAllCollections = useCallback(
    async (options = {}) => {
      setLoading(true);
      setError(null);
      clearState();

      try {
        console.log("[useSyncCollections] Syncing all collections");

        const allCollections = await syncCollectionsService.syncAllCollections({
          ...options,
          onPageReceived: (pageCollections, pageNum, response) => {
            console.log(
              `[useSyncCollections] Received page ${pageNum} with ${pageCollections.length} collections`,
            );
            setPageCount(pageNum);
            setCollections((prev) => [...prev, ...pageCollections]);

            // Call user's callback if provided
            if (options.onPageReceived) {
              options.onPageReceived(pageCollections, pageNum, response);
            }
          },
        });

        setHasMore(false);
        setNextCursor(null);

        return allCollections;
      } catch (err) {
        console.error("[useSyncCollections] Sync all failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionsService, clearState],
  );

  // Sync collections since a specific time
  const syncCollectionsSince = useCallback(
    async (sinceModified, options = {}) => {
      setLoading(true);
      setError(null);
      clearState();

      try {
        console.log(
          "[useSyncCollections] Syncing collections since:",
          sinceModified,
        );

        const recentCollections =
          await syncCollectionsService.syncCollectionsSince(
            sinceModified,
            options,
          );

        setCollections(recentCollections);
        setHasMore(false);
        setNextCursor(null);
        setPageCount(1);

        return recentCollections;
      } catch (err) {
        console.error("[useSyncCollections] Sync since failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [syncCollectionsService, clearState],
  );

  // Refresh current page
  const refresh = useCallback(
    async (limit = 1000) => {
      return await syncCollections({ limit });
    },
    [syncCollections],
  );

  // Get collections by state
  const getCollectionsByState = useCallback(
    (state) => {
      return collections.filter((collection) => collection.state === state);
    },
    [collections],
  );

  // Get collections by parent ID
  const getCollectionsByParent = useCallback(
    (parentId) => {
      return collections.filter(
        (collection) => collection.parent_id === parentId,
      );
    },
    [collections],
  );

  // Get root collections (no parent)
  const getRootCollections = useCallback(() => {
    return collections.filter((collection) => !collection.parent_id);
  }, [collections]);

  // Get tombstone collections
  const getTombstoneCollections = useCallback(() => {
    return collections.filter((collection) => collection.tombstone_version > 0);
  }, [collections]);

  return {
    // State
    collections,
    loading,
    error,
    nextCursor,
    hasMore,
    pageCount,

    // Actions
    syncCollections,
    loadNextPage,
    syncAllCollections,
    syncCollectionsSince,
    refresh,
    clearState,

    // Helpers
    getCollectionsByState,
    getCollectionsByParent,
    getRootCollections,
    getTombstoneCollections,

    // Computed values
    collectionsCount: collections.length,
    activeCollections: collections.filter((c) => c.state === "active"),
    deletedCollections: collections.filter((c) => c.state === "deleted"),
    archivedCollections: collections.filter((c) => c.state === "archived"),

    // Service state
    isServiceLoading: syncCollectionsService.isLoadingSync(),

    // Debug info
    debugInfo: {
      collections: collections.length,
      loading,
      error,
      nextCursor: nextCursor ? `${nextCursor.substring(0, 20)}...` : null,
      hasMore,
      pageCount,
      serviceDebug: syncCollectionsService.getDebugInfo(),
    },
  };
};

export default useSyncCollections;

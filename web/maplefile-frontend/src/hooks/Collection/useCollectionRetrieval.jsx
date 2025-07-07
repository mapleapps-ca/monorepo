// File: monorepo/web/maplefile-frontend/src/hooks/Collection/useCollectionRetrieval.jsx
// Custom hook for collection retrieval with convenient API
import { useState, useEffect, useCallback } from "react";
import { useCollections, useAuth } from "../useService.jsx";

/**
 * Hook for collection retrieval with state management and convenience methods
 * @returns {Object} Collection retrieval API
 */
const useCollectionRetrieval = () => {
  const { getCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [retrievedCollections, setRetrievedCollections] = useState([]);
  const [cachedCollections, setCachedCollections] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Load cached collections
  const loadCachedCollections = useCallback(() => {
    if (!getCollectionManager) return;

    try {
      const cached = getCollectionManager.getCachedCollections();
      setCachedCollections(cached);
      console.log(
        "[useCollectionRetrieval] Cached collections loaded:",
        cached.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionRetrieval] Failed to load cached collections:",
        err,
      );
      setError(`Failed to load cached collections: ${err.message}`);
    }
  }, [getCollectionManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!getCollectionManager) return;

    try {
      const status = getCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionRetrieval] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionRetrieval] Failed to load manager status:",
        err,
      );
    }
  }, [getCollectionManager]);

  // Get collection with enhanced error handling (uses PasswordStorageService automatically)
  const getCollection = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!getCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionRetrieval] Getting collection:",
          collectionId,
        );

        const result = await getCollectionManager.getCollection(
          collectionId,
          forceRefresh,
        );

        setSuccess(`Collection retrieved successfully from ${result.source}!`);
        loadCachedCollections(); // Reload cached collections
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionRetrieval] Collection retrieved successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionRetrieval] Collection retrieval failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager, loadCachedCollections, loadManagerStatus],
  );

  // Get collection from cache only (uses PasswordStorageService automatically)
  const getCachedCollection = useCallback(
    async (collectionId) => {
      if (!getCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionRetrieval] Getting cached collection:",
          collectionId,
        );

        const result =
          await getCollectionManager.getCachedCollection(collectionId);

        console.log(
          "[useCollectionRetrieval] Cached collection retrieved:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionRetrieval] Cached collection retrieval failed:",
          err,
        );
        setError(`Failed to get cached collection: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager],
  );

  // Refresh collection from API (uses PasswordStorageService automatically)
  const refreshCollection = useCallback(
    async (collectionId) => {
      console.log(
        "[useCollectionRetrieval] Refreshing collection:",
        collectionId,
      );
      return getCollection(collectionId, true);
    },
    [getCollection],
  );

  // Get multiple collections (uses PasswordStorageService automatically)
  const getCollections = useCallback(
    async (collectionIds, forceRefresh = false) => {
      if (!getCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionRetrieval] Getting multiple collections:",
          collectionIds.length,
        );

        const result = await getCollectionManager.getCollections(
          collectionIds,
          forceRefresh,
        );

        setSuccess(
          `Retrieved ${result.successCount}/${collectionIds.length} collections successfully!`,
        );
        loadCachedCollections(); // Reload cached collections
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionRetrieval] Multiple collections retrieved:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionRetrieval] Multiple collections retrieval failed:",
          err,
        );
        setError(`Failed to get collections: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager, loadCachedCollections, loadManagerStatus],
  );

  // Check if collection exists
  const collectionExists = useCallback(
    async (collectionId) => {
      if (!getCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionRetrieval] Checking collection existence:",
          collectionId,
        );

        const exists =
          await getCollectionManager.collectionExists(collectionId);

        console.log(
          "[useCollectionRetrieval] Collection existence check:",
          exists,
        );
        return exists;
      } catch (err) {
        console.error(
          "[useCollectionRetrieval] Collection existence check failed:",
          err,
        );
        setError(`Failed to check collection existence: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager],
  );

  // Get collection cache status
  const getCollectionCacheStatus = useCallback(
    (collectionId) => {
      if (!getCollectionManager) return null;

      return getCollectionManager.getCollectionCacheStatus(collectionId);
    },
    [getCollectionManager],
  );

  // Remove collection from cache
  const removeFromCache = useCallback(
    async (collectionId) => {
      if (!getCollectionManager) {
        throw new Error("Collection manager not available");
      }

      try {
        console.log(
          "[useCollectionRetrieval] Removing collection from cache:",
          collectionId,
        );

        await getCollectionManager.removeFromCache(collectionId);
        setSuccess("Collection removed from cache successfully!");
        loadCachedCollections(); // Reload cached collections
      } catch (err) {
        console.error(
          "[useCollectionRetrieval] Failed to remove collection from cache:",
          err,
        );
        setError(`Failed to remove collection from cache: ${err.message}`);
        throw err;
      }
    },
    [getCollectionManager, loadCachedCollections],
  );

  // Clear all cache
  const clearAllCache = useCallback(async () => {
    if (!getCollectionManager) {
      throw new Error("Collection manager not available");
    }

    try {
      console.log("[useCollectionRetrieval] Clearing all cache");

      await getCollectionManager.clearAllCache();
      setSuccess("All cached collections cleared successfully!");
      loadCachedCollections(); // Reload cached collections
      loadManagerStatus(); // Reload status
    } catch (err) {
      console.error("[useCollectionRetrieval] Failed to clear all cache:", err);
      setError(`Failed to clear all cache: ${err.message}`);
      throw err;
    }
  }, [getCollectionManager, loadCachedCollections, loadManagerStatus]);

  // Clear expired collections
  const clearExpiredCollections = useCallback(async () => {
    if (!getCollectionManager) {
      throw new Error("Collection manager not available");
    }

    try {
      console.log("[useCollectionRetrieval] Clearing expired collections");

      const expiredCount = await getCollectionManager.clearExpiredCollections();
      setSuccess(`Cleared ${expiredCount} expired collections from cache!`);
      loadCachedCollections(); // Reload cached collections
      loadManagerStatus(); // Reload status

      return expiredCount;
    } catch (err) {
      console.error(
        "[useCollectionRetrieval] Failed to clear expired collections:",
        err,
      );
      setError(`Failed to clear expired collections: ${err.message}`);
      throw err;
    }
  }, [getCollectionManager, loadCachedCollections, loadManagerStatus]);

  // Search cached collections
  const searchCachedCollections = useCallback(
    (searchTerm) => {
      if (!getCollectionManager) return [];

      try {
        return getCollectionManager.searchCachedCollections(searchTerm);
      } catch (err) {
        console.error("[useCollectionRetrieval] Search failed:", err);
        return [];
      }
    },
    [getCollectionManager],
  );

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!getCollectionManager) return null;

    try {
      return await getCollectionManager.getUserPassword();
    } catch (err) {
      console.error(
        "[useCollectionRetrieval] Failed to get user password:",
        err,
      );
      return null;
    }
  }, [getCollectionManager]);

  // Clear success/error messages
  const clearMessages = useCallback(() => {
    setError(null);
    setSuccess(null);
  }, []);

  // Reset all state
  const reset = useCallback(() => {
    setError(null);
    setSuccess(null);
    setIsLoading(false);
    setRetrievedCollections([]);
    loadCachedCollections();
    loadManagerStatus();
  }, [loadCachedCollections, loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (getCollectionManager) {
      loadCachedCollections();
      loadManagerStatus();
    }
  }, [getCollectionManager, loadCachedCollections, loadManagerStatus]);

  // Set up event listeners for collection events
  useEffect(() => {
    if (!getCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionRetrieval] Collection event:",
        eventType,
        eventData,
      );

      // Reload collections and status on certain events
      if (
        [
          "collection_retrieved_from_cache",
          "collection_retrieved_from_api",
          "collection_removed_from_cache",
          "all_cache_cleared",
          "multiple_collections_retrieved",
        ].includes(eventType)
      ) {
        loadCachedCollections();
        loadManagerStatus();
      }
    };

    getCollectionManager.addCollectionRetrievalListener(handleCollectionEvent);

    return () => {
      getCollectionManager.removeCollectionRetrievalListener(
        handleCollectionEvent,
      );
    };
  }, [getCollectionManager, loadCachedCollections, loadManagerStatus]);

  return {
    // State
    isLoading,
    error,
    success,
    retrievedCollections,
    cachedCollections,
    managerStatus,

    // Core operations
    getCollection,
    getCachedCollection,
    refreshCollection,
    getCollections,
    collectionExists,

    // Cache operations
    getCollectionCacheStatus,
    removeFromCache,
    clearAllCache,
    clearExpiredCollections,

    // Utility operations
    searchCachedCollections,
    getUserPassword,
    loadCachedCollections,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canGetCollections: managerStatus.canGetCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Cache statistics
    totalCachedCollections: cachedCollections.length,
    cacheStats: managerStatus.cache || {},

    // Helper methods
    isCached: (collectionId) => {
      const status = getCollectionCacheStatus(collectionId);
      return status?.cached && !status?.expired;
    },

    isExpired: (collectionId) => {
      const status = getCollectionCacheStatus(collectionId);
      return status?.cached && status?.expired;
    },
  };
};

export default useCollectionRetrieval;

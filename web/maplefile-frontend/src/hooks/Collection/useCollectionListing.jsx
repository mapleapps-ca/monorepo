// File: monorepo/web/maplefile-frontend/src/hooks/Collection/useCollectionListing.jsx
// Custom hook for collection listing with convenient API

import { useState, useEffect, useCallback } from "react";
import { useCollections } from "../useService.jsx";
import useAuth from "../useAuth.js";

/**
 * Hook for collection listing with state management and convenience methods
 * @returns {Object} Collection listing API
 */
const useCollectionListing = () => {
  const { listCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [collections, setCollections] = useState([]);
  const [filteredCollections, setFilteredCollections] = useState({
    owned_collections: [],
    shared_collections: [],
    total_count: 0,
  });
  const [rootCollections, setRootCollections] = useState([]);
  const [collectionsByParent, setCollectionsByParent] = useState({});
  const [managerStatus, setManagerStatus] = useState({});

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!listCollectionManager) return;

    try {
      const status = listCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionListing] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionListing] Failed to load manager status:",
        err,
      );
    }
  }, [listCollectionManager]);

  // List user collections with enhanced error handling
  const listCollections = useCallback(
    async (forceRefresh = false) => {
      if (!listCollectionManager) {
        throw new Error("List collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log("[useCollectionListing] Listing collections:", {
          forceRefresh,
        });

        const result =
          await listCollectionManager.listCollections(forceRefresh);

        setCollections(result.collections);
        setSuccess(
          `Listed ${result.totalCount} collections successfully from ${result.source}!`,
        );
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionListing] Collections listed successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error("[useCollectionListing] Collection listing failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [listCollectionManager, loadManagerStatus],
  );

  // List filtered collections (owned/shared)
  const listFilteredCollections = useCallback(
    async (
      includeOwned = true,
      includeShared = false,
      forceRefresh = false,
    ) => {
      if (!listCollectionManager) {
        throw new Error("List collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log("[useCollectionListing] Listing filtered collections:", {
          includeOwned,
          includeShared,
          forceRefresh,
        });

        const result = await listCollectionManager.listFilteredCollections(
          includeOwned,
          includeShared,
          forceRefresh,
        );

        setFilteredCollections(result);
        setSuccess(
          `Listed ${result.total_count} filtered collections successfully from ${result.source}!`,
        );
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionListing] Filtered collections listed successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionListing] Filtered collection listing failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [listCollectionManager, loadManagerStatus],
  );

  // List root collections
  const listRootCollections = useCallback(
    async (forceRefresh = false) => {
      if (!listCollectionManager) {
        throw new Error("List collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log("[useCollectionListing] Listing root collections:", {
          forceRefresh,
        });

        const result =
          await listCollectionManager.listRootCollections(forceRefresh);

        setRootCollections(result.collections);
        setSuccess(
          `Listed ${result.totalCount} root collections successfully from ${result.source}!`,
        );
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionListing] Root collections listed successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionListing] Root collection listing failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [listCollectionManager, loadManagerStatus],
  );

  // List collections by parent
  const listCollectionsByParent = useCallback(
    async (parentId, forceRefresh = false) => {
      if (!listCollectionManager) {
        throw new Error("List collection manager not available");
      }

      if (!parentId) {
        throw new Error("Parent ID is required");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log("[useCollectionListing] Listing collections by parent:", {
          parentId,
          forceRefresh,
        });

        const result = await listCollectionManager.listCollectionsByParent(
          parentId,
          forceRefresh,
        );

        setCollectionsByParent((prev) => ({
          ...prev,
          [parentId]: result.collections,
        }));

        setSuccess(
          `Listed ${result.totalCount} collections for parent ${parentId} successfully from ${result.source}!`,
        );
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionListing] Collections by parent listed successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionListing] Collections by parent listing failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [listCollectionManager, loadManagerStatus],
  );

  // Get cached collections
  const getCachedCollections = useCallback(() => {
    if (!listCollectionManager) return { collections: [], isExpired: false };

    try {
      return listCollectionManager.getCachedCollections();
    } catch (err) {
      console.error(
        "[useCollectionListing] Failed to get cached collections:",
        err,
      );
      return { collections: [], isExpired: false };
    }
  }, [listCollectionManager]);

  // Get cached filtered collections
  const getCachedFilteredCollections = useCallback(() => {
    if (!listCollectionManager) {
      return {
        owned_collections: [],
        shared_collections: [],
        total_count: 0,
        isExpired: false,
      };
    }

    try {
      return listCollectionManager.getCachedFilteredCollections();
    } catch (err) {
      console.error(
        "[useCollectionListing] Failed to get cached filtered collections:",
        err,
      );
      return {
        owned_collections: [],
        shared_collections: [],
        total_count: 0,
        isExpired: false,
      };
    }
  }, [listCollectionManager]);

  // Clear all cache
  const clearAllCache = useCallback(async () => {
    if (!listCollectionManager) {
      throw new Error("List collection manager not available");
    }

    try {
      console.log("[useCollectionListing] Clearing all cache");

      await listCollectionManager.clearAllCache();
      setSuccess("All collection list cache cleared successfully!");

      // Reset state
      setCollections([]);
      setFilteredCollections({
        owned_collections: [],
        shared_collections: [],
        total_count: 0,
      });
      setRootCollections([]);
      setCollectionsByParent({});

      loadManagerStatus(); // Reload status
    } catch (err) {
      console.error("[useCollectionListing] Failed to clear all cache:", err);
      setError(`Failed to clear all cache: ${err.message}`);
      throw err;
    }
  }, [listCollectionManager, loadManagerStatus]);

  // Clear specific cache
  const clearSpecificCache = useCallback(
    async (cacheType) => {
      if (!listCollectionManager) {
        throw new Error("List collection manager not available");
      }

      try {
        console.log(
          "[useCollectionListing] Clearing specific cache:",
          cacheType,
        );

        const success =
          await listCollectionManager.clearSpecificCache(cacheType);

        if (success) {
          setSuccess(`${cacheType} cache cleared successfully!`);

          // Reset relevant state
          switch (cacheType) {
            case "listed":
              setCollections([]);
              break;
            case "filtered":
              setFilteredCollections({
                owned_collections: [],
                shared_collections: [],
                total_count: 0,
              });
              break;
            case "root":
              setRootCollections([]);
              break;
            case "byParent":
              setCollectionsByParent({});
              break;
          }

          loadManagerStatus(); // Reload status
        }

        return success;
      } catch (err) {
        console.error(
          "[useCollectionListing] Failed to clear specific cache:",
          err,
        );
        setError(`Failed to clear ${cacheType} cache: ${err.message}`);
        throw err;
      }
    },
    [listCollectionManager, loadManagerStatus],
  );

  // Search collections
  const searchCollections = useCallback(
    (searchTerm, collectionsToSearch = collections) => {
      if (!listCollectionManager) return [];

      try {
        return listCollectionManager.searchCollections(
          searchTerm,
          collectionsToSearch,
        );
      } catch (err) {
        console.error("[useCollectionListing] Search failed:", err);
        return [];
      }
    },
    [listCollectionManager, collections],
  );

  // Filter collections by type
  const filterCollectionsByType = useCallback(
    (collectionsToFilter = collections, type) => {
      if (!listCollectionManager) return [];

      try {
        return listCollectionManager.filterCollectionsByType(
          collectionsToFilter,
          type,
        );
      } catch (err) {
        console.error("[useCollectionListing] Filter failed:", err);
        return [];
      }
    },
    [listCollectionManager, collections],
  );

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!listCollectionManager) return null;

    try {
      return await listCollectionManager.getUserPassword();
    } catch (err) {
      console.error("[useCollectionListing] Failed to get user password:", err);
      return null;
    }
  }, [listCollectionManager]);

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
    setCollections([]);
    setFilteredCollections({
      owned_collections: [],
      shared_collections: [],
      total_count: 0,
    });
    setRootCollections([]);
    setCollectionsByParent({});
    loadManagerStatus();
  }, [loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (listCollectionManager) {
      loadManagerStatus();
    }
  }, [listCollectionManager, loadManagerStatus]);

  // Set up event listeners for collection listing events
  useEffect(() => {
    if (!listCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionListing] Collection event:",
        eventType,
        eventData,
      );

      // Reload status on certain events
      if (
        [
          "collections_listed_from_cache",
          "collections_listed_from_api",
          "filtered_collections_listed_from_cache",
          "filtered_collections_listed_from_api",
          "root_collections_listed_from_cache",
          "root_collections_listed_from_api",
          "collections_by_parent_listed_from_cache",
          "collections_by_parent_listed_from_api",
          "all_list_cache_cleared",
        ].includes(eventType)
      ) {
        loadManagerStatus();
      }
    };

    listCollectionManager.addCollectionListingListener(handleCollectionEvent);

    return () => {
      listCollectionManager.removeCollectionListingListener(
        handleCollectionEvent,
      );
    };
  }, [listCollectionManager, loadManagerStatus]);

  return {
    // State
    isLoading,
    error,
    success,
    collections,
    filteredCollections,
    rootCollections,
    collectionsByParent,
    managerStatus,

    // Core operations
    listCollections,
    listFilteredCollections,
    listRootCollections,
    listCollectionsByParent,

    // Cache operations
    getCachedCollections,
    getCachedFilteredCollections,
    clearAllCache,
    clearSpecificCache,

    // Utility operations
    searchCollections,
    filterCollectionsByType,
    getUserPassword,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canListCollections: managerStatus.canListCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Statistics
    totalCollections: collections.length,
    totalFilteredCollections: filteredCollections.total_count,
    totalRootCollections: rootCollections.length,
    collectionsByType: collections.reduce((acc, col) => {
      const type = col.collection_type || "unknown";
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {}),

    // Cache status
    hasCachedCollections: managerStatus.storage?.hasListedCollections || false,
    hasCachedFilteredCollections:
      managerStatus.storage?.hasFilteredCollections || false,
    hasCachedRootCollections:
      managerStatus.storage?.hasRootCollections || false,

    // Helper methods
    getCollectionsByParent: (parentId) => collectionsByParent[parentId] || [],
    getFolders: () => filterCollectionsByType(collections, "folder"),
    getAlbums: () => filterCollectionsByType(collections, "album"),
  };
};

export default useCollectionListing;

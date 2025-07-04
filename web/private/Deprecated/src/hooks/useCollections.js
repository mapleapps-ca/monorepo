// Custom hook for collection management
import { useState, useEffect, useCallback } from "react";
import { useServices } from "./useService.jsx";

const useCollections = () => {
  const { collectionService, authService } = useServices();
  const [collections, setCollections] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  // Load user's collections
  const loadUserCollections = useCallback(async () => {
    if (!authService.isAuthenticated()) {
      console.log("[useCollections] User not authenticated, skipping load");
      return [];
    }

    try {
      setIsLoading(true);
      setError(null);

      const userCollections = await collectionService.listUserCollections();
      setCollections(userCollections);

      return userCollections;
    } catch (err) {
      console.error("[useCollections] Failed to load user collections:", err);
      setError(err.message);
      return [];
    } finally {
      setIsLoading(false);
    }
  }, [collectionService, authService]);

  // Load shared collections
  const loadSharedCollections = useCallback(async () => {
    if (!authService.isAuthenticated()) {
      return [];
    }

    try {
      setIsLoading(true);
      setError(null);

      const shared = await collectionService.listSharedCollections();
      setSharedCollections(shared);

      return shared;
    } catch (err) {
      console.error("[useCollections] Failed to load shared collections:", err);
      setError(err.message);
      return [];
    } finally {
      setIsLoading(false);
    }
  }, [collectionService, authService]);

  // Load all collections (owned + shared)
  const loadAllCollections = useCallback(async () => {
    if (!authService.isAuthenticated()) {
      return { owned: [], shared: [] };
    }

    try {
      setIsLoading(true);
      setError(null);

      const result = await collectionService.getFilteredCollections(true, true);

      setCollections(result.owned_collections || []);
      setSharedCollections(result.shared_collections || []);

      return result;
    } catch (err) {
      console.error("[useCollections] Failed to load all collections:", err);
      setError(err.message);
      return { owned_collections: [], shared_collections: [] };
    } finally {
      setIsLoading(false);
    }
  }, [collectionService, authService]);

  // Create a new collection
  const createCollection = useCallback(
    async (collectionData) => {
      try {
        setIsLoading(true);
        setError(null);

        const newCollection =
          await collectionService.createCollection(collectionData);

        // Add to local state
        setCollections((prev) => [...prev, newCollection]);

        return newCollection;
      } catch (err) {
        console.error("[useCollections] Failed to create collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Update a collection
  const updateCollection = useCallback(
    async (collectionId, updateData) => {
      try {
        setIsLoading(true);
        setError(null);

        const updatedCollection = await collectionService.updateCollection(
          collectionId,
          updateData,
        );

        // Update local state
        setCollections((prev) =>
          prev.map((col) =>
            col.id === collectionId ? updatedCollection : col,
          ),
        );

        return updatedCollection;
      } catch (err) {
        console.error("[useCollections] Failed to update collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Delete a collection
  const deleteCollection = useCallback(
    async (collectionId) => {
      try {
        setIsLoading(true);
        setError(null);

        await collectionService.deleteCollection(collectionId);

        // Remove from local state
        setCollections((prev) => prev.filter((col) => col.id !== collectionId));
        setSharedCollections((prev) =>
          prev.filter((col) => col.id !== collectionId),
        );

        return true;
      } catch (err) {
        console.error("[useCollections] Failed to delete collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Archive a collection
  const archiveCollection = useCallback(
    async (collectionId) => {
      try {
        setIsLoading(true);
        setError(null);

        await collectionService.archiveCollection(collectionId);

        // Update local state
        setCollections((prev) =>
          prev.map((col) =>
            col.id === collectionId ? { ...col, state: "archived" } : col,
          ),
        );

        return true;
      } catch (err) {
        console.error("[useCollections] Failed to archive collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Restore a collection
  const restoreCollection = useCallback(
    async (collectionId) => {
      try {
        setIsLoading(true);
        setError(null);

        await collectionService.restoreCollection(collectionId);

        // Update local state
        setCollections((prev) =>
          prev.map((col) =>
            col.id === collectionId ? { ...col, state: "active" } : col,
          ),
        );

        return true;
      } catch (err) {
        console.error("[useCollections] Failed to restore collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Share a collection
  const shareCollection = useCallback(
    async (collectionId, shareData) => {
      try {
        setIsLoading(true);
        setError(null);

        const result = await collectionService.shareCollection(
          collectionId,
          shareData,
        );

        // Reload collections to reflect sharing changes
        await loadAllCollections();

        return result;
      } catch (err) {
        console.error("[useCollections] Failed to share collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService, loadAllCollections],
  );

  // Remove member from collection
  const removeMember = useCallback(
    async (collectionId, recipientId, removeFromDescendants = true) => {
      try {
        setIsLoading(true);
        setError(null);

        await collectionService.removeMember(
          collectionId,
          recipientId,
          removeFromDescendants,
        );

        // Reload collections to reflect membership changes
        await loadAllCollections();

        return true;
      } catch (err) {
        console.error("[useCollections] Failed to remove member:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService, loadAllCollections],
  );

  // Move collection
  const moveCollection = useCallback(
    async (collectionId, moveData) => {
      try {
        setIsLoading(true);
        setError(null);

        await collectionService.moveCollection(collectionId, moveData);

        // Reload collections to reflect hierarchy changes
        await loadAllCollections();

        return true;
      } catch (err) {
        console.error("[useCollections] Failed to move collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService, loadAllCollections],
  );

  // Get collection by ID (from cache or API)
  const getCollection = useCallback(
    async (collectionId) => {
      try {
        setIsLoading(true);
        setError(null);

        return await collectionService.getCollection(collectionId);
      } catch (err) {
        console.error("[useCollections] Failed to get collection:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Get root collections
  const getRootCollections = useCallback(() => {
    return collections.filter((col) => !col.parent_id);
  }, [collections]);

  // Get child collections
  const getChildCollections = useCallback(
    (parentId) => {
      return collections.filter((col) => col.parent_id === parentId);
    },
    [collections],
  );

  // Get collection tree
  const getCollectionTree = useCallback(
    async (parentId = null) => {
      try {
        setIsLoading(true);
        setError(null);

        return await collectionService.getCollectionTree(parentId);
      } catch (err) {
        console.error("[useCollections] Failed to get collection tree:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Get collection hierarchy (path to root)
  const getCollectionHierarchy = useCallback(
    async (collectionId) => {
      try {
        setIsLoading(true);
        setError(null);

        return await collectionService.getCollectionHierarchy(collectionId);
      } catch (err) {
        console.error(
          "[useCollections] Failed to get collection hierarchy:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Sync collections for offline support
  const syncCollections = useCallback(
    async (cursor = null, limit = 1000) => {
      try {
        setIsLoading(true);
        setError(null);

        return await collectionService.syncCollections(cursor, limit);
      } catch (err) {
        console.error("[useCollections] Failed to sync collections:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [collectionService],
  );

  // Clear cache
  const clearCache = useCallback(() => {
    collectionService.clearCache();
    setCollections([]);
    setSharedCollections([]);
  }, [collectionService]);

  // Initial load when authenticated
  useEffect(() => {
    if (authService.isAuthenticated()) {
      loadAllCollections();
    } else {
      // Clear collections when not authenticated
      setCollections([]);
      setSharedCollections([]);
    }
  }, [authService.isAuthenticated, loadAllCollections]);

  return {
    // State
    collections,
    sharedCollections,
    allCollections: [...collections, ...sharedCollections],
    isLoading,
    error,

    // Collection operations
    loadUserCollections,
    loadSharedCollections,
    loadAllCollections,
    createCollection,
    updateCollection,
    deleteCollection,
    archiveCollection,
    restoreCollection,
    shareCollection,
    removeMember,
    moveCollection,
    getCollection,
    syncCollections,

    // Utility functions
    getRootCollections,
    getChildCollections,
    getCollectionTree,
    getCollectionHierarchy,
    clearCache,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Collection states
    COLLECTION_STATES: {
      ACTIVE: "active",
      DELETED: "deleted",
      ARCHIVED: "archived",
    },

    // Permission levels
    PERMISSION_LEVELS: {
      READ_ONLY: "read_only",
      READ_WRITE: "read_write",
      ADMIN: "admin",
    },
  };
};

export default useCollections;

// File: monorepo/web/maplefile-frontend/src/hooks/useCollectionCreation.jsx
// Custom hook for collection creation with convenient API
import { useState, useEffect, useCallback } from "react";
import { useCollections, useAuth } from "./useService.jsx";

/**
 * Hook for collection creation with state management and convenience methods
 * @returns {Object} Collection creation API
 */
const useCollectionCreation = () => {
  const { createCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [collections, setCollections] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Load collections from storage
  const loadCollections = useCallback(() => {
    if (!createCollectionManager) return;

    try {
      const storedCollections = createCollectionManager.getCreatedCollections();
      setCollections(storedCollections);
      console.log(
        "[useCollectionCreation] Collections loaded:",
        storedCollections.length,
      );
    } catch (err) {
      console.error("[useCollectionCreation] Failed to load collections:", err);
      setError(`Failed to load collections: ${err.message}`);
    }
  }, [createCollectionManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!createCollectionManager) return;

    try {
      const status = createCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionCreation] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionCreation] Failed to load manager status:",
        err,
      );
    }
  }, [createCollectionManager]);

  // Create collection with enhanced error handling
  const createCollection = useCallback(
    async (collectionData, password = null) => {
      if (!createCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionCreation] Creating collection:",
          collectionData,
        );

        const result = await createCollectionManager.createCollection(
          collectionData,
          password,
        );

        setSuccess(`Collection "${collectionData.name}" created successfully!`);
        loadCollections(); // Reload collections
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionCreation] Collection created successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionCreation] Collection creation failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [createCollectionManager, loadCollections, loadManagerStatus],
  );

  // Create folder (convenience method)
  const createFolder = useCallback(
    async (name, password = null, parentId = null) => {
      return createCollection(
        {
          name,
          collection_type: "folder",
          parent_id: parentId,
        },
        password,
      );
    },
    [createCollection],
  );

  // Create album (convenience method)
  const createAlbum = useCallback(
    async (name, password = null, parentId = null) => {
      return createCollection(
        {
          name,
          collection_type: "album",
          parent_id: parentId,
        },
        password,
      );
    },
    [createCollection],
  );

  // Decrypt collection
  const decryptCollection = useCallback(
    async (collection, password = null) => {
      if (!createCollectionManager) {
        throw new Error("Collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionCreation] Decrypting collection:",
          collection.id,
        );

        const decrypted = await createCollectionManager.decryptCollection(
          collection,
          password,
        );

        console.log("[useCollectionCreation] Collection decrypted:", decrypted);
        return decrypted;
      } catch (err) {
        console.error("[useCollectionCreation] Decryption failed:", err);
        setError(`Decryption failed: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [createCollectionManager],
  );

  // Remove collection
  const removeCollection = useCallback(
    async (collectionId) => {
      if (!createCollectionManager) {
        throw new Error("Collection manager not available");
      }

      try {
        console.log(
          "[useCollectionCreation] Removing collection:",
          collectionId,
        );

        await createCollectionManager.removeCollection(collectionId);
        setSuccess("Collection removed successfully!");
        loadCollections(); // Reload collections
      } catch (err) {
        console.error(
          "[useCollectionCreation] Failed to remove collection:",
          err,
        );
        setError(`Failed to remove collection: ${err.message}`);
        throw err;
      }
    },
    [createCollectionManager, loadCollections],
  );

  // Clear all collections
  const clearAllCollections = useCallback(async () => {
    if (!createCollectionManager) {
      throw new Error("Collection manager not available");
    }

    try {
      console.log("[useCollectionCreation] Clearing all collections");

      await createCollectionManager.clearAllCollections();
      setSuccess("All collections cleared successfully!");
      loadCollections(); // Reload collections
    } catch (err) {
      console.error(
        "[useCollectionCreation] Failed to clear collections:",
        err,
      );
      setError(`Failed to clear collections: ${err.message}`);
      throw err;
    }
  }, [createCollectionManager, loadCollections]);

  // Search collections
  const searchCollections = useCallback(
    (searchTerm) => {
      if (!createCollectionManager) return [];

      try {
        return createCollectionManager.searchCollections(searchTerm);
      } catch (err) {
        console.error("[useCollectionCreation] Search failed:", err);
        return [];
      }
    },
    [createCollectionManager],
  );

  // Get collection by ID
  const getCollectionById = useCallback(
    (collectionId) => {
      if (!createCollectionManager) return null;

      return createCollectionManager.getCollectionById(collectionId);
    },
    [createCollectionManager],
  );

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!createCollectionManager) return null;

    try {
      return await createCollectionManager.getUserPassword();
    } catch (err) {
      console.error(
        "[useCollectionCreation] Failed to get user password:",
        err,
      );
      return null;
    }
  }, [createCollectionManager]);

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
    loadCollections();
    loadManagerStatus();
  }, [loadCollections, loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (createCollectionManager) {
      loadCollections();
      loadManagerStatus();
    }
  }, [createCollectionManager, loadCollections, loadManagerStatus]);

  // Set up event listeners for collection events
  useEffect(() => {
    if (!createCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionCreation] Collection event:",
        eventType,
        eventData,
      );

      // Reload collections on certain events
      if (
        [
          "collection_created",
          "collection_removed",
          "all_collections_cleared",
        ].includes(eventType)
      ) {
        loadCollections();
        loadManagerStatus();
      }
    };

    createCollectionManager.addCollectionCreationListener(
      handleCollectionEvent,
    );

    return () => {
      createCollectionManager.removeCollectionCreationListener(
        handleCollectionEvent,
      );
    };
  }, [createCollectionManager, loadCollections, loadManagerStatus]);

  return {
    // State
    isLoading,
    error,
    success,
    collections,
    managerStatus,

    // Core operations
    createCollection,
    createFolder,
    createAlbum,
    decryptCollection,
    removeCollection,
    clearAllCollections,

    // Utility operations
    searchCollections,
    getCollectionById,
    getUserPassword,
    loadCollections,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canCreateCollections: managerStatus.canCreateCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Statistics
    totalCollections: collections.length,
    collectionsByType: collections.reduce((acc, col) => {
      const type = col.collection_type || "unknown";
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {}),
  };
};

export default useCollectionCreation;

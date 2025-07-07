// File: monorepo/web/maplefile-frontend/src/hooks/Collection/useCollectionUpdate.jsx
// Custom hook for collection updates with convenient API
import { useState, useEffect, useCallback } from "react";
import { useCollections, useAuth } from "../useService.jsx";

/**
 * Hook for collection updates with state management and convenience methods
 * @returns {Object} Collection update API
 */
const useCollectionUpdate = () => {
  const { updateCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [updatedCollections, setUpdatedCollections] = useState([]);
  const [updateHistory, setUpdateHistory] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Load updated collections from storage
  const loadUpdatedCollections = useCallback(() => {
    if (!updateCollectionManager) return;

    try {
      const storedCollections = updateCollectionManager.getUpdatedCollections();
      setUpdatedCollections(storedCollections);
      console.log(
        "[useCollectionUpdate] Updated collections loaded:",
        storedCollections.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionUpdate] Failed to load updated collections:",
        err,
      );
      setError(`Failed to load updated collections: ${err.message}`);
    }
  }, [updateCollectionManager]);

  // Load update history
  const loadUpdateHistory = useCallback(() => {
    if (!updateCollectionManager) return;

    try {
      const history = updateCollectionManager.getUpdateHistory();
      setUpdateHistory(history);
      console.log(
        "[useCollectionUpdate] Update history loaded:",
        history.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionUpdate] Failed to load update history:",
        err,
      );
    }
  }, [updateCollectionManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!updateCollectionManager) return;

    try {
      const status = updateCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionUpdate] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionUpdate] Failed to load manager status:",
        err,
      );
    }
  }, [updateCollectionManager]);

  // Update collection with enhanced error handling
  const updateCollection = useCallback(
    async (collectionId, updateData, password = null) => {
      if (!updateCollectionManager) {
        throw new Error("Update collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionUpdate] Updating collection:",
          collectionId,
          updateData,
        );

        const result = await updateCollectionManager.updateCollection(
          collectionId,
          updateData,
          password,
        );

        setSuccess(
          `Collection "${updateData.name || "collection"}" updated successfully!`,
        );
        loadUpdatedCollections(); // Reload collections
        loadUpdateHistory(); // Reload history
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionUpdate] Collection updated successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error("[useCollectionUpdate] Collection update failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      updateCollectionManager,
      loadUpdatedCollections,
      loadUpdateHistory,
      loadManagerStatus,
    ],
  );

  // Update collection name only (convenience method)
  const updateCollectionName = useCallback(
    async (collectionId, newName, version, password = null) => {
      return updateCollection(
        collectionId,
        {
          name: newName,
          version: version,
        },
        password,
      );
    },
    [updateCollection],
  );

  // Update collection type only (convenience method)
  const updateCollectionType = useCallback(
    async (collectionId, newType, version, password = null) => {
      const validTypes = ["folder", "album"];
      if (!validTypes.includes(newType)) {
        throw new Error(`Invalid collection type: ${newType}`);
      }

      return updateCollection(
        collectionId,
        {
          collection_type: newType,
          version: version,
        },
        password,
      );
    },
    [updateCollection],
  );

  // Rotate collection key (convenience method)
  const rotateCollectionKey = useCallback(
    async (collectionId, version, password = null) => {
      return updateCollection(
        collectionId,
        {
          version: version,
          rotateCollectionKey: true,
        },
        password,
      );
    },
    [updateCollection],
  );

  // Decrypt collection (uses PasswordStorageService automatically)
  const decryptCollection = useCallback(
    async (collection, password = null) => {
      if (!updateCollectionManager) {
        throw new Error("Update collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionUpdate] Decrypting collection:",
          collection.id,
        );

        const decrypted = await updateCollectionManager.decryptCollection(
          collection,
          password,
        );

        console.log("[useCollectionUpdate] Collection decrypted:", decrypted);
        return decrypted;
      } catch (err) {
        console.error("[useCollectionUpdate] Decryption failed:", err);
        setError(`Decryption failed: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [updateCollectionManager],
  );

  // Remove updated collection from local storage
  const removeUpdatedCollection = useCallback(
    async (collectionId) => {
      if (!updateCollectionManager) {
        throw new Error("Update collection manager not available");
      }

      try {
        console.log(
          "[useCollectionUpdate] Removing updated collection:",
          collectionId,
        );

        await updateCollectionManager.removeUpdatedCollection(collectionId);
        setSuccess("Updated collection removed successfully!");
        loadUpdatedCollections(); // Reload collections
        loadUpdateHistory(); // Reload history
      } catch (err) {
        console.error(
          "[useCollectionUpdate] Failed to remove updated collection:",
          err,
        );
        setError(`Failed to remove updated collection: ${err.message}`);
        throw err;
      }
    },
    [updateCollectionManager, loadUpdatedCollections, loadUpdateHistory],
  );

  // Clear all updated collections
  const clearAllUpdatedCollections = useCallback(async () => {
    if (!updateCollectionManager) {
      throw new Error("Update collection manager not available");
    }

    try {
      console.log("[useCollectionUpdate] Clearing all updated collections");

      await updateCollectionManager.clearAllUpdatedCollections();
      setSuccess("All updated collections cleared successfully!");
      loadUpdatedCollections(); // Reload collections
      loadUpdateHistory(); // Reload history
    } catch (err) {
      console.error(
        "[useCollectionUpdate] Failed to clear updated collections:",
        err,
      );
      setError(`Failed to clear updated collections: ${err.message}`);
      throw err;
    }
  }, [updateCollectionManager, loadUpdatedCollections, loadUpdateHistory]);

  // Search updated collections
  const searchUpdatedCollections = useCallback(
    (searchTerm) => {
      if (!updateCollectionManager) return [];

      try {
        return updateCollectionManager.searchUpdatedCollections(searchTerm);
      } catch (err) {
        console.error("[useCollectionUpdate] Search failed:", err);
        return [];
      }
    },
    [updateCollectionManager],
  );

  // Get updated collection by ID
  const getUpdatedCollectionById = useCallback(
    (collectionId) => {
      if (!updateCollectionManager) return null;

      return updateCollectionManager.getUpdatedCollectionById(collectionId);
    },
    [updateCollectionManager],
  );

  // Get update history for specific collection
  const getCollectionUpdateHistory = useCallback(
    (collectionId) => {
      if (!updateCollectionManager) return [];

      return updateCollectionManager.getUpdateHistory(collectionId);
    },
    [updateCollectionManager],
  );

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!updateCollectionManager) return null;

    try {
      return await updateCollectionManager.getUserPassword();
    } catch (err) {
      console.error("[useCollectionUpdate] Failed to get user password:", err);
      return null;
    }
  }, [updateCollectionManager]);

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
    loadUpdatedCollections();
    loadUpdateHistory();
    loadManagerStatus();
  }, [loadUpdatedCollections, loadUpdateHistory, loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (updateCollectionManager) {
      loadUpdatedCollections();
      loadUpdateHistory();
      loadManagerStatus();
    }
  }, [
    updateCollectionManager,
    loadUpdatedCollections,
    loadUpdateHistory,
    loadManagerStatus,
  ]);

  // Set up event listeners for collection update events
  useEffect(() => {
    if (!updateCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionUpdate] Collection event:",
        eventType,
        eventData,
      );

      // Reload data on certain events
      if (
        [
          "collection_updated",
          "updated_collection_removed",
          "all_updated_collections_cleared",
        ].includes(eventType)
      ) {
        loadUpdatedCollections();
        loadUpdateHistory();
        loadManagerStatus();
      }
    };

    updateCollectionManager.addCollectionUpdateListener(handleCollectionEvent);

    return () => {
      updateCollectionManager.removeCollectionUpdateListener(
        handleCollectionEvent,
      );
    };
  }, [
    updateCollectionManager,
    loadUpdatedCollections,
    loadUpdateHistory,
    loadManagerStatus,
  ]);

  return {
    // State
    isLoading,
    error,
    success,
    updatedCollections,
    updateHistory,
    managerStatus,

    // Core operations
    updateCollection,
    updateCollectionName,
    updateCollectionType,
    rotateCollectionKey,
    decryptCollection,
    removeUpdatedCollection,
    clearAllUpdatedCollections,

    // Utility operations
    searchUpdatedCollections,
    getUpdatedCollectionById,
    getCollectionUpdateHistory,
    getUserPassword,
    loadUpdatedCollections,
    loadUpdateHistory,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canUpdateCollections: managerStatus.canUpdateCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Statistics
    totalUpdatedCollections: updatedCollections.length,
    totalUpdateHistory: updateHistory.length,
    updatedCollectionsByType: updatedCollections.reduce((acc, col) => {
      const type = col.collection_type || "unknown";
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {}),

    // Helper methods
    getLatestUpdateForCollection: (collectionId) => {
      return (
        updateHistory
          .filter((h) => h.collectionId === collectionId)
          .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))[0] ||
        null
      );
    },

    getRecentUpdates: (hours = 24) => {
      const cutoff = new Date(Date.now() - hours * 60 * 60 * 1000);
      return updateHistory.filter((h) => new Date(h.timestamp) > cutoff);
    },
  };
};

export default useCollectionUpdate;

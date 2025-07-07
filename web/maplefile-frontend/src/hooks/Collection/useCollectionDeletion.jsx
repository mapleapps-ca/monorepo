// File: monorepo/web/maplefile-frontend/src/hooks/Collection/useCollectionDeletion.jsx
// Custom hook for collection deletion with convenient API

import { useState, useEffect, useCallback } from "react";
import { useCollections, useAuth } from "../useService.jsx";

/**
 * Hook for collection deletion with state management and convenience methods
 * @returns {Object} Collection deletion API
 */
const useCollectionDeletion = () => {
  const { deleteCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [deletedCollections, setDeletedCollections] = useState([]);
  const [deletionHistory, setDeletionHistory] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Load deleted collections from storage
  const loadDeletedCollections = useCallback(() => {
    if (!deleteCollectionManager) return;

    try {
      const storedCollections = deleteCollectionManager.getDeletedCollections();
      setDeletedCollections(storedCollections);
      console.log(
        "[useCollectionDeletion] Deleted collections loaded:",
        storedCollections.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionDeletion] Failed to load deleted collections:",
        err,
      );
      setError(`Failed to load deleted collections: ${err.message}`);
    }
  }, [deleteCollectionManager]);

  // Load deletion history
  const loadDeletionHistory = useCallback(() => {
    if (!deleteCollectionManager) return;

    try {
      const history = deleteCollectionManager.getDeletionHistory();
      setDeletionHistory(history);
      console.log(
        "[useCollectionDeletion] Deletion history loaded:",
        history.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionDeletion] Failed to load deletion history:",
        err,
      );
    }
  }, [deleteCollectionManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!deleteCollectionManager) return;

    try {
      const status = deleteCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionDeletion] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionDeletion] Failed to load manager status:",
        err,
      );
    }
  }, [deleteCollectionManager]);

  // Delete collection with enhanced error handling
  const deleteCollection = useCallback(
    async (collectionId, password = null) => {
      if (!deleteCollectionManager) {
        throw new Error("Delete collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionDeletion] Deleting collection:",
          collectionId,
        );

        const result = await deleteCollectionManager.deleteCollection(
          collectionId,
          password,
        );

        setSuccess(
          `Collection "${result.collection.name || "collection"}" deleted successfully!`,
        );
        loadDeletedCollections(); // Reload collections
        loadDeletionHistory(); // Reload history
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionDeletion] Collection deleted successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionDeletion] Collection deletion failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      deleteCollectionManager,
      loadDeletedCollections,
      loadDeletionHistory,
      loadManagerStatus,
    ],
  );

  // Delete multiple collections (batch operation)
  const deleteCollections = useCallback(
    async (collectionIds, password = null) => {
      if (!deleteCollectionManager) {
        throw new Error("Delete collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionDeletion] Deleting multiple collections:",
          collectionIds.length,
        );

        const result = await deleteCollectionManager.deleteCollections(
          collectionIds,
          password,
        );

        setSuccess(
          `Deleted ${result.successCount}/${collectionIds.length} collections successfully!`,
        );
        loadDeletedCollections(); // Reload collections
        loadDeletionHistory(); // Reload history
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionDeletion] Multiple collections deleted:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionDeletion] Multiple collections deletion failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      deleteCollectionManager,
      loadDeletedCollections,
      loadDeletionHistory,
      loadManagerStatus,
    ],
  );

  // Restore collection from soft delete
  const restoreCollection = useCallback(
    async (collectionId) => {
      if (!deleteCollectionManager) {
        throw new Error("Delete collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionDeletion] Restoring collection:",
          collectionId,
        );

        const result =
          await deleteCollectionManager.restoreCollection(collectionId);

        setSuccess(
          `Collection "${result.collection.name || "collection"}" restored successfully!`,
        );
        loadDeletedCollections(); // Reload collections
        loadDeletionHistory(); // Reload history
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionDeletion] Collection restored successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error(
          "[useCollectionDeletion] Collection restoration failed:",
          err,
        );
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      deleteCollectionManager,
      loadDeletedCollections,
      loadDeletionHistory,
      loadManagerStatus,
    ],
  );

  // Decrypt collection (uses PasswordStorageService automatically)
  const decryptCollection = useCallback(
    async (collection, password = null) => {
      if (!deleteCollectionManager) {
        throw new Error("Delete collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionDeletion] Decrypting collection:",
          collection.id,
        );

        const decrypted = await deleteCollectionManager.decryptCollection(
          collection,
          password,
        );

        console.log("[useCollectionDeletion] Collection decrypted:", decrypted);
        return decrypted;
      } catch (err) {
        console.error("[useCollectionDeletion] Decryption failed:", err);
        setError(`Decryption failed: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [deleteCollectionManager],
  );

  // Permanently remove deleted collection from local storage
  const permanentlyRemoveCollection = useCallback(
    async (collectionId) => {
      if (!deleteCollectionManager) {
        throw new Error("Delete collection manager not available");
      }

      try {
        console.log(
          "[useCollectionDeletion] Permanently removing collection:",
          collectionId,
        );

        await deleteCollectionManager.permanentlyRemoveCollection(collectionId);
        setSuccess("Collection permanently removed from local storage!");
        loadDeletedCollections(); // Reload collections
        loadDeletionHistory(); // Reload history
      } catch (err) {
        console.error(
          "[useCollectionDeletion] Failed to permanently remove collection:",
          err,
        );
        setError(`Failed to permanently remove collection: ${err.message}`);
        throw err;
      }
    },
    [deleteCollectionManager, loadDeletedCollections, loadDeletionHistory],
  );

  // Clear all deleted collections
  const clearAllDeletedCollections = useCallback(async () => {
    if (!deleteCollectionManager) {
      throw new Error("Delete collection manager not available");
    }

    try {
      console.log("[useCollectionDeletion] Clearing all deleted collections");

      await deleteCollectionManager.clearAllDeletedCollections();
      setSuccess("All deleted collections cleared successfully!");
      loadDeletedCollections(); // Reload collections
      loadDeletionHistory(); // Reload history
    } catch (err) {
      console.error(
        "[useCollectionDeletion] Failed to clear deleted collections:",
        err,
      );
      setError(`Failed to clear deleted collections: ${err.message}`);
      throw err;
    }
  }, [deleteCollectionManager, loadDeletedCollections, loadDeletionHistory]);

  // Search deleted collections
  const searchDeletedCollections = useCallback(
    (searchTerm) => {
      if (!deleteCollectionManager) return [];

      try {
        return deleteCollectionManager.searchDeletedCollections(searchTerm);
      } catch (err) {
        console.error("[useCollectionDeletion] Search failed:", err);
        return [];
      }
    },
    [deleteCollectionManager],
  );

  // Get deleted collection by ID
  const getDeletedCollectionById = useCallback(
    (collectionId) => {
      if (!deleteCollectionManager) return null;

      return deleteCollectionManager.getDeletedCollectionById(collectionId);
    },
    [deleteCollectionManager],
  );

  // Get deletion history for specific collection
  const getCollectionDeletionHistory = useCallback(
    (collectionId) => {
      if (!deleteCollectionManager) return [];

      return deleteCollectionManager.getDeletionHistory(collectionId);
    },
    [deleteCollectionManager],
  );

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!deleteCollectionManager) return null;

    try {
      return await deleteCollectionManager.getUserPassword();
    } catch (err) {
      console.error(
        "[useCollectionDeletion] Failed to get user password:",
        err,
      );
      return null;
    }
  }, [deleteCollectionManager]);

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
    loadDeletedCollections();
    loadDeletionHistory();
    loadManagerStatus();
  }, [loadDeletedCollections, loadDeletionHistory, loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (deleteCollectionManager) {
      loadDeletedCollections();
      loadDeletionHistory();
      loadManagerStatus();
    }
  }, [
    deleteCollectionManager,
    loadDeletedCollections,
    loadDeletionHistory,
    loadManagerStatus,
  ]);

  // Set up event listeners for collection deletion events
  useEffect(() => {
    if (!deleteCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionDeletion] Collection event:",
        eventType,
        eventData,
      );

      // Reload data on certain events
      if (
        [
          "collection_deleted",
          "collection_restored",
          "multiple_collections_deleted",
          "deleted_collection_permanently_removed",
          "all_deleted_collections_cleared",
        ].includes(eventType)
      ) {
        loadDeletedCollections();
        loadDeletionHistory();
        loadManagerStatus();
      }
    };

    deleteCollectionManager.addCollectionDeletionListener(
      handleCollectionEvent,
    );

    return () => {
      deleteCollectionManager.removeCollectionDeletionListener(
        handleCollectionEvent,
      );
    };
  }, [
    deleteCollectionManager,
    loadDeletedCollections,
    loadDeletionHistory,
    loadManagerStatus,
  ]);

  return {
    // State
    isLoading,
    error,
    success,
    deletedCollections,
    deletionHistory,
    managerStatus,

    // Core operations
    deleteCollection,
    deleteCollections,
    restoreCollection,
    decryptCollection,
    permanentlyRemoveCollection,
    clearAllDeletedCollections,

    // Utility operations
    searchDeletedCollections,
    getDeletedCollectionById,
    getCollectionDeletionHistory,
    getUserPassword,
    loadDeletedCollections,
    loadDeletionHistory,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canDeleteCollections: managerStatus.canDeleteCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Collection types
    COLLECTION_TYPES: {
      FOLDER: "folder",
      ALBUM: "album",
    },

    // Statistics
    totalDeletedCollections: deletedCollections.length,
    totalDeletionHistory: deletionHistory.length,
    deletedCollectionsByType: deletedCollections.reduce((acc, col) => {
      const type = col.collection_type || "unknown";
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {}),

    // Helper methods
    getLatestDeletionForCollection: (collectionId) => {
      return (
        deletionHistory
          .filter((h) => h.collectionId === collectionId)
          .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))[0] ||
        null
      );
    },

    getRecentDeletions: (hours = 24) => {
      const cutoff = new Date(Date.now() - hours * 60 * 60 * 1000);
      return deletionHistory.filter((h) => new Date(h.timestamp) > cutoff);
    },
  };
};

export default useCollectionDeletion;

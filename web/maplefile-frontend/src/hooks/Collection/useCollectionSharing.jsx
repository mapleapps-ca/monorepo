// File: monorepo/web/maplefile-frontend/src/hooks/Collection/useCollectionSharing.jsx
// Custom hook for collection sharing with convenient API

import { useState, useEffect, useCallback } from "react";
import { useCollections, useAuthServices } from "../useService.jsx";

/**
 * Hook for collection sharing with state management and convenience methods
 * @returns {Object} Collection sharing API
 */
const useCollectionSharing = () => {
  const { shareCollectionManager } = useCollections();
  const { authManager } = useAuthServices();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [collectionMembers, setCollectionMembers] = useState({});
  const [sharingHistory, setSharingHistory] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Load shared collections from storage
  const loadSharedCollections = useCallback(() => {
    if (!shareCollectionManager) return;

    try {
      const stored = shareCollectionManager.getSharedCollections();
      setSharedCollections(stored);
      console.log(
        "[useCollectionSharing] Shared collections loaded:",
        stored.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionSharing] Failed to load shared collections:",
        err,
      );
      setError(`Failed to load shared collections: ${err.message}`);
    }
  }, [shareCollectionManager]);

  // Load sharing history
  const loadSharingHistory = useCallback(() => {
    if (!shareCollectionManager) return;

    try {
      const history = shareCollectionManager.getSharingHistory();
      setSharingHistory(history);
      console.log(
        "[useCollectionSharing] Sharing history loaded:",
        history.length,
      );
    } catch (err) {
      console.error(
        "[useCollectionSharing] Failed to load sharing history:",
        err,
      );
    }
  }, [shareCollectionManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!shareCollectionManager) return;

    try {
      const status = shareCollectionManager.getManagerStatus();
      setManagerStatus(status);
      console.log("[useCollectionSharing] Manager status loaded:", status);
    } catch (err) {
      console.error(
        "[useCollectionSharing] Failed to load manager status:",
        err,
      );
    }
  }, [shareCollectionManager]);

  // Share collection with enhanced error handling
  const shareCollection = useCallback(
    async (collectionId, shareData, password = null) => {
      if (!shareCollectionManager) {
        throw new Error("Share collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionSharing] Sharing collection:",
          collectionId,
          "->",
          shareData.recipient_email,
        );

        const result = await shareCollectionManager.shareCollection(
          collectionId,
          shareData,
          password,
        );

        setSuccess(
          `Collection shared successfully with ${shareData.recipient_email}!`,
        );
        loadSharedCollections(); // Reload shared collections
        loadSharingHistory(); // Reload history
        loadManagerStatus(); // Reload status

        console.log(
          "[useCollectionSharing] Collection shared successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error("[useCollectionSharing] Collection sharing failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      shareCollectionManager,
      loadSharedCollections,
      loadSharingHistory,
      loadManagerStatus,
    ],
  );

  // Remove member from collection
  const removeMember = useCallback(
    async (collectionId, recipientId, removeFromDescendants = true) => {
      if (!shareCollectionManager) {
        throw new Error("Share collection manager not available");
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        console.log(
          "[useCollectionSharing] Removing member:",
          collectionId,
          "->",
          recipientId,
        );

        const result = await shareCollectionManager.removeMember(
          collectionId,
          recipientId,
          removeFromDescendants,
        );

        setSuccess("Member removed successfully!");
        loadSharedCollections(); // Reload shared collections
        loadSharingHistory(); // Reload history
        loadManagerStatus(); // Reload status

        // Update collection members if we have them cached
        setCollectionMembers((prev) => ({
          ...prev,
          [collectionId]: (prev[collectionId] || []).filter(
            (member) => member.recipient_id !== recipientId,
          ),
        }));

        console.log(
          "[useCollectionSharing] Member removed successfully:",
          result,
        );
        return result;
      } catch (err) {
        console.error("[useCollectionSharing] Member removal failed:", err);
        setError(err.message);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [
      shareCollectionManager,
      loadSharedCollections,
      loadSharingHistory,
      loadManagerStatus,
    ],
  );

  // Get collection members
  const getCollectionMembers = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!shareCollectionManager) {
        throw new Error("Share collection manager not available");
      }

      setIsLoading(true);
      setError(null);

      try {
        console.log(
          "[useCollectionSharing] Getting collection members:",
          collectionId,
        );

        const members = await shareCollectionManager.getCollectionMembers(
          collectionId,
          forceRefresh,
        );

        setCollectionMembers((prev) => ({
          ...prev,
          [collectionId]: members,
        }));

        console.log(
          "[useCollectionSharing] Collection members retrieved:",
          members.length,
        );
        return members;
      } catch (err) {
        console.error(
          "[useCollectionSharing] Failed to get collection members:",
          err,
        );
        setError(`Failed to get collection members: ${err.message}`);
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [shareCollectionManager],
  );

  // Share collection with read-only access (convenience method)
  const shareCollectionReadOnly = useCallback(
    async (
      collectionId,
      recipientId,
      recipientEmail,
      shareDescendants = true,
      password = null,
    ) => {
      return shareCollection(
        collectionId,
        {
          recipient_id: recipientId,
          recipient_email: recipientEmail,
          permission_level: "read_only",
          share_with_descendants: shareDescendants,
        },
        password,
      );
    },
    [shareCollection],
  );

  // Share collection with read-write access (convenience method)
  const shareCollectionReadWrite = useCallback(
    async (
      collectionId,
      recipientId,
      recipientEmail,
      shareDescendants = true,
      password = null,
    ) => {
      return shareCollection(
        collectionId,
        {
          recipient_id: recipientId,
          recipient_email: recipientEmail,
          permission_level: "read_write",
          share_with_descendants: shareDescendants,
        },
        password,
      );
    },
    [shareCollection],
  );

  // Share collection with admin access (convenience method)
  const shareCollectionAdmin = useCallback(
    async (
      collectionId,
      recipientId,
      recipientEmail,
      shareDescendants = true,
      password = null,
    ) => {
      return shareCollection(
        collectionId,
        {
          recipient_id: recipientId,
          recipient_email: recipientEmail,
          permission_level: "admin",
          share_with_descendants: shareDescendants,
        },
        password,
      );
    },
    [shareCollection],
  );

  // Get shared collections by collection ID
  const getSharedCollectionsByCollectionId = useCallback(
    (collectionId) => {
      if (!shareCollectionManager) return [];

      return shareCollectionManager.getSharedCollectionsByCollectionId(
        collectionId,
      );
    },
    [shareCollectionManager],
  );

  // Get shared collections by recipient
  const getSharedCollectionsByRecipient = useCallback(
    (recipientId) => {
      if (!shareCollectionManager) return [];

      return shareCollectionManager.getSharedCollectionsByRecipient(
        recipientId,
      );
    },
    [shareCollectionManager],
  );

  // Search shared collections
  const searchSharedCollections = useCallback(
    (searchTerm) => {
      if (!shareCollectionManager) return [];

      try {
        return shareCollectionManager.searchSharedCollections(searchTerm);
      } catch (err) {
        console.error("[useCollectionSharing] Search failed:", err);
        return [];
      }
    },
    [shareCollectionManager],
  );

  // Get sharing history for specific collection
  const getCollectionSharingHistory = useCallback(
    (collectionId) => {
      if (!shareCollectionManager) return [];

      return shareCollectionManager.getSharingHistory(collectionId);
    },
    [shareCollectionManager],
  );

  // Remove all shares for a collection
  const removeAllSharesForCollection = useCallback(
    async (collectionId) => {
      if (!shareCollectionManager) {
        throw new Error("Share collection manager not available");
      }

      try {
        console.log(
          "[useCollectionSharing] Removing all shares for collection:",
          collectionId,
        );

        const removedCount =
          await shareCollectionManager.removeAllSharesForCollection(
            collectionId,
          );
        setSuccess(`Removed ${removedCount} shares for collection!`);
        loadSharedCollections(); // Reload collections
        loadSharingHistory(); // Reload history

        return removedCount;
      } catch (err) {
        console.error(
          "[useCollectionSharing] Failed to remove all shares for collection:",
          err,
        );
        setError(`Failed to remove all shares: ${err.message}`);
        throw err;
      }
    },
    [shareCollectionManager, loadSharedCollections, loadSharingHistory],
  );

  // Clear all shared collections
  const clearAllSharedCollections = useCallback(async () => {
    if (!shareCollectionManager) {
      throw new Error("Share collection manager not available");
    }

    try {
      console.log("[useCollectionSharing] Clearing all shared collections");

      await shareCollectionManager.clearAllSharedCollections();
      setSuccess("All shared collections cleared successfully!");
      loadSharedCollections(); // Reload collections
      loadSharingHistory(); // Reload history
    } catch (err) {
      console.error(
        "[useCollectionSharing] Failed to clear shared collections:",
        err,
      );
      setError(`Failed to clear shared collections: ${err.message}`);
      throw err;
    }
  }, [shareCollectionManager, loadSharedCollections, loadSharingHistory]);

  // Get user password from storage
  const getUserPassword = useCallback(async () => {
    if (!shareCollectionManager) return null;

    try {
      return await shareCollectionManager.getUserPassword();
    } catch (err) {
      console.error("[useCollectionSharing] Failed to get user password:", err);
      return null;
    }
  }, [shareCollectionManager]);

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
    loadSharedCollections();
    loadSharingHistory();
    loadManagerStatus();
  }, [loadSharedCollections, loadSharingHistory, loadManagerStatus]);

  // Load data on mount and when manager changes
  useEffect(() => {
    if (shareCollectionManager) {
      loadSharedCollections();
      loadSharingHistory();
      loadManagerStatus();
    }
  }, [
    shareCollectionManager,
    loadSharedCollections,
    loadSharingHistory,
    loadManagerStatus,
  ]);

  // Set up event listeners for collection sharing events
  useEffect(() => {
    if (!shareCollectionManager) return;

    const handleCollectionEvent = (eventType, eventData) => {
      console.log(
        "[useCollectionSharing] Collection event:",
        eventType,
        eventData,
      );

      // Reload data on certain events
      if (
        [
          "collection_shared",
          "member_removed",
          "all_shares_removed_for_collection",
          "all_shared_collections_cleared",
        ].includes(eventType)
      ) {
        loadSharedCollections();
        loadSharingHistory();
        loadManagerStatus();
      }
    };

    shareCollectionManager.addCollectionSharingListener(handleCollectionEvent);

    return () => {
      shareCollectionManager.removeCollectionSharingListener(
        handleCollectionEvent,
      );
    };
  }, [
    shareCollectionManager,
    loadSharedCollections,
    loadSharingHistory,
    loadManagerStatus,
  ]);

  return {
    // State
    isLoading,
    error,
    success,
    sharedCollections,
    collectionMembers,
    sharingHistory,
    managerStatus,

    // Core operations
    shareCollection,
    removeMember,
    getCollectionMembers,

    // Convenience methods
    shareCollectionReadOnly,
    shareCollectionReadWrite,
    shareCollectionAdmin,
    removeAllSharesForCollection,
    clearAllSharedCollections,

    // Utility operations
    getSharedCollectionsByCollectionId,
    getSharedCollectionsByRecipient,
    searchSharedCollections,
    getCollectionSharingHistory,
    getUserPassword,
    loadSharedCollections,
    loadSharingHistory,
    loadManagerStatus,

    // State management
    clearMessages,
    reset,

    // Status checks
    isAuthenticated: authManager?.isAuthenticated() || false,
    canShareCollections: managerStatus.canShareCollections || false,
    hasStoredPassword: !!managerStatus.hasPasswordService,

    // Permission levels
    PERMISSION_LEVELS: {
      READ_ONLY: "read_only",
      READ_WRITE: "read_write",
      ADMIN: "admin",
    },

    // Statistics
    totalSharedCollections: sharedCollections.length,
    totalSharingHistory: sharingHistory.length,
    sharedCollectionsByPermission: sharedCollections.reduce((acc, share) => {
      const permission = share.permission_level || "unknown";
      acc[permission] = (acc[permission] || 0) + 1;
      return acc;
    }, {}),
    sharedCollectionsByRecipient: sharedCollections.reduce((acc, share) => {
      const recipient = share.recipient_email || "unknown";
      acc[recipient] = (acc[recipient] || 0) + 1;
      return acc;
    }, {}),

    // Helper methods
    getLatestShareForCollection: (collectionId) => {
      return (
        sharingHistory
          .filter((h) => h.collection_id === collectionId)
          .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))[0] ||
        null
      );
    },

    getRecentShares: (hours = 24) => {
      const cutoff = new Date(Date.now() - hours * 60 * 60 * 1000);
      return sharingHistory.filter((h) => new Date(h.timestamp) > cutoff);
    },

    getCollectionMembersById: (collectionId) => {
      return collectionMembers[collectionId] || [];
    },

    hasCollectionMembers: (collectionId) => {
      return (collectionMembers[collectionId] || []).length > 0;
    },
  };
};

export default useCollectionSharing;

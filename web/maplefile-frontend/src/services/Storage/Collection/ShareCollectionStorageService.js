// File: monorepo/web/maplefile-frontend/src/services/Storage/Collection/ShareCollectionStorageService.js
// Share Collection Storage Service - Handles localStorage operations for collection sharing

class ShareCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      SHARED_COLLECTIONS: "mapleapps_shared_collections",
      SHARING_HISTORY: "mapleapps_collection_sharing_history",
      COLLECTION_MEMBERS: "mapleapps_collection_members",
      PENDING_SHARES: "mapleapps_pending_shares",
    };

    console.log("[ShareCollectionStorageService] Storage service initialized");
  }

  // === Shared Collection Storage ===

  // Store shared collection data
  storeSharedCollection(collectionId, shareInfo) {
    try {
      const existingShared = this.getSharedCollections();

      // Remove any existing share for same collection-recipient pair
      const filteredShared = existingShared.filter(
        (s) =>
          !(
            s.collection_id === collectionId &&
            s.recipient_id === shareInfo.recipient_id
          ),
      );

      // Add the new share
      const sharedData = {
        collection_id: collectionId,
        ...shareInfo,
        shared_at: new Date().toISOString(),
        locally_stored_at: new Date().toISOString(),
      };

      filteredShared.push(sharedData);

      localStorage.setItem(
        this.STORAGE_KEYS.SHARED_COLLECTIONS,
        JSON.stringify(filteredShared),
      );

      console.log(
        "[ShareCollectionStorageService] Shared collection stored:",
        collectionId,
        "->",
        shareInfo.recipient_email,
      );

      // Store in sharing history
      this.addToSharingHistory(collectionId, {
        action: "shared",
        recipient_id: shareInfo.recipient_id,
        recipient_email: shareInfo.recipient_email,
        permission_level: shareInfo.permission_level,
        share_with_descendants: shareInfo.share_with_descendants,
        timestamp: new Date().toISOString(),
      });
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to store shared collection:",
        error,
      );
    }
  }

  // Get all shared collections
  getSharedCollections() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.SHARED_COLLECTIONS);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to get shared collections:",
        error,
      );
      return [];
    }
  }

  // Get shared collections by collection ID
  getSharedCollectionsByCollectionId(collectionId) {
    const shared = this.getSharedCollections();
    return shared.filter((s) => s.collection_id === collectionId);
  }

  // Get shared collections by recipient
  getSharedCollectionsByRecipient(recipientId) {
    const shared = this.getSharedCollections();
    return shared.filter((s) => s.recipient_id === recipientId);
  }

  // === Collection Members Storage ===

  // Store collection members data (from API responses)
  storeCollectionMembers(collectionId, members) {
    try {
      const existingMembers = this.getAllCollectionMembers();

      // Update members for this collection
      existingMembers[collectionId] = {
        members: members,
        updated_at: new Date().toISOString(),
      };

      localStorage.setItem(
        this.STORAGE_KEYS.COLLECTION_MEMBERS,
        JSON.stringify(existingMembers),
      );

      console.log(
        "[ShareCollectionStorageService] Collection members stored:",
        collectionId,
        "->",
        members.length,
        "members",
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to store collection members:",
        error,
      );
    }
  }

  // Get collection members
  getCollectionMembers(collectionId) {
    try {
      const allMembers = this.getAllCollectionMembers();
      const collectionData = allMembers[collectionId];
      return collectionData ? collectionData.members : [];
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to get collection members:",
        error,
      );
      return [];
    }
  }

  // Get all collection members data
  getAllCollectionMembers() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.COLLECTION_MEMBERS);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to get all collection members:",
        error,
      );
      return {};
    }
  }

  // === Sharing History Management ===

  // Add entry to sharing history
  addToSharingHistory(collectionId, shareInfo) {
    try {
      const history = this.getSharingHistory();

      const historyEntry = {
        collection_id: collectionId,
        timestamp: new Date().toISOString(),
        ...shareInfo,
      };

      history.push(historyEntry);

      // Keep only last 200 shares
      const limitedHistory = history.slice(-200);

      localStorage.setItem(
        this.STORAGE_KEYS.SHARING_HISTORY,
        JSON.stringify(limitedHistory),
      );

      console.log(
        "[ShareCollectionStorageService] Sharing history entry added:",
        collectionId,
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to add sharing history:",
        error,
      );
    }
  }

  // Get sharing history
  getSharingHistory(collectionId = null) {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.SHARING_HISTORY);
      const history = stored ? JSON.parse(stored) : [];

      if (collectionId) {
        return history.filter((entry) => entry.collection_id === collectionId);
      }

      return history;
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to get sharing history:",
        error,
      );
      return [];
    }
  }

  // === Pending Shares Management ===

  // Store pending share (for offline support)
  storePendingShare(shareData) {
    try {
      const pending = this.getPendingShares();

      const pendingShare = {
        ...shareData,
        id: this.generateId(),
        created_at: new Date().toISOString(),
        status: "pending",
      };

      pending.push(pendingShare);

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_SHARES,
        JSON.stringify(pending),
      );

      console.log(
        "[ShareCollectionStorageService] Pending share stored:",
        pendingShare.id,
      );

      return pendingShare.id;
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to store pending share:",
        error,
      );
      return null;
    }
  }

  // Get pending shares
  getPendingShares() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.PENDING_SHARES);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to get pending shares:",
        error,
      );
      return [];
    }
  }

  // Mark pending share as completed
  markPendingShareCompleted(pendingId) {
    try {
      const pending = this.getPendingShares();
      const updated = pending.map((share) =>
        share.id === pendingId
          ? {
              ...share,
              status: "completed",
              completed_at: new Date().toISOString(),
            }
          : share,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_SHARES,
        JSON.stringify(updated),
      );

      console.log(
        "[ShareCollectionStorageService] Pending share marked completed:",
        pendingId,
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to mark pending share completed:",
        error,
      );
    }
  }

  // Remove pending share
  removePendingShare(pendingId) {
    try {
      const pending = this.getPendingShares();
      const filtered = pending.filter((share) => share.id !== pendingId);

      localStorage.setItem(
        this.STORAGE_KEYS.PENDING_SHARES,
        JSON.stringify(filtered),
      );

      console.log(
        "[ShareCollectionStorageService] Pending share removed:",
        pendingId,
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to remove pending share:",
        error,
      );
    }
  }

  // === Collection Statistics ===

  // Get sharing statistics
  getSharingStats() {
    const shared = this.getSharedCollections();
    const history = this.getSharingHistory();
    const pending = this.getPendingShares();

    const stats = {
      totalShared: shared.length,
      byPermission: {},
      byRecipient: {},
      recent: 0, // shared in last 24 hours
      totalHistoryEntries: history.length,
      pendingShares: pending.filter((p) => p.status === "pending").length,
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    shared.forEach((share) => {
      // Count by permission level
      const permission = share.permission_level || "unknown";
      stats.byPermission[permission] =
        (stats.byPermission[permission] || 0) + 1;

      // Count by recipient
      const recipient = share.recipient_email || "unknown";
      stats.byRecipient[recipient] = (stats.byRecipient[recipient] || 0) + 1;

      // Count recent shares
      const sharedAt = new Date(
        share.shared_at || share.locally_stored_at,
      ).getTime();
      if (sharedAt > oneDayAgo) {
        stats.recent++;
      }
    });

    return stats;
  }

  // === Collection Search ===

  // Search shared collections
  searchSharedCollections(searchTerm, sharedCollections = null) {
    if (!searchTerm) return this.getSharedCollections();

    const shares = sharedCollections || this.getSharedCollections();
    const term = searchTerm.toLowerCase();

    return shares.filter((share) => {
      // Search in recipient email
      if (
        share.recipient_email &&
        share.recipient_email.toLowerCase().includes(term)
      ) {
        return true;
      }

      // Search in permission level
      if (
        share.permission_level &&
        share.permission_level.toLowerCase().includes(term)
      ) {
        return true;
      }

      // Search in collection ID (partial)
      if (
        share.collection_id &&
        share.collection_id.toLowerCase().includes(term)
      ) {
        return true;
      }

      return false;
    });
  }

  // === Data Management ===

  // Remove shared collection
  removeSharedCollection(collectionId, recipientId) {
    try {
      const shared = this.getSharedCollections();
      const filtered = shared.filter(
        (s) =>
          !(s.collection_id === collectionId && s.recipient_id === recipientId),
      );

      localStorage.setItem(
        this.STORAGE_KEYS.SHARED_COLLECTIONS,
        JSON.stringify(filtered),
      );

      console.log(
        "[ShareCollectionStorageService] Shared collection removed:",
        collectionId,
        "->",
        recipientId,
      );

      // Add to history
      this.addToSharingHistory(collectionId, {
        action: "unshared",
        recipient_id: recipientId,
        timestamp: new Date().toISOString(),
      });

      return true;
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to remove shared collection:",
        error,
      );
      return false;
    }
  }

  // Remove all shares for a collection
  removeAllSharedForCollection(collectionId) {
    try {
      const shared = this.getSharedCollections();
      const toRemove = shared.filter((s) => s.collection_id === collectionId);
      const filtered = shared.filter((s) => s.collection_id !== collectionId);

      localStorage.setItem(
        this.STORAGE_KEYS.SHARED_COLLECTIONS,
        JSON.stringify(filtered),
      );

      // Also clear members for this collection
      const allMembers = this.getAllCollectionMembers();
      delete allMembers[collectionId];
      localStorage.setItem(
        this.STORAGE_KEYS.COLLECTION_MEMBERS,
        JSON.stringify(allMembers),
      );

      console.log(
        "[ShareCollectionStorageService] All shares removed for collection:",
        collectionId,
        "->",
        toRemove.length,
        "shares",
      );

      return toRemove.length;
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to remove all shares for collection:",
        error,
      );
      return 0;
    }
  }

  // === Cache Management ===

  // Clear all shared collections
  clearAllSharedCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.SHARED_COLLECTIONS);
      console.log(
        "[ShareCollectionStorageService] All shared collections cleared",
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to clear shared collections:",
        error,
      );
    }
  }

  // Clear sharing history
  clearSharingHistory() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.SHARING_HISTORY);
      console.log("[ShareCollectionStorageService] Sharing history cleared");
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to clear sharing history:",
        error,
      );
    }
  }

  // Clear all collection members
  clearAllCollectionMembers() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.COLLECTION_MEMBERS);
      console.log(
        "[ShareCollectionStorageService] All collection members cleared",
      );
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to clear collection members:",
        error,
      );
    }
  }

  // Clear pending shares
  clearPendingShares() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.PENDING_SHARES);
      console.log("[ShareCollectionStorageService] Pending shares cleared");
    } catch (error) {
      console.error(
        "[ShareCollectionStorageService] Failed to clear pending shares:",
        error,
      );
    }
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const shared = this.getSharedCollections();
    const stats = this.getSharingStats();
    const allMembers = this.getAllCollectionMembers();
    const pending = this.getPendingShares();

    return {
      sharedCollectionsCount: shared.length,
      collectionMembersCount: Object.keys(allMembers).length,
      pendingSharesCount: pending.length,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasSharedCollections: shared.length > 0,
      hasPendingShares: pending.length > 0,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "ShareCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      recentHistory: this.getSharingHistory().slice(-5), // Last 5 shares
      pendingShares: this.getPendingShares(),
    };
  }

  // === Helper Methods ===

  // Generate simple ID for pending shares
  generateId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  // Validate share data structure
  validateShareStructure(shareData) {
    const requiredFields = [
      "collection_id",
      "recipient_id",
      "recipient_email",
      "permission_level",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (!shareData[field]) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    const validPermissions = ["read_only", "read_write", "admin"];
    if (
      shareData.permission_level &&
      !validPermissions.includes(shareData.permission_level)
    ) {
      errors.push(`Invalid permission level: ${shareData.permission_level}`);
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }
}

export default ShareCollectionStorageService;

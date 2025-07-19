// File: monorepo/web/maplefile-frontend/src/services/Storage/Collection/DeleteCollectionStorageService.js
// Delete Collection Storage Service - Handles localStorage operations for collection deletions

class DeleteCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      DELETED_COLLECTIONS: "mapleapps_deleted_collections",
      DELETION_HISTORY: "mapleapps_collection_deletion_history",
    };

    // In-memory cache for collection keys (NEVER stored in localStorage)
    this._collectionKeyCache = new Map();

    console.log("[DeleteCollectionStorageService] Storage service initialized");
  }

  // === Collection Deletion Storage ===

  // Store deleted collection
  storeDeletedCollection(collection) {
    try {
      const existingCollections = this.getDeletedCollections();

      // Remove any existing collection with same ID
      const filteredCollections = existingCollections.filter(
        (c) => c.id !== collection.id,
      );

      // Add the deleted collection
      filteredCollections.push({
        ...collection,
        deleted_at: new Date().toISOString(),
        locally_deleted_at: new Date().toISOString(),
        state: "deleted",
      });

      localStorage.setItem(
        this.STORAGE_KEYS.DELETED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      console.log(
        "[DeleteCollectionStorageService] Deleted collection stored:",
        collection.id,
      );

      // Store deletion in history
      this.addToDeletionHistory(collection.id, {
        action: "deleted",
        timestamp: new Date().toISOString(),
        collection_type: collection.collection_type,
        collection_name: collection.name || "[Encrypted]",
      });
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to store deleted collection:",
        error,
      );
    }
  }

  // Get all deleted collections
  getDeletedCollections() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.DELETED_COLLECTIONS,
      );
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to get deleted collections:",
        error,
      );
      return [];
    }
  }

  // Get deleted collection by ID
  getDeletedCollectionById(collectionId) {
    const collections = this.getDeletedCollections();
    return collections.find((c) => c.id === collectionId) || null;
  }

  // === Deletion History Management ===

  // Add entry to deletion history
  addToDeletionHistory(collectionId, deletionInfo) {
    try {
      const history = this.getDeletionHistory();

      const historyEntry = {
        collectionId,
        timestamp: new Date().toISOString(),
        ...deletionInfo,
      };

      history.push(historyEntry);

      // Keep only last 100 deletions
      const limitedHistory = history.slice(-100);

      localStorage.setItem(
        this.STORAGE_KEYS.DELETION_HISTORY,
        JSON.stringify(limitedHistory),
      );

      console.log(
        "[DeleteCollectionStorageService] Deletion history entry added:",
        collectionId,
      );
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to add deletion history:",
        error,
      );
    }
  }

  // Get deletion history
  getDeletionHistory(collectionId = null) {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DELETION_HISTORY);
      const history = stored ? JSON.parse(stored) : [];

      if (collectionId) {
        return history.filter((entry) => entry.collectionId === collectionId);
      }

      return history;
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to get deletion history:",
        error,
      );
      return [];
    }
  }

  // === Collection Key Management (In-Memory Only) ===

  // Store collection key in memory (NEVER in localStorage)
  storeCollectionKey(collectionId, collectionKey) {
    this._collectionKeyCache.set(collectionId, {
      key: collectionKey,
      stored_at: Date.now(),
    });
    console.log(
      `[DeleteCollectionStorageService] Collection key cached in memory for: ${collectionId}`,
    );
  }

  // Get collection key from memory
  getCollectionKey(collectionId) {
    const cached = this._collectionKeyCache.get(collectionId);
    if (cached) {
      console.log(
        `[DeleteCollectionStorageService] Retrieved collection key from memory for: ${collectionId}`,
      );
      return cached.key;
    }
    return null;
  }

  // Check if collection key exists in memory
  hasCollectionKey(collectionId) {
    return this._collectionKeyCache.has(collectionId);
  }

  // Clear collection key from memory
  clearCollectionKey(collectionId) {
    const deleted = this._collectionKeyCache.delete(collectionId);
    if (deleted) {
      console.log(
        `[DeleteCollectionStorageService] Collection key cleared from memory for: ${collectionId}`,
      );
    }
    return deleted;
  }

  // === Collection Statistics ===

  // Get deletion statistics
  getDeletionStats() {
    const collections = this.getDeletedCollections();
    const history = this.getDeletionHistory();

    const stats = {
      totalDeleted: collections.length,
      byType: {},
      recent: 0, // deleted in last 24 hours
      totalHistoryEntries: history.length,
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    collections.forEach((collection) => {
      // Count by type
      const type = collection.collection_type || "unknown";
      stats.byType[type] = (stats.byType[type] || 0) + 1;

      // Count recent deletions
      const deletedAt = new Date(
        collection.locally_deleted_at || collection.deleted_at,
      ).getTime();
      if (deletedAt > oneDayAgo) {
        stats.recent++;
      }
    });

    return stats;
  }

  // === Collection Search ===

  // Search deleted collections by name (requires decryption)
  searchDeletedCollections(searchTerm, decryptedCollections = null) {
    if (!searchTerm) return this.getDeletedCollections();

    const collections = decryptedCollections || this.getDeletedCollections();
    const term = searchTerm.toLowerCase();

    return collections.filter((collection) => {
      // Search in decrypted name if available
      if (collection.name && collection.name.toLowerCase().includes(term)) {
        return true;
      }

      // Search in collection type
      if (
        collection.collection_type &&
        collection.collection_type.toLowerCase().includes(term)
      ) {
        return true;
      }

      // Search in ID (partial)
      if (collection.id && collection.id.toLowerCase().includes(term)) {
        return true;
      }

      return false;
    });
  }

  // === Data Management ===

  // Restore collection (remove from deleted list)
  restoreDeletedCollection(collectionId) {
    try {
      const collections = this.getDeletedCollections();
      const filteredCollections = collections.filter(
        (c) => c.id !== collectionId,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.DELETED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      console.log(
        "[DeleteCollectionStorageService] Deleted collection restored:",
        collectionId,
      );

      // Add to history
      this.addToDeletionHistory(collectionId, {
        action: "restored",
        timestamp: new Date().toISOString(),
      });

      return true;
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to restore deleted collection:",
        error,
      );
      return false;
    }
  }

  // Permanently remove deleted collection from storage
  permanentlyRemoveCollection(collectionId) {
    try {
      const collections = this.getDeletedCollections();
      const filteredCollections = collections.filter(
        (c) => c.id !== collectionId,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.DELETED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      // Also clear the collection key from memory
      this.clearCollectionKey(collectionId);

      console.log(
        "[DeleteCollectionStorageService] Deleted collection permanently removed:",
        collectionId,
      );

      // Add to history
      this.addToDeletionHistory(collectionId, {
        action: "permanently_removed_from_local_storage",
        timestamp: new Date().toISOString(),
      });

      return true;
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to permanently remove deleted collection:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Clear all deleted collections
  clearAllDeletedCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DELETED_COLLECTIONS);
      this._collectionKeyCache.clear();
      console.log(
        "[DeleteCollectionStorageService] All deleted collections cleared",
      );
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to clear deleted collections:",
        error,
      );
    }
  }

  // Clear deletion history
  clearDeletionHistory() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DELETION_HISTORY);
      console.log("[DeleteCollectionStorageService] Deletion history cleared");
    } catch (error) {
      console.error(
        "[DeleteCollectionStorageService] Failed to clear deletion history:",
        error,
      );
    }
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const collections = this.getDeletedCollections();
    const stats = this.getDeletionStats();

    return {
      deletedCollectionsCount: collections.length,
      collectionKeysInMemory: this._collectionKeyCache.size,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasDeletedCollections: collections.length > 0,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "DeleteCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      cacheKeys: Array.from(this._collectionKeyCache.keys()),
      recentHistory: this.getDeletionHistory().slice(-5), // Last 5 deletions
    };
  }

  // === Collection Validation ===

  // Validate collection data structure
  validateCollectionStructure(collection) {
    const requiredFields = ["id", "collection_type"];
    const errors = [];

    requiredFields.forEach((field) => {
      if (!collection[field]) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    return {
      isValid: errors.length === 0,
      errors,
    };
  }
}

export default DeleteCollectionStorageService;

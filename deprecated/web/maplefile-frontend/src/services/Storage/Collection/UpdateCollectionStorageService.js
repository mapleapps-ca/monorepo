// File: monorepo/web/maplefile-frontend/src/services/Storage/Collection/UpdateCollectionStorageService.js
// Update Collection Storage Service - Handles localStorage operations for collection updates

class UpdateCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      UPDATED_COLLECTIONS: "mapleapps_updated_collections",
      UPDATE_HISTORY: "mapleapps_collection_update_history",
    };

    // In-memory cache for collection keys (NEVER stored in localStorage)
    this._collectionKeyCache = new Map();

    console.log("[UpdateCollectionStorageService] Storage service initialized");
  }

  // === Collection Update Storage ===

  // Store updated collection
  storeUpdatedCollection(collection) {
    try {
      const existingCollections = this.getUpdatedCollections();

      // Remove any existing collection with same ID
      const filteredCollections = existingCollections.filter(
        (c) => c.id !== collection.id,
      );

      // Add the updated collection
      filteredCollections.push({
        ...collection,
        updated_at: new Date().toISOString(),
        locally_updated_at: new Date().toISOString(),
      });

      localStorage.setItem(
        this.STORAGE_KEYS.UPDATED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      console.log(
        "[UpdateCollectionStorageService] Updated collection stored:",
        collection.id,
      );

      // Store update in history
      this.addToUpdateHistory(collection.id, {
        action: "updated",
        timestamp: new Date().toISOString(),
        version: collection.version,
        collection_type: collection.collection_type,
        hasNameChange: !!collection.name,
      });
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to store updated collection:",
        error,
      );
    }
  }

  // Get all updated collections
  getUpdatedCollections() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.UPDATED_COLLECTIONS,
      );
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to get updated collections:",
        error,
      );
      return [];
    }
  }

  // Get updated collection by ID
  getUpdatedCollectionById(collectionId) {
    const collections = this.getUpdatedCollections();
    return collections.find((c) => c.id === collectionId) || null;
  }

  // === Update History Management ===

  // Add entry to update history
  addToUpdateHistory(collectionId, updateInfo) {
    try {
      const history = this.getUpdateHistory();

      const historyEntry = {
        collectionId,
        timestamp: new Date().toISOString(),
        ...updateInfo,
      };

      history.push(historyEntry);

      // Keep only last 100 updates
      const limitedHistory = history.slice(-100);

      localStorage.setItem(
        this.STORAGE_KEYS.UPDATE_HISTORY,
        JSON.stringify(limitedHistory),
      );

      console.log(
        "[UpdateCollectionStorageService] Update history entry added:",
        collectionId,
      );
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to add update history:",
        error,
      );
    }
  }

  // Get update history
  getUpdateHistory(collectionId = null) {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.UPDATE_HISTORY);
      const history = stored ? JSON.parse(stored) : [];

      if (collectionId) {
        return history.filter((entry) => entry.collectionId === collectionId);
      }

      return history;
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to get update history:",
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
      `[UpdateCollectionStorageService] Collection key cached in memory for: ${collectionId}`,
    );
  }

  // Get collection key from memory
  getCollectionKey(collectionId) {
    const cached = this._collectionKeyCache.get(collectionId);
    if (cached) {
      console.log(
        `[UpdateCollectionStorageService] Retrieved collection key from memory for: ${collectionId}`,
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
        `[UpdateCollectionStorageService] Collection key cleared from memory for: ${collectionId}`,
      );
    }
    return deleted;
  }

  // === Collection Statistics ===

  // Get update statistics
  getUpdateStats() {
    const collections = this.getUpdatedCollections();
    const history = this.getUpdateHistory();

    const stats = {
      totalUpdated: collections.length,
      byType: {},
      recent: 0, // updated in last 24 hours
      totalHistoryEntries: history.length,
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    collections.forEach((collection) => {
      // Count by type
      const type = collection.collection_type || "unknown";
      stats.byType[type] = (stats.byType[type] || 0) + 1;

      // Count recent updates
      const updatedAt = new Date(
        collection.locally_updated_at || collection.updated_at,
      ).getTime();
      if (updatedAt > oneDayAgo) {
        stats.recent++;
      }
    });

    return stats;
  }

  // === Collection Search ===

  // Search updated collections by name (requires decryption)
  searchUpdatedCollections(searchTerm, decryptedCollections = null) {
    if (!searchTerm) return this.getUpdatedCollections();

    const collections = decryptedCollections || this.getUpdatedCollections();
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

  // Remove updated collection from storage
  removeUpdatedCollection(collectionId) {
    try {
      const collections = this.getUpdatedCollections();
      const filteredCollections = collections.filter(
        (c) => c.id !== collectionId,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.UPDATED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      // Also clear the collection key from memory
      this.clearCollectionKey(collectionId);

      console.log(
        "[UpdateCollectionStorageService] Updated collection removed:",
        collectionId,
      );

      // Add to history
      this.addToUpdateHistory(collectionId, {
        action: "removed_from_local_storage",
        timestamp: new Date().toISOString(),
      });

      return true;
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to remove updated collection:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Clear all updated collections
  clearAllUpdatedCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.UPDATED_COLLECTIONS);
      this._collectionKeyCache.clear();
      console.log(
        "[UpdateCollectionStorageService] All updated collections cleared",
      );
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to clear updated collections:",
        error,
      );
    }
  }

  // Clear update history
  clearUpdateHistory() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.UPDATE_HISTORY);
      console.log("[UpdateCollectionStorageService] Update history cleared");
    } catch (error) {
      console.error(
        "[UpdateCollectionStorageService] Failed to clear update history:",
        error,
      );
    }
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const collections = this.getUpdatedCollections();
    const stats = this.getUpdateStats();

    return {
      updatedCollectionsCount: collections.length,
      collectionKeysInMemory: this._collectionKeyCache.size,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasUpdatedCollections: collections.length > 0,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "UpdateCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      cacheKeys: Array.from(this._collectionKeyCache.keys()),
      recentHistory: this.getUpdateHistory().slice(-5), // Last 5 updates
    };
  }

  // === Collection Validation ===

  // Validate collection data structure
  validateCollectionStructure(collection) {
    const requiredFields = ["id", "version"];
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

export default UpdateCollectionStorageService;

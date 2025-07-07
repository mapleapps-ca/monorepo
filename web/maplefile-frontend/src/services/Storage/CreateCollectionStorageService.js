// File: monorepo/web/maplefile-frontend/src/services/Storage/CreateCollectionStorageService.js
// Create Collection Storage Service - Handles localStorage operations for collections

class CreateCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      COLLECTIONS: "mapleapps_collections",
      COLLECTION_KEYS: "mapleapps_collection_keys", // In-memory only
      CREATED_COLLECTIONS: "mapleapps_created_collections",
    };

    // In-memory cache for collection keys (NEVER stored in localStorage)
    this._collectionKeyCache = new Map();

    console.log("[CreateCollectionStorageService] Storage service initialized");
  }

  // === Collection Storage Operations ===

  // Store created collection
  storeCreatedCollection(collection) {
    try {
      const existingCollections = this.getCreatedCollections();

      // Remove any existing collection with same ID
      const filteredCollections = existingCollections.filter(
        (c) => c.id !== collection.id,
      );

      // Add the new collection
      filteredCollections.push({
        ...collection,
        stored_at: new Date().toISOString(),
      });

      localStorage.setItem(
        this.STORAGE_KEYS.CREATED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );
      console.log(
        "[CreateCollectionStorageService] Collection stored:",
        collection.id,
      );
    } catch (error) {
      console.error(
        "[CreateCollectionStorageService] Failed to store collection:",
        error,
      );
    }
  }

  // Get all created collections
  getCreatedCollections() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.CREATED_COLLECTIONS,
      );
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error(
        "[CreateCollectionStorageService] Failed to get collections:",
        error,
      );
      return [];
    }
  }

  // Get collection by ID
  getCollectionById(collectionId) {
    const collections = this.getCreatedCollections();
    return collections.find((c) => c.id === collectionId) || null;
  }

  // === Collection Key Management (In-Memory Only) ===

  // Store collection key in memory (NEVER in localStorage)
  storeCollectionKey(collectionId, collectionKey) {
    this._collectionKeyCache.set(collectionId, {
      key: collectionKey,
      stored_at: Date.now(),
    });
    console.log(
      `[CreateCollectionStorageService] Collection key cached in memory for: ${collectionId}`,
    );
  }

  // Get collection key from memory
  getCollectionKey(collectionId) {
    const cached = this._collectionKeyCache.get(collectionId);
    if (cached) {
      console.log(
        `[CreateCollectionStorageService] Retrieved collection key from memory for: ${collectionId}`,
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
        `[CreateCollectionStorageService] Collection key cleared from memory for: ${collectionId}`,
      );
    }
    return deleted;
  }

  // Clear all collection keys from memory
  clearAllCollectionKeys() {
    const count = this._collectionKeyCache.size;
    this._collectionKeyCache.clear();
    console.log(
      `[CreateCollectionStorageService] ${count} collection keys cleared from memory`,
    );
  }

  // === Collection Statistics ===

  // Get collection creation statistics
  getCollectionStats() {
    const collections = this.getCreatedCollections();
    const stats = {
      total: collections.length,
      byType: {},
      recent: 0, // created in last 24 hours
    };

    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;

    collections.forEach((collection) => {
      // Count by type
      const type = collection.collection_type || "unknown";
      stats.byType[type] = (stats.byType[type] || 0) + 1;

      // Count recent
      const createdAt = new Date(
        collection.created_at || collection.stored_at,
      ).getTime();
      if (createdAt > oneDayAgo) {
        stats.recent++;
      }
    });

    return stats;
  }

  // === Collection Search ===

  // Search collections by name (requires decryption)
  searchCollections(searchTerm, decryptedCollections = null) {
    if (!searchTerm) return this.getCreatedCollections();

    const collections = decryptedCollections || this.getCreatedCollections();
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

  // Update collection data
  updateCollection(collectionId, updates) {
    try {
      const collections = this.getCreatedCollections();
      const index = collections.findIndex((c) => c.id === collectionId);

      if (index !== -1) {
        collections[index] = {
          ...collections[index],
          ...updates,
          updated_at: new Date().toISOString(),
        };

        localStorage.setItem(
          this.STORAGE_KEYS.CREATED_COLLECTIONS,
          JSON.stringify(collections),
        );
        console.log(
          "[CreateCollectionStorageService] Collection updated:",
          collectionId,
        );
        return collections[index];
      }

      throw new Error(`Collection not found: ${collectionId}`);
    } catch (error) {
      console.error(
        "[CreateCollectionStorageService] Failed to update collection:",
        error,
      );
      throw error;
    }
  }

  // Remove collection from storage
  removeCollection(collectionId) {
    try {
      const collections = this.getCreatedCollections();
      const filteredCollections = collections.filter(
        (c) => c.id !== collectionId,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.CREATED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      // Also clear the collection key from memory
      this.clearCollectionKey(collectionId);

      console.log(
        "[CreateCollectionStorageService] Collection removed:",
        collectionId,
      );
      return true;
    } catch (error) {
      console.error(
        "[CreateCollectionStorageService] Failed to remove collection:",
        error,
      );
      return false;
    }
  }

  // === Cache Management ===

  // Clear all stored collections
  clearAllCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.CREATED_COLLECTIONS);
      this.clearAllCollectionKeys();
      console.log("[CreateCollectionStorageService] All collections cleared");
    } catch (error) {
      console.error(
        "[CreateCollectionStorageService] Failed to clear collections:",
        error,
      );
    }
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const collections = this.getCreatedCollections();
    const stats = this.getCollectionStats();

    return {
      collectionsCount: collections.length,
      collectionKeysInMemory: this._collectionKeyCache.size,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasStoredCollections: collections.length > 0,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CreateCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      cacheKeys: Array.from(this._collectionKeyCache.keys()),
    };
  }

  // === Collection Validation ===

  // Validate collection data structure
  validateCollectionStructure(collection) {
    const requiredFields = ["id", "collection_type", "created_at"];
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

export default CreateCollectionStorageService;

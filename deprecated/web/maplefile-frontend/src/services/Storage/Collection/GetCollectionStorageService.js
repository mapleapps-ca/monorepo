// File: monorepo/web/maplefile-frontend/src/services/Storage/Collection/GetCollectionStorageService.js
// Get Collection Storage Service - Handles localStorage operations for retrieved collections

class GetCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      RETRIEVED_COLLECTIONS: "mapleapps_retrieved_collections",
      COLLECTION_CACHE_METADATA: "mapleapps_collection_cache_metadata",
    };

    // Cache expiry time (30 minutes)
    this.CACHE_EXPIRY_MS = 30 * 60 * 1000;

    // In-memory cache for collection keys (NEVER stored in localStorage)
    this._collectionKeyCache = new Map();

    console.log("[GetCollectionStorageService] Storage service initialized");
  }

  // === Collection Caching Operations ===

  // Store retrieved collection in cache
  storeRetrievedCollection(collection) {
    try {
      const existingCollections = this.getRetrievedCollections();

      // Remove any existing collection with same ID
      const filteredCollections = existingCollections.filter(
        (c) => c.id !== collection.id,
      );

      // Add the new collection with cache metadata
      const cachedCollection = {
        ...collection,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
      };

      filteredCollections.push(cachedCollection);

      localStorage.setItem(
        this.STORAGE_KEYS.RETRIEVED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      console.log(
        "[GetCollectionStorageService] Collection cached:",
        collection.id,
      );

      // Update cache metadata
      this.updateCacheMetadata();
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to cache collection:",
        error,
      );
    }
  }

  // Get all retrieved collections from cache
  getRetrievedCollections(includeExpired = false) {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.RETRIEVED_COLLECTIONS,
      );
      const collections = stored ? JSON.parse(stored) : [];

      if (!includeExpired) {
        // Filter out expired collections
        const now = new Date();
        return collections.filter((collection) => {
          if (!collection.cache_expiry) return true; // Keep collections without expiry
          return new Date(collection.cache_expiry) > now;
        });
      }

      return collections;
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to get cached collections:",
        error,
      );
      return [];
    }
  }

  // Get collection by ID from cache
  getCachedCollection(collectionId) {
    const collections = this.getRetrievedCollections();
    const collection = collections.find((c) => c.id === collectionId);

    if (collection) {
      console.log(
        `[GetCollectionStorageService] Collection found in cache: ${collectionId}`,
      );
      return collection;
    }

    console.log(
      `[GetCollectionStorageService] Collection not found in cache: ${collectionId}`,
    );
    return null;
  }

  // Check if collection is cached and not expired
  isCollectionCached(collectionId) {
    const collection = this.getCachedCollection(collectionId);
    return !!collection;
  }

  // Get collection cache status
  getCollectionCacheStatus(collectionId) {
    const collections = this.getRetrievedCollections(true); // Include expired
    const collection = collections.find((c) => c.id === collectionId);

    if (!collection) {
      return {
        cached: false,
        expired: false,
        cachedAt: null,
        expiresAt: null,
      };
    }

    const now = new Date();
    const expiryDate = new Date(collection.cache_expiry);
    const expired = expiryDate <= now;

    return {
      cached: true,
      expired: expired,
      cachedAt: collection.cached_at,
      expiresAt: collection.cache_expiry,
      timeUntilExpiry: expired ? 0 : expiryDate.getTime() - now.getTime(),
    };
  }

  // === Collection Key Management (In-Memory Only) ===

  // Store collection key in memory (NEVER in localStorage)
  storeCollectionKey(collectionId, collectionKey) {
    this._collectionKeyCache.set(collectionId, {
      key: collectionKey,
      stored_at: Date.now(),
    });
    console.log(
      `[GetCollectionStorageService] Collection key cached in memory for: ${collectionId}`,
    );
  }

  // Get collection key from memory
  getCollectionKey(collectionId) {
    const cached = this._collectionKeyCache.get(collectionId);
    if (cached) {
      console.log(
        `[GetCollectionStorageService] Retrieved collection key from memory for: ${collectionId}`,
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
        `[GetCollectionStorageService] Collection key cleared from memory for: ${collectionId}`,
      );
    }
    return deleted;
  }

  // === Cache Management ===

  // Clear expired collections from cache
  clearExpiredCollections() {
    try {
      const allCollections = this.getRetrievedCollections(true); // Include expired
      const validCollections = this.getRetrievedCollections(false); // Exclude expired
      const expiredCount = allCollections.length - validCollections.length;

      if (expiredCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.RETRIEVED_COLLECTIONS,
          JSON.stringify(validCollections),
        );

        console.log(
          `[GetCollectionStorageService] Cleared ${expiredCount} expired collections`,
        );
        this.updateCacheMetadata();
      }

      return expiredCount;
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to clear expired collections:",
        error,
      );
      return 0;
    }
  }

  // Remove specific collection from cache
  removeFromCache(collectionId) {
    try {
      const collections = this.getRetrievedCollections(true); // Include expired
      const filteredCollections = collections.filter(
        (c) => c.id !== collectionId,
      );

      localStorage.setItem(
        this.STORAGE_KEYS.RETRIEVED_COLLECTIONS,
        JSON.stringify(filteredCollections),
      );

      // Also clear the collection key from memory
      this.clearCollectionKey(collectionId);

      console.log(
        "[GetCollectionStorageService] Collection removed from cache:",
        collectionId,
      );
      this.updateCacheMetadata();
      return true;
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to remove collection from cache:",
        error,
      );
      return false;
    }
  }

  // Clear all cached collections
  clearAllCachedCollections() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.RETRIEVED_COLLECTIONS);
      localStorage.removeItem(this.STORAGE_KEYS.COLLECTION_CACHE_METADATA);

      // Clear all collection keys from memory
      this._collectionKeyCache.clear();

      console.log(
        "[GetCollectionStorageService] All cached collections cleared",
      );
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to clear cached collections:",
        error,
      );
    }
  }

  // === Cache Metadata ===

  // Update cache metadata
  updateCacheMetadata() {
    try {
      const collections = this.getRetrievedCollections(true); // Include expired
      const validCollections = this.getRetrievedCollections(false); // Exclude expired

      const metadata = {
        totalCached: collections.length,
        validCached: validCollections.length,
        expiredCached: collections.length - validCollections.length,
        lastUpdated: new Date().toISOString(),
        cacheExpiryMs: this.CACHE_EXPIRY_MS,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.COLLECTION_CACHE_METADATA,
        JSON.stringify(metadata),
      );
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to update cache metadata:",
        error,
      );
    }
  }

  // Get cache metadata
  getCacheMetadata() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.COLLECTION_CACHE_METADATA,
      );
      return stored ? JSON.parse(stored) : null;
    } catch (error) {
      console.error(
        "[GetCollectionStorageService] Failed to get cache metadata:",
        error,
      );
      return null;
    }
  }

  // === Collection Search ===

  // Search cached collections by name (requires decryption)
  searchCachedCollections(searchTerm, decryptedCollections = null) {
    if (!searchTerm) return this.getRetrievedCollections();

    const collections = decryptedCollections || this.getRetrievedCollections();
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

  // === Collection Statistics ===

  // Get cache statistics
  getCacheStats() {
    const allCollections = this.getRetrievedCollections(true);
    const validCollections = this.getRetrievedCollections(false);
    const metadata = this.getCacheMetadata();

    const stats = {
      total: allCollections.length,
      valid: validCollections.length,
      expired: allCollections.length - validCollections.length,
      byType: {},
      collectionKeysInMemory: this._collectionKeyCache.size,
      cacheExpiryMinutes: this.CACHE_EXPIRY_MS / (60 * 1000),
      lastUpdated: metadata?.lastUpdated || null,
    };

    validCollections.forEach((collection) => {
      const type = collection.collection_type || "unknown";
      stats.byType[type] = (stats.byType[type] || 0) + 1;
    });

    return stats;
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const stats = this.getCacheStats();

    return {
      cachedCollectionsCount: stats.valid,
      expiredCollectionsCount: stats.expired,
      collectionKeysInMemory: this._collectionKeyCache.size,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasCachedCollections: stats.valid > 0,
      cacheExpiryMs: this.CACHE_EXPIRY_MS,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "GetCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      cacheKeys: Array.from(this._collectionKeyCache.keys()),
      cacheMetadata: this.getCacheMetadata(),
    };
  }

  // === Collection Validation ===

  // Validate collection data structure
  validateCollectionStructure(collection) {
    const requiredFields = [
      "id",
      "collection_type",
      "encrypted_name",
      "created_at",
    ];
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

export default GetCollectionStorageService;

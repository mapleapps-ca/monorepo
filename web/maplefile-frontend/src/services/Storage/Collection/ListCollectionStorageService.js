// File: monorepo/web/maplefile-frontend/src/services/Storage/Collection/ListCollectionStorageService.js
// List Collection Storage Service - Handles localStorage operations for collection lists

class ListCollectionStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      LISTED_COLLECTIONS: "mapleapps_listed_collections",
      SHARED_COLLECTIONS: "mapleapps_shared_collections", // NEW KEY
      LIST_METADATA: "mapleapps_list_metadata",
      FILTERED_COLLECTIONS: "mapleapps_filtered_collections",
      ROOT_COLLECTIONS: "mapleapps_root_collections",
      COLLECTIONS_BY_PARENT: "mapleapps_collections_by_parent",
    };

    // Cache expiry time (15 minutes for list data)
    this.CACHE_EXPIRY_MS = 15 * 60 * 1000;

    console.log("[ListCollectionStorageService] Storage service initialized");
  }

  // === User Collections Operations ===

  // Store user's collections list
  storeListedCollections(collections) {
    try {
      const collectionData = {
        collections: collections,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
        total_count: collections.length,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.LISTED_COLLECTIONS,
        JSON.stringify(collectionData),
      );

      console.log(
        "[ListCollectionStorageService] Collections list cached:",
        collections.length,
      );

      this.updateListMetadata("listed_collections", collections.length);
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to store collections list:",
        error,
      );
      return false;
    }
  }

  // Get user's collections list
  getListedCollections() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.LISTED_COLLECTIONS);
      if (!stored) return { collections: [], isExpired: false };

      const data = JSON.parse(stored);
      const now = new Date();
      const isExpired = new Date(data.cache_expiry) <= now;

      if (isExpired) {
        console.log(
          "[ListCollectionStorageService] Collections list cache expired",
        );
        return { collections: [], isExpired: true };
      }

      console.log(
        "[ListCollectionStorageService] Retrieved collections list from cache:",
        data.collections.length,
      );

      return { collections: data.collections, isExpired: false };
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get collections list:",
        error,
      );
      return { collections: [], isExpired: false };
    }
  }

  // === Shared Collections Operations - NEW METHODS ===

  // Store shared collections list
  storeSharedCollections(collections) {
    try {
      const collectionData = {
        collections: collections,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
        total_count: collections.length,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.SHARED_COLLECTIONS,
        JSON.stringify(collectionData),
      );

      console.log(
        "[ListCollectionStorageService] Shared collections list cached:",
        collections.length,
      );

      this.updateListMetadata("shared_collections", collections.length);
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to store shared collections list:",
        error,
      );
      return false;
    }
  }

  // Get shared collections list
  getSharedCollections() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.SHARED_COLLECTIONS);
      if (!stored) return { collections: [], isExpired: false };

      const data = JSON.parse(stored);
      const now = new Date();
      const isExpired = new Date(data.cache_expiry) <= now;

      if (isExpired) {
        console.log(
          "[ListCollectionStorageService] Shared collections list cache expired",
        );
        return { collections: [], isExpired: true };
      }

      console.log(
        "[ListCollectionStorageService] Retrieved shared collections list from cache:",
        data.collections.length,
      );

      return { collections: data.collections, isExpired: false };
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get shared collections list:",
        error,
      );
      return { collections: [], isExpired: false };
    }
  }

  // === Filtered Collections Operations ===

  // Store filtered collections
  storeFilteredCollections(ownedCollections, sharedCollections, totalCount) {
    try {
      const filteredData = {
        owned_collections: ownedCollections || [],
        shared_collections: sharedCollections || [],
        total_count: totalCount,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
      };

      localStorage.setItem(
        this.STORAGE_KEYS.FILTERED_COLLECTIONS,
        JSON.stringify(filteredData),
      );

      console.log(
        "[ListCollectionStorageService] Filtered collections cached:",
        {
          owned: ownedCollections.length,
          shared: sharedCollections.length,
          total: totalCount,
        },
      );

      this.updateListMetadata("filtered_collections", totalCount);
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to store filtered collections:",
        error,
      );
      return false;
    }
  }

  // Get filtered collections
  getFilteredCollections() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.FILTERED_COLLECTIONS,
      );
      if (!stored) {
        return {
          owned_collections: [],
          shared_collections: [],
          total_count: 0,
          isExpired: false,
        };
      }

      const data = JSON.parse(stored);
      const now = new Date();
      const isExpired = new Date(data.cache_expiry) <= now;

      if (isExpired) {
        console.log(
          "[ListCollectionStorageService] Filtered collections cache expired",
        );
        return {
          owned_collections: [],
          shared_collections: [],
          total_count: 0,
          isExpired: true,
        };
      }

      console.log(
        "[ListCollectionStorageService] Retrieved filtered collections from cache:",
        {
          owned: data.owned_collections.length,
          shared: data.shared_collections.length,
          total: data.total_count,
        },
      );

      return { ...data, isExpired: false };
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get filtered collections:",
        error,
      );
      return {
        owned_collections: [],
        shared_collections: [],
        total_count: 0,
        isExpired: false,
      };
    }
  }

  // === Root Collections Operations ===

  // Store root collections
  storeRootCollections(collections) {
    try {
      const rootData = {
        collections: collections,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
        total_count: collections.length,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.ROOT_COLLECTIONS,
        JSON.stringify(rootData),
      );

      console.log(
        "[ListCollectionStorageService] Root collections cached:",
        collections.length,
      );

      this.updateListMetadata("root_collections", collections.length);
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to store root collections:",
        error,
      );
      return false;
    }
  }

  // Get root collections
  getRootCollections() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.ROOT_COLLECTIONS);
      if (!stored) return { collections: [], isExpired: false };

      const data = JSON.parse(stored);
      const now = new Date();
      const isExpired = new Date(data.cache_expiry) <= now;

      if (isExpired) {
        console.log(
          "[ListCollectionStorageService] Root collections cache expired",
        );
        return { collections: [], isExpired: true };
      }

      console.log(
        "[ListCollectionStorageService] Retrieved root collections from cache:",
        data.collections.length,
      );

      return { collections: data.collections, isExpired: false };
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get root collections:",
        error,
      );
      return { collections: [], isExpired: false };
    }
  }

  // === Collections by Parent Operations ===

  // Store collections by parent
  storeCollectionsByParent(parentId, collections) {
    try {
      const existingData = this.getCollectionsByParentData();

      const parentData = {
        collections: collections,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
        total_count: collections.length,
      };

      existingData[parentId] = parentData;

      localStorage.setItem(
        this.STORAGE_KEYS.COLLECTIONS_BY_PARENT,
        JSON.stringify(existingData),
      );

      console.log(
        "[ListCollectionStorageService] Collections by parent cached:",
        { parentId, count: collections.length },
      );

      this.updateListMetadata("collections_by_parent", collections.length);
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to store collections by parent:",
        error,
      );
      return false;
    }
  }

  // Get collections by parent
  getCollectionsByParent(parentId) {
    try {
      const allData = this.getCollectionsByParentData();
      const parentData = allData[parentId];

      if (!parentData) {
        return { collections: [], isExpired: false };
      }

      const now = new Date();
      const isExpired = new Date(parentData.cache_expiry) <= now;

      if (isExpired) {
        console.log(
          "[ListCollectionStorageService] Collections by parent cache expired:",
          parentId,
        );
        // Remove expired data
        delete allData[parentId];
        localStorage.setItem(
          this.STORAGE_KEYS.COLLECTIONS_BY_PARENT,
          JSON.stringify(allData),
        );
        return { collections: [], isExpired: true };
      }

      console.log(
        "[ListCollectionStorageService] Retrieved collections by parent from cache:",
        { parentId, count: parentData.collections.length },
      );

      return { collections: parentData.collections, isExpired: false };
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get collections by parent:",
        error,
      );
      return { collections: [], isExpired: false };
    }
  }

  // Get all collections by parent data
  getCollectionsByParentData() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.COLLECTIONS_BY_PARENT,
      );
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get collections by parent data:",
        error,
      );
      return {};
    }
  }

  // === Cache Management ===

  // Check if collections list is cached and valid
  isListCached() {
    const data = this.getListedCollections();
    return !data.isExpired && data.collections.length > 0;
  }

  // Check if shared collections are cached and valid - NEW METHOD
  isSharedListCached() {
    const data = this.getSharedCollections();
    return !data.isExpired && data.collections.length > 0;
  }

  // Clear all cached collection lists
  clearAllListCache() {
    try {
      Object.values(this.STORAGE_KEYS).forEach((key) => {
        localStorage.removeItem(key);
      });

      console.log("[ListCollectionStorageService] All list cache cleared");
      return true;
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to clear list cache:",
        error,
      );
      return false;
    }
  }

  // Clear specific cache type
  clearSpecificCache(cacheType) {
    try {
      const keyMap = {
        listed: this.STORAGE_KEYS.LISTED_COLLECTIONS,
        shared: this.STORAGE_KEYS.SHARED_COLLECTIONS, // NEW MAPPING
        filtered: this.STORAGE_KEYS.FILTERED_COLLECTIONS,
        root: this.STORAGE_KEYS.ROOT_COLLECTIONS,
        byParent: this.STORAGE_KEYS.COLLECTIONS_BY_PARENT,
      };

      const key = keyMap[cacheType];
      if (key) {
        localStorage.removeItem(key);
        console.log(
          `[ListCollectionStorageService] ${cacheType} cache cleared`,
        );
        return true;
      }

      console.warn(
        `[ListCollectionStorageService] Unknown cache type: ${cacheType}`,
      );
      return false;
    } catch (error) {
      console.error(
        `[ListCollectionStorageService] Failed to clear ${cacheType} cache:`,
        error,
      );
      return false;
    }
  }

  // === Metadata Management ===

  // Update list metadata
  updateListMetadata(listType, count) {
    try {
      const metadata = this.getListMetadata() || {};

      metadata[listType] = {
        count: count,
        lastUpdated: new Date().toISOString(),
        cacheExpiryMs: this.CACHE_EXPIRY_MS,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.LIST_METADATA,
        JSON.stringify(metadata),
      );
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to update list metadata:",
        error,
      );
    }
  }

  // Get list metadata
  getListMetadata() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.LIST_METADATA);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error(
        "[ListCollectionStorageService] Failed to get list metadata:",
        error,
      );
      return {};
    }
  }

  // === Search and Filter ===

  // Search collections in any cached list
  searchCachedCollections(searchTerm, collections) {
    if (!searchTerm || !collections || !Array.isArray(collections)) {
      return collections || [];
    }

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

  // Filter collections by type
  filterCollectionsByType(collections, type) {
    if (!collections || !Array.isArray(collections)) return [];

    if (!type) return collections;

    return collections.filter(
      (collection) => collection.collection_type === type,
    );
  }

  // === Statistics ===

  // Get storage statistics
  getStorageStats() {
    const listedData = this.getListedCollections();
    const sharedData = this.getSharedCollections(); // NEW
    const filteredData = this.getFilteredCollections();
    const rootData = this.getRootCollections();
    const metadata = this.getListMetadata();

    return {
      listed: {
        count: listedData.collections.length,
        isExpired: listedData.isExpired,
      },
      shared: {
        // NEW
        count: sharedData.collections.length,
        isExpired: sharedData.isExpired,
      },
      filtered: {
        ownedCount: filteredData.owned_collections.length,
        sharedCount: filteredData.shared_collections.length,
        totalCount: filteredData.total_count,
        isExpired: filteredData.isExpired,
      },
      root: {
        count: rootData.collections.length,
        isExpired: rootData.isExpired,
      },
      cacheExpiryMinutes: this.CACHE_EXPIRY_MS / (60 * 1000),
      metadata,
    };
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const stats = this.getStorageStats();

    return {
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasListedCollections: stats.listed.count > 0 && !stats.listed.isExpired,
      hasSharedCollections: stats.shared.count > 0 && !stats.shared.isExpired, // NEW
      hasFilteredCollections:
        stats.filtered.totalCount > 0 && !stats.filtered.isExpired,
      hasRootCollections: stats.root.count > 0 && !stats.root.isExpired,
      cacheExpiryMs: this.CACHE_EXPIRY_MS,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "ListCollectionStorageService",
      storageInfo: this.getStorageInfo(),
      metadata: this.getListMetadata(),
    };
  }
}

export default ListCollectionStorageService;

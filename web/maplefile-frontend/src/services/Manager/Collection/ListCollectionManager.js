// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/ListCollectionManager.js
// List Collection Manager - Orchestrates API, Storage, and Crypto services for collection listing

import ListCollectionAPIService from "../../API/Collection/ListCollectionAPIService.js";
import ListCollectionStorageService from "../../Storage/Collection/ListCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class ListCollectionManager {
  constructor(authManager) {
    // ListCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new ListCollectionAPIService(authManager);
    this.storageService = new ListCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection listing events
    this.collectionListingListeners = new Set();

    console.log(
      "[ListCollectionManager] Collection listing manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[ListCollectionManager] Initializing collection listing manager...",
      );

      // Initialize crypto service
      await this.cryptoService.initialize();

      console.log(
        "[ListCollectionManager] Collection listing manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[ListCollectionManager] Failed to initialize collection listing manager:",
        error,
      );
    }
  }

  // === Collection Listing with Caching ===

  // List user collections with caching - Uses PasswordStorageService automatically
  async listCollections(forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[ListCollectionManager] Listing user collections");

      // Check cache first (unless force refresh is requested)
      if (!forceRefresh) {
        const cachedData = this.storageService.getListedCollections();
        if (!cachedData.isExpired && cachedData.collections.length > 0) {
          console.log(
            "[ListCollectionManager] Collections found in cache, attempting decryption",
          );

          try {
            const decryptedCollections = await this.decryptCollections(
              cachedData.collections,
            );

            // Notify listeners
            this.notifyCollectionListingListeners(
              "collections_listed_from_cache",
              {
                totalCount: decryptedCollections.length,
                fromCache: true,
              },
            );

            console.log(
              "[ListCollectionManager] Collections listed from cache successfully",
            );
            return {
              collections: decryptedCollections,
              source: "cache",
              success: true,
              totalCount: decryptedCollections.length,
            };
          } catch (decryptError) {
            console.warn(
              "[ListCollectionManager] Failed to decrypt cached collections, fetching from API:",
              decryptError.message,
            );
            // Fall through to API fetch
          }
        }
      }

      console.log("[ListCollectionManager] Fetching collections from API");

      // Fetch from API
      const response = await this.apiService.listCollections();

      // Validate response
      this.apiService.validateCollectionsResponse(response);

      // Cache the encrypted collections
      this.storageService.storeListedCollections(response.collections);

      // Decrypt the collections
      const decryptedCollections = await this.decryptCollections(
        response.collections,
      );

      // Notify listeners
      this.notifyCollectionListingListeners("collections_listed_from_api", {
        totalCount: decryptedCollections.length,
        fromCache: false,
        forceRefresh,
      });

      console.log(
        "[ListCollectionManager] Collections listed from API successfully",
      );

      return {
        collections: decryptedCollections,
        source: forceRefresh ? "api_refresh" : "api",
        success: true,
        totalCount: decryptedCollections.length,
      };
    } catch (error) {
      console.error(
        "[ListCollectionManager] Collection listing failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionListingListeners("collection_listing_failed", {
        error: error.message,
        forceRefresh,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // List filtered collections (owned/shared) - Uses PasswordStorageService automatically
  async listFilteredCollections(
    includeOwned = true,
    includeShared = false,
    forceRefresh = false,
  ) {
    try {
      this.isLoading = true;
      console.log("[ListCollectionManager] Listing filtered collections:", {
        includeOwned,
        includeShared,
        forceRefresh,
      });

      // Check cache first (unless force refresh is requested)
      if (!forceRefresh) {
        const cachedData = this.storageService.getFilteredCollections();
        if (!cachedData.isExpired && cachedData.total_count > 0) {
          console.log(
            "[ListCollectionManager] Filtered collections found in cache, attempting decryption",
          );

          try {
            const decryptedOwned = await this.decryptCollections(
              cachedData.owned_collections,
            );
            const decryptedShared = await this.decryptCollections(
              cachedData.shared_collections,
            );

            // Notify listeners
            this.notifyCollectionListingListeners(
              "filtered_collections_listed_from_cache",
              {
                ownedCount: decryptedOwned.length,
                sharedCount: decryptedShared.length,
                totalCount: cachedData.total_count,
                fromCache: true,
              },
            );

            console.log(
              "[ListCollectionManager] Filtered collections listed from cache successfully",
            );

            return {
              owned_collections: decryptedOwned,
              shared_collections: decryptedShared,
              total_count: cachedData.total_count,
              source: "cache",
              success: true,
            };
          } catch (decryptError) {
            console.warn(
              "[ListCollectionManager] Failed to decrypt cached filtered collections, fetching from API:",
              decryptError.message,
            );
            // Fall through to API fetch
          }
        }
      }

      console.log(
        "[ListCollectionManager] Fetching filtered collections from API",
      );

      // Fetch from API
      const response = await this.apiService.listFilteredCollections(
        includeOwned,
        includeShared,
      );

      // Cache the encrypted collections
      this.storageService.storeFilteredCollections(
        response.owned_collections || [],
        response.shared_collections || [],
        response.total_count,
      );

      // Decrypt the collections
      const decryptedOwned = await this.decryptCollections(
        response.owned_collections || [],
      );
      const decryptedShared = await this.decryptCollections(
        response.shared_collections || [],
      );

      // Notify listeners
      this.notifyCollectionListingListeners(
        "filtered_collections_listed_from_api",
        {
          ownedCount: decryptedOwned.length,
          sharedCount: decryptedShared.length,
          totalCount: response.total_count,
          fromCache: false,
          forceRefresh,
        },
      );

      console.log(
        "[ListCollectionManager] Filtered collections listed from API successfully",
      );

      return {
        owned_collections: decryptedOwned,
        shared_collections: decryptedShared,
        total_count: response.total_count,
        source: forceRefresh ? "api_refresh" : "api",
        success: true,
      };
    } catch (error) {
      console.error(
        "[ListCollectionManager] Filtered collection listing failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionListingListeners(
        "filtered_collection_listing_failed",
        {
          error: error.message,
          includeOwned,
          includeShared,
          forceRefresh,
        },
      );

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // List root collections - Uses PasswordStorageService automatically
  async listRootCollections(forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[ListCollectionManager] Listing root collections");

      // Check cache first (unless force refresh is requested)
      if (!forceRefresh) {
        const cachedData = this.storageService.getRootCollections();
        if (!cachedData.isExpired && cachedData.collections.length > 0) {
          console.log(
            "[ListCollectionManager] Root collections found in cache, attempting decryption",
          );

          try {
            const decryptedCollections = await this.decryptCollections(
              cachedData.collections,
            );

            // Notify listeners
            this.notifyCollectionListingListeners(
              "root_collections_listed_from_cache",
              {
                totalCount: decryptedCollections.length,
                fromCache: true,
              },
            );

            console.log(
              "[ListCollectionManager] Root collections listed from cache successfully",
            );

            return {
              collections: decryptedCollections,
              source: "cache",
              success: true,
              totalCount: decryptedCollections.length,
            };
          } catch (decryptError) {
            console.warn(
              "[ListCollectionManager] Failed to decrypt cached root collections, fetching from API:",
              decryptError.message,
            );
            // Fall through to API fetch
          }
        }
      }

      console.log("[ListCollectionManager] Fetching root collections from API");

      // Fetch from API
      const response = await this.apiService.listRootCollections();

      // Validate response
      this.apiService.validateCollectionsResponse(response);

      // Cache the encrypted collections
      this.storageService.storeRootCollections(response.collections);

      // Decrypt the collections
      const decryptedCollections = await this.decryptCollections(
        response.collections,
      );

      // Notify listeners
      this.notifyCollectionListingListeners(
        "root_collections_listed_from_api",
        {
          totalCount: decryptedCollections.length,
          fromCache: false,
          forceRefresh,
        },
      );

      console.log(
        "[ListCollectionManager] Root collections listed from API successfully",
      );

      return {
        collections: decryptedCollections,
        source: forceRefresh ? "api_refresh" : "api",
        success: true,
        totalCount: decryptedCollections.length,
      };
    } catch (error) {
      console.error(
        "[ListCollectionManager] Root collection listing failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionListingListeners("root_collection_listing_failed", {
        error: error.message,
        forceRefresh,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // List collections by parent - Uses PasswordStorageService automatically
  async listCollectionsByParent(parentId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log(
        "[ListCollectionManager] Listing collections by parent:",
        parentId,
      );

      if (!parentId) {
        throw new Error("Parent ID is required");
      }

      // Check cache first (unless force refresh is requested)
      if (!forceRefresh) {
        const cachedData = this.storageService.getCollectionsByParent(parentId);
        if (!cachedData.isExpired && cachedData.collections.length > 0) {
          console.log(
            "[ListCollectionManager] Collections by parent found in cache, attempting decryption",
          );

          try {
            const decryptedCollections = await this.decryptCollections(
              cachedData.collections,
            );

            // Notify listeners
            this.notifyCollectionListingListeners(
              "collections_by_parent_listed_from_cache",
              {
                parentId,
                totalCount: decryptedCollections.length,
                fromCache: true,
              },
            );

            console.log(
              "[ListCollectionManager] Collections by parent listed from cache successfully",
            );

            return {
              collections: decryptedCollections,
              parentId,
              source: "cache",
              success: true,
              totalCount: decryptedCollections.length,
            };
          } catch (decryptError) {
            console.warn(
              "[ListCollectionManager] Failed to decrypt cached collections by parent, fetching from API:",
              decryptError.message,
            );
            // Fall through to API fetch
          }
        }
      }

      console.log(
        "[ListCollectionManager] Fetching collections by parent from API",
      );

      // Fetch from API
      const response = await this.apiService.listCollectionsByParent(parentId);

      // Validate response
      this.apiService.validateCollectionsResponse(response);

      // Cache the encrypted collections
      this.storageService.storeCollectionsByParent(
        parentId,
        response.collections,
      );

      // Decrypt the collections
      const decryptedCollections = await this.decryptCollections(
        response.collections,
      );

      // Notify listeners
      this.notifyCollectionListingListeners(
        "collections_by_parent_listed_from_api",
        {
          parentId,
          totalCount: decryptedCollections.length,
          fromCache: false,
          forceRefresh,
        },
      );

      console.log(
        "[ListCollectionManager] Collections by parent listed from API successfully",
      );

      return {
        collections: decryptedCollections,
        parentId,
        source: forceRefresh ? "api_refresh" : "api",
        success: true,
        totalCount: decryptedCollections.length,
      };
    } catch (error) {
      console.error(
        "[ListCollectionManager] Collections by parent listing failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionListingListeners(
        "collections_by_parent_listing_failed",
        {
          parentId,
          error: error.message,
          forceRefresh,
        },
      );

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Collection Decryption ===

  // Decrypt multiple collections (uses PasswordStorageService automatically)
  async decryptCollections(encryptedCollections) {
    if (!encryptedCollections || !Array.isArray(encryptedCollections)) {
      return [];
    }

    console.log(
      "[ListCollectionManager] Decrypting collections:",
      encryptedCollections.length,
    );

    const decryptedCollections = [];

    for (const encryptedCollection of encryptedCollections) {
      try {
        // Use CollectionCryptoService to decrypt the collection (it uses PasswordStorageService automatically)
        const decryptedCollection =
          await this.cryptoService.decryptCollectionFromAPI(
            encryptedCollection,
          );

        decryptedCollections.push(decryptedCollection);
      } catch (error) {
        console.warn(
          `[ListCollectionManager] Failed to decrypt collection ${encryptedCollection.id}:`,
          error.message,
        );

        // Add collection with error marker
        decryptedCollections.push({
          ...encryptedCollection,
          name: "[Unable to decrypt]",
          _isDecrypted: false,
          _decryptionError: error.message,
        });
      }
    }

    console.log(
      `[ListCollectionManager] Decrypted ${decryptedCollections.filter((c) => c._isDecrypted).length}/${encryptedCollections.length} collections successfully`,
    );

    return decryptedCollections;
  }

  // === Password Management ===

  // Get user password from password storage service
  async getUserPassword() {
    try {
      const { default: passwordStorageService } = await import(
        "../../PasswordStorageService.js"
      );
      return passwordStorageService.getPassword();
    } catch (error) {
      console.error(
        "[ListCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Cache Management ===

  // Get cached collections
  getCachedCollections() {
    return this.storageService.getListedCollections();
  }

  // Get cached filtered collections
  getCachedFilteredCollections() {
    return this.storageService.getFilteredCollections();
  }

  // Get cached root collections
  getCachedRootCollections() {
    return this.storageService.getRootCollections();
  }

  // Get cached collections by parent
  getCachedCollectionsByParent(parentId) {
    return this.storageService.getCollectionsByParent(parentId);
  }

  // Clear all cache
  clearAllCache() {
    try {
      console.log(
        "[ListCollectionManager] Clearing all cached collection lists",
      );

      this.storageService.clearAllListCache();

      this.notifyCollectionListingListeners("all_list_cache_cleared", {});

      console.log(
        "[ListCollectionManager] All cached collection lists cleared",
      );
    } catch (error) {
      console.error(
        "[ListCollectionManager] Failed to clear all cache:",
        error,
      );
      throw error;
    }
  }

  // Clear specific cache
  clearSpecificCache(cacheType) {
    try {
      console.log(`[ListCollectionManager] Clearing ${cacheType} cache`);

      const success = this.storageService.clearSpecificCache(cacheType);

      if (success) {
        this.notifyCollectionListingListeners(`${cacheType}_cache_cleared`, {
          cacheType,
        });
      }

      return success;
    } catch (error) {
      console.error(
        `[ListCollectionManager] Failed to clear ${cacheType} cache:`,
        error,
      );
      throw error;
    }
  }

  // === Search and Filter ===

  // Search collections
  searchCollections(searchTerm, collections) {
    return this.storageService.searchCachedCollections(searchTerm, collections);
  }

  // Filter collections by type
  filterCollectionsByType(collections, type) {
    return this.storageService.filterCollectionsByType(collections, type);
  }

  // === Event Management ===

  // Add collection listing listener
  addCollectionListingListener(callback) {
    if (typeof callback === "function") {
      this.collectionListingListeners.add(callback);
      console.log(
        "[ListCollectionManager] Collection listing listener added. Total listeners:",
        this.collectionListingListeners.size,
      );
    }
  }

  // Remove collection listing listener
  removeCollectionListingListener(callback) {
    this.collectionListingListeners.delete(callback);
    console.log(
      "[ListCollectionManager] Collection listing listener removed. Total listeners:",
      this.collectionListingListeners.size,
    );
  }

  // Notify collection listing listeners
  notifyCollectionListingListeners(eventType, eventData) {
    console.log(
      `[ListCollectionManager] Notifying ${this.collectionListingListeners.size} listeners of ${eventType}`,
    );

    this.collectionListingListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[ListCollectionManager] Error in collection listing listener:",
          error,
        );
      }
    });
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();
    const cryptoStatus = this.cryptoService.getStatus();
    const storageStats = this.storageService.getStorageStats();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canListCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      crypto: cryptoStatus,
      stats: storageStats,
      listenerCount: this.collectionListingListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "ListCollectionManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      cryptoService: this.cryptoService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default ListCollectionManager;

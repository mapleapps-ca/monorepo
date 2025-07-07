// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/GetCollectionManager.js
// Get Collection Manager - Orchestrates API, Storage, and Crypto services for collection retrieval

import GetCollectionAPIService from "../../API/Collection/GetCollectionAPIService.js";
import GetCollectionStorageService from "../../Storage/Collection/GetCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class GetCollectionManager {
  constructor(authManager) {
    // GetCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new GetCollectionAPIService(authManager);
    this.storageService = new GetCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection retrieval events
    this.collectionRetrievalListeners = new Set();

    console.log(
      "[GetCollectionManager] Collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[GetCollectionManager] Initializing collection manager...");

      // Initialize crypto service
      await this.cryptoService.initialize();

      // Clear expired collections from cache
      const expiredCount = this.storageService.clearExpiredCollections();
      if (expiredCount > 0) {
        console.log(
          `[GetCollectionManager] Cleared ${expiredCount} expired collections from cache`,
        );
      }

      console.log(
        "[GetCollectionManager] Collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[GetCollectionManager] Failed to initialize collection manager:",
        error,
      );
    }
  }

  // === Collection Retrieval with Caching ===

  // Get collection with caching (check cache first, then API) - Uses PasswordStorageService automatically
  async getCollection(collectionId, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[GetCollectionManager] Getting collection:", collectionId);

      // Validate input
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      // Check cache first (unless force refresh is requested)
      if (!forceRefresh) {
        const cachedCollection =
          this.storageService.getCachedCollection(collectionId);
        if (cachedCollection) {
          console.log(
            "[GetCollectionManager] Collection found in cache, attempting decryption",
          );

          try {
            const decryptedCollection =
              await this.decryptCollection(cachedCollection);

            // Notify listeners
            this.notifyCollectionRetrievalListeners(
              "collection_retrieved_from_cache",
              {
                collectionId,
                name: decryptedCollection.name || "[Encrypted]",
                fromCache: true,
              },
            );

            console.log(
              "[GetCollectionManager] Collection retrieved from cache successfully",
            );
            return {
              collection: decryptedCollection,
              source: "cache",
              success: true,
            };
          } catch (decryptError) {
            console.warn(
              "[GetCollectionManager] Failed to decrypt cached collection, fetching from API:",
              decryptError.message,
            );
            // Fall through to API fetch
          }
        }
      }

      console.log("[GetCollectionManager] Fetching collection from API");

      // Fetch from API
      const encryptedCollection =
        await this.apiService.getCollection(collectionId);

      // Cache the encrypted collection
      this.storageService.storeRetrievedCollection(encryptedCollection);

      // Decrypt the collection
      const decryptedCollection =
        await this.decryptCollection(encryptedCollection);

      // Notify listeners
      this.notifyCollectionRetrievalListeners("collection_retrieved_from_api", {
        collectionId,
        name: decryptedCollection.name || "[Encrypted]",
        fromCache: false,
        forceRefresh,
      });

      console.log(
        "[GetCollectionManager] Collection retrieved from API successfully",
      );

      return {
        collection: decryptedCollection,
        source: forceRefresh ? "api_refresh" : "api",
        success: true,
      };
    } catch (error) {
      console.error(
        "[GetCollectionManager] Collection retrieval failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionRetrievalListeners("collection_retrieval_failed", {
        collectionId,
        error: error.message,
        forceRefresh,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get collection from cache only (no API call) - Uses PasswordStorageService automatically
  async getCachedCollection(collectionId) {
    try {
      console.log(
        "[GetCollectionManager] Getting collection from cache only:",
        collectionId,
      );

      const cachedCollection =
        this.storageService.getCachedCollection(collectionId);
      if (!cachedCollection) {
        throw new Error("Collection not found in cache");
      }

      const decryptedCollection =
        await this.decryptCollection(cachedCollection);

      console.log(
        "[GetCollectionManager] Collection retrieved from cache successfully",
      );
      return {
        collection: decryptedCollection,
        source: "cache_only",
        success: true,
      };
    } catch (error) {
      console.error(
        "[GetCollectionManager] Failed to get cached collection:",
        error,
      );
      throw error;
    }
  }

  // Refresh collection from API (force update cache) - Uses PasswordStorageService automatically
  async refreshCollection(collectionId) {
    console.log(
      "[GetCollectionManager] Force refreshing collection from API:",
      collectionId,
    );
    return this.getCollection(collectionId, true);
  }

  // Get multiple collections (batch operation) - Uses PasswordStorageService automatically
  async getCollections(collectionIds, forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log(
        "[GetCollectionManager] Getting multiple collections:",
        collectionIds.length,
      );

      if (!Array.isArray(collectionIds) || collectionIds.length === 0) {
        throw new Error("Collection IDs array is required");
      }

      const results = [];
      const errors = [];

      for (const collectionId of collectionIds) {
        try {
          const result = await this.getCollection(collectionId, forceRefresh);
          results.push(result);
        } catch (error) {
          errors.push({ collectionId, error: error.message });
          console.warn(
            `[GetCollectionManager] Failed to get collection ${collectionId}:`,
            error.message,
          );
        }
      }

      // Notify listeners
      this.notifyCollectionRetrievalListeners(
        "multiple_collections_retrieved",
        {
          requestedCount: collectionIds.length,
          successCount: results.length,
          errorCount: errors.length,
          forceRefresh,
        },
      );

      console.log(
        `[GetCollectionManager] Batch retrieval completed: ${results.length} successful, ${errors.length} failed`,
      );

      return {
        collections: results,
        errors: errors,
        successCount: results.length,
        errorCount: errors.length,
        success: true,
      };
    } catch (error) {
      console.error(
        "[GetCollectionManager] Batch collection retrieval failed:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Collection Decryption ===

  // Decrypt collection data (uses PasswordStorageService automatically)
  async decryptCollection(encryptedCollection) {
    try {
      console.log(
        "[GetCollectionManager] Decrypting collection:",
        encryptedCollection.id,
      );

      // Get collection key from storage
      let collectionKey = this.storageService.getCollectionKey(
        encryptedCollection.id,
      );

      // Use CollectionCryptoService to decrypt the collection (it uses PasswordStorageService automatically)
      const decryptedCollection =
        await this.cryptoService.decryptCollectionFromAPI(
          encryptedCollection,
          collectionKey,
        );

      // If we successfully decrypted and didn't have the key cached, cache it now
      if (decryptedCollection._isDecrypted && !collectionKey) {
        try {
          // The crypto service already cached the key during decryption
          console.log(
            "[GetCollectionManager] Collection key was cached during decryption",
          );
        } catch (keyError) {
          console.warn(
            "[GetCollectionManager] Could not verify collection key cache:",
            keyError,
          );
        }
      }

      return decryptedCollection;
    } catch (error) {
      console.error(
        "[GetCollectionManager] Failed to decrypt collection:",
        error,
      );
      throw error;
    }
  }

  // === Collection Existence Check ===

  // Check if collection exists (lightweight check)
  async collectionExists(collectionId) {
    try {
      console.log(
        "[GetCollectionManager] Checking if collection exists:",
        collectionId,
      );

      // First check cache
      if (this.storageService.isCollectionCached(collectionId)) {
        console.log("[GetCollectionManager] Collection exists in cache");
        return true;
      }

      // Check via API
      const exists = await this.apiService.collectionExists(collectionId);
      console.log(
        `[GetCollectionManager] Collection exists via API: ${exists}`,
      );
      return exists;
    } catch (error) {
      console.error(
        "[GetCollectionManager] Failed to check collection existence:",
        error,
      );
      throw error;
    }
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
        "[GetCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Cache Management ===

  // Get all cached collections
  getCachedCollections() {
    return this.storageService.getRetrievedCollections();
  }

  // Get cache status for a collection
  getCollectionCacheStatus(collectionId) {
    return this.storageService.getCollectionCacheStatus(collectionId);
  }

  // Clear expired collections from cache
  clearExpiredCollections() {
    return this.storageService.clearExpiredCollections();
  }

  // Remove specific collection from cache
  removeFromCache(collectionId) {
    try {
      console.log(
        "[GetCollectionManager] Removing collection from cache:",
        collectionId,
      );

      const removed = this.storageService.removeFromCache(collectionId);

      if (removed) {
        this.notifyCollectionRetrievalListeners(
          "collection_removed_from_cache",
          {
            collectionId,
          },
        );
      }

      return removed;
    } catch (error) {
      console.error(
        "[GetCollectionManager] Failed to remove collection from cache:",
        error,
      );
      throw error;
    }
  }

  // Clear all cached collections
  clearAllCache() {
    try {
      console.log("[GetCollectionManager] Clearing all cached collections");

      this.storageService.clearAllCachedCollections();

      this.notifyCollectionRetrievalListeners("all_cache_cleared", {});

      console.log("[GetCollectionManager] All cached collections cleared");
    } catch (error) {
      console.error("[GetCollectionManager] Failed to clear all cache:", error);
      throw error;
    }
  }

  // === Search and Filter ===

  // Search cached collections
  searchCachedCollections(searchTerm) {
    const cachedCollections = this.getCachedCollections();
    return this.storageService.searchCachedCollections(
      searchTerm,
      cachedCollections,
    );
  }

  // === Event Management ===

  // Add collection retrieval listener
  addCollectionRetrievalListener(callback) {
    if (typeof callback === "function") {
      this.collectionRetrievalListeners.add(callback);
      console.log(
        "[GetCollectionManager] Collection retrieval listener added. Total listeners:",
        this.collectionRetrievalListeners.size,
      );
    }
  }

  // Remove collection retrieval listener
  removeCollectionRetrievalListener(callback) {
    this.collectionRetrievalListeners.delete(callback);
    console.log(
      "[GetCollectionManager] Collection retrieval listener removed. Total listeners:",
      this.collectionRetrievalListeners.size,
    );
  }

  // Notify collection retrieval listeners
  notifyCollectionRetrievalListeners(eventType, eventData) {
    console.log(
      `[GetCollectionManager] Notifying ${this.collectionRetrievalListeners.size} listeners of ${eventType}`,
    );

    this.collectionRetrievalListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[GetCollectionManager] Error in collection retrieval listener:",
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
    const cacheStats = this.storageService.getCacheStats();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canGetCollections: this.authManager.canMakeAuthenticatedRequests(),
      cache: cacheStats,
      storage: storageInfo,
      crypto: cryptoStatus,
      listenerCount: this.collectionRetrievalListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "GetCollectionManager",
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

export default GetCollectionManager;

// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/CreateCollectionManager.js
// Create Collection Manager - Orchestrates API, Storage, and Crypto services for collection creation (FIXED with cache invalidation)

import CreateCollectionAPIService from "../../API/Collection/CreateCollectionAPIService.js";
import CreateCollectionStorageService from "../../Storage/Collection/CreateCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class CreateCollectionManager {
  constructor(authManager) {
    // CreateCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new CreateCollectionAPIService(authManager);
    this.storageService = new CreateCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection creation events
    this.collectionCreationListeners = new Set();

    // ADDED: Reference to other managers for cache invalidation
    this.listCollectionManager = null;

    console.log(
      "[CreateCollectionManager] Collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[CreateCollectionManager] Initializing collection manager...",
      );

      // Initialize crypto service
      await this.cryptoService.initialize();

      console.log(
        "[CreateCollectionManager] Collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to initialize collection manager:",
        error,
      );
    }
  }

  // ADDED: Set reference to ListCollectionManager for cache invalidation
  setListCollectionManager(listCollectionManager) {
    this.listCollectionManager = listCollectionManager;
    console.log(
      "[CreateCollectionManager] ListCollectionManager reference set for cache invalidation",
    );
  }

  // ADDED: Invalidate related caches when collection is created
  invalidateRelatedCaches(collectionData) {
    try {
      console.log(
        "[CreateCollectionManager] Invalidating related caches after collection creation",
      );

      // Invalidate ListCollectionManager cache
      if (this.listCollectionManager) {
        this.listCollectionManager.clearAllCache();
        console.log(
          "[CreateCollectionManager] ListCollectionManager cache cleared",
        );
      }

      // Dispatch appropriate events based on collection type
      if (typeof window !== "undefined") {
        // Always dispatch the generic collection created event
        window.dispatchEvent(
          new CustomEvent("collectionCreated", {
            detail: {
              timestamp: Date.now(),
              source: "CreateCollectionManager",
              collectionData: collectionData,
            },
          }),
        );

        // If it's a root collection (no parent), also dispatch rootCollectionCreated
        if (!collectionData.parent_id) {
          window.dispatchEvent(
            new CustomEvent("rootCollectionCreated", {
              detail: {
                timestamp: Date.now(),
                source: "CreateCollectionManager",
                collectionData: collectionData,
              },
            }),
          );
          console.log(
            "[CreateCollectionManager] rootCollectionCreated event dispatched",
          );
        }

        console.log(
          "[CreateCollectionManager] collectionCreated event dispatched",
        );
      }
    } catch (error) {
      console.warn(
        "[CreateCollectionManager] Failed to invalidate some caches:",
        error,
      );
    }
  }

  // === Collection Creation with Encryption ===

  // Create collection with full E2EE encryption
  async createCollection(collectionData, password = null) {
    try {
      this.isLoading = true;
      console.log("[CreateCollectionManager] Starting collection creation");
      console.log("[CreateCollectionManager] Collection data:", {
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
        hasParent: !!collectionData.parent_id,
      });

      // Validate input early
      if (!collectionData.name || !collectionData.name.trim()) {
        throw new Error("Collection name is required");
      }

      // Get password for encryption
      const userPassword = password || (await this.getUserPassword());
      if (!userPassword) {
        throw new Error("Password required for collection encryption");
      }

      console.log("[CreateCollectionManager] Encrypting collection data");

      // Use CollectionCryptoService to encrypt all collection data for API
      const { apiData, collectionKey, collectionId } =
        await this.cryptoService.encryptCollectionForAPI(
          collectionData,
          userPassword,
        );

      console.log(
        "[CreateCollectionManager] Collection data encrypted successfully",
      );
      console.log("[CreateCollectionManager] API data prepared:", {
        id: apiData.id,
        hasEncryptedName: !!apiData.encrypted_name,
        hasEncryptedKey: !!apiData.encrypted_collection_key,
        collection_type: apiData.collection_type,
      });

      console.log("[CreateCollectionManager] Sending to API");

      // Create collection via API
      const createdCollection = await this.apiService.createCollection(apiData);

      console.log(
        "[CreateCollectionManager] Collection created via API successfully",
      );

      // Store collection key in memory for future use
      this.storageService.storeCollectionKey(collectionId, collectionKey);

      // Prepare decrypted collection for local storage
      const decryptedCollection = {
        ...createdCollection,
        name: collectionData.name, // Store decrypted name
        _originalEncryptedName: apiData.encrypted_name,
        _hasCollectionKey: true,
      };

      // Store in local storage
      this.storageService.storeCreatedCollection(decryptedCollection);

      // ADDED: Invalidate related caches to ensure fresh data
      this.invalidateRelatedCaches(collectionData);

      // Notify listeners
      this.notifyCollectionCreationListeners("collection_created", {
        collectionId,
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
        hasParent: !!collectionData.parent_id,
        parentId: collectionData.parent_id,
      });

      console.log(
        "[CreateCollectionManager] Collection created successfully:",
        collectionId,
      );

      return {
        collection: decryptedCollection,
        collectionId,
        success: true,
      };
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Collection creation failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionCreationListeners("collection_creation_failed", {
        error: error.message,
        collectionData,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Collection Decryption ===

  // Decrypt collection data (for displaying stored collections)
  async decryptCollection(encryptedCollection, password = null) {
    try {
      console.log(
        "[CreateCollectionManager] Decrypting collection:",
        encryptedCollection.id,
      );

      // Get collection key from storage
      let collectionKey = this.storageService.getCollectionKey(
        encryptedCollection.id,
      );

      // Get password if needed
      const userPassword = password || (await this.getUserPassword());

      // Use CollectionCryptoService to decrypt the collection
      const decryptedCollection =
        await this.cryptoService.decryptCollectionFromAPI(
          encryptedCollection,
          collectionKey,
          userPassword,
        );

      // If we successfully decrypted and didn't have the key cached, cache it now
      if (decryptedCollection._isDecrypted && !collectionKey && userPassword) {
        try {
          const newCollectionKey =
            await this.cryptoService.decryptCollectionKeyWithPassword(
              encryptedCollection.encrypted_collection_key,
              userPassword,
            );
          this.storageService.storeCollectionKey(
            encryptedCollection.id,
            newCollectionKey,
          );
        } catch (keyError) {
          console.warn(
            "[CreateCollectionManager] Could not cache collection key:",
            keyError,
          );
        }
      }

      return decryptedCollection;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to decrypt collection:",
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
        "[CreateCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Collection Management ===

  // Get all created collections
  getCreatedCollections() {
    return this.storageService.getCreatedCollections();
  }

  // Get collection by ID
  getCollectionById(collectionId) {
    return this.storageService.getCollectionById(collectionId);
  }

  // Search collections
  searchCollections(searchTerm) {
    const collections = this.getCreatedCollections();
    return this.storageService.searchCollections(searchTerm, collections);
  }

  // Remove collection
  async removeCollection(collectionId) {
    try {
      console.log(
        "[CreateCollectionManager] Removing collection:",
        collectionId,
      );

      const removed = this.storageService.removeCollection(collectionId);

      if (removed) {
        // ADDED: Invalidate caches when collection is removed
        this.invalidateRelatedCaches({});

        this.notifyCollectionCreationListeners("collection_removed", {
          collectionId,
        });
      }

      return removed;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to remove collection:",
        error,
      );
      throw error;
    }
  }

  // Clear all collections
  async clearAllCollections() {
    try {
      console.log("[CreateCollectionManager] Clearing all collections");

      this.storageService.clearAllCollections();

      // ADDED: Invalidate caches when all collections are cleared
      this.invalidateRelatedCaches({});

      this.notifyCollectionCreationListeners("all_collections_cleared", {});

      console.log("[CreateCollectionManager] All collections cleared");
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to clear collections:",
        error,
      );
      throw error;
    }
  }

  // === Collection Key Management ===

  // Generate a new collection key (delegates to crypto service)
  generateCollectionKey() {
    return this.cryptoService.generateCollectionKey();
  }

  // Encrypt collection key for sharing (delegates to crypto service)
  async encryptCollectionKeyForRecipient(collectionKey, recipientPublicKey) {
    return this.cryptoService.encryptCollectionKeyForRecipient(
      collectionKey,
      recipientPublicKey,
    );
  }

  // Decrypt shared collection key (delegates to crypto service)
  async decryptSharedCollectionKey(
    encryptedKey,
    userPrivateKey,
    userPublicKey,
  ) {
    return this.cryptoService.decryptSharedCollectionKey(
      encryptedKey,
      userPrivateKey,
      userPublicKey,
    );
  }

  // === Event Management ===

  // Add collection creation listener
  addCollectionCreationListener(callback) {
    if (typeof callback === "function") {
      this.collectionCreationListeners.add(callback);
      console.log(
        "[CreateCollectionManager] Collection creation listener added. Total listeners:",
        this.collectionCreationListeners.size,
      );
    }
  }

  // Remove collection creation listener
  removeCollectionCreationListener(callback) {
    this.collectionCreationListeners.delete(callback);
    console.log(
      "[CreateCollectionManager] Collection creation listener removed. Total listeners:",
      this.collectionCreationListeners.size,
    );
  }

  // Notify collection creation listeners
  notifyCollectionCreationListeners(eventType, eventData) {
    console.log(
      `[CreateCollectionManager] Notifying ${this.collectionCreationListeners.size} listeners of ${eventType}`,
    );

    this.collectionCreationListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[CreateCollectionManager] Error in collection creation listener:",
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

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canCreateCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      crypto: cryptoStatus,
      listenerCount: this.collectionCreationListeners.size,
      hasPasswordService: !!this.getUserPassword,
      hasListCollectionManager: !!this.listCollectionManager,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "CreateCollectionManager",
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

export default CreateCollectionManager;
